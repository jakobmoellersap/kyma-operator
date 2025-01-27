package remote

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kyma-project/lifecycle-manager/api/v1beta1"
	"github.com/kyma-project/lifecycle-manager/pkg/adapter"
	corev1 "k8s.io/api/core/v1"
	v1extensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ClientFunc func() *rest.Config

var (
	LocalClient             ClientFunc //nolint:gochecknoglobals
	ErrNoLocalClientDefined = errors.New("no local client defined")
)

type KymaSynchronizationContext struct {
	ControlPlaneClient Client
	RuntimeClient      Client
}

func InitializeKymaSynchronizationContext(
	ctx context.Context, kcp Client, cache *ClientCache, kyma *v1beta1.Kyma,
) (*KymaSynchronizationContext, error) {
	skr, err := NewClientLookup(kcp, cache, kyma.Spec.Sync.Strategy).Lookup(ctx, client.ObjectKeyFromObject(kyma))
	if err != nil {
		return nil, err
	}

	sync := &KymaSynchronizationContext{
		ControlPlaneClient: kcp,
		RuntimeClient:      skr,
	}

	if err := sync.ensureRemoteNamespaceExists(ctx, kyma); err != nil {
		return nil, err
	}

	return sync, nil
}

func (c *KymaSynchronizationContext) GetRemotelySyncedKyma(
	ctx context.Context, controlPlaneKyma *v1beta1.Kyma,
) (*v1beta1.Kyma, error) {
	remoteKyma := &v1beta1.Kyma{}
	if err := c.RuntimeClient.Get(ctx, GetRemoteObjectKey(controlPlaneKyma), remoteKyma); err != nil {
		return nil, err
	}

	return remoteKyma, nil
}

func RemoveFinalizerFromRemoteKyma(
	ctx context.Context, kyma *v1beta1.Kyma,
) error {
	syncContext := SyncContextFromContext(ctx)

	remoteKyma, err := syncContext.GetRemotelySyncedKyma(ctx, kyma)
	if err != nil {
		return err
	}

	controllerutil.RemoveFinalizer(remoteKyma, v1beta1.Finalizer)

	return syncContext.RuntimeClient.Update(ctx, remoteKyma)
}

func DeleteRemotelySyncedKyma(
	ctx context.Context, kyma *v1beta1.Kyma,
) error {
	syncContext := SyncContextFromContext(ctx)
	remoteKyma, err := syncContext.GetRemotelySyncedKyma(ctx, kyma)
	if err != nil {
		return err
	}

	return syncContext.RuntimeClient.Delete(ctx, remoteKyma)
}

// ensureRemoteNamespaceExists tries to ensure existence of a namespace for synchronization based on
// 1. name of namespace if controlPlaneKyma.spec.sync.namespace is set
// 2. name of controlPlaneKyma.namespace
// in this order.
func (c *KymaSynchronizationContext) ensureRemoteNamespaceExists(ctx context.Context, kyma *v1beta1.Kyma) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        kyma.GetNamespace(),
			Labels:      map[string]string{v1beta1.ManagedBy: v1beta1.OperatorName},
			Annotations: map[string]string{v1beta1.LastSync: time.Now().Format(time.RFC3339)},
		},
		// setting explicit type meta is required for SSA on Namespaces
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
	}
	if kyma.Spec.Sync.Namespace != "" {
		namespace.SetName(kyma.Spec.Sync.Namespace)
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(namespace); err != nil {
		return err
	}

	patch := client.RawPatch(types.ApplyPatchType, buf.Bytes())
	force := true
	fieldManager := "kyma-sync-context"

	if err := c.RuntimeClient.Patch(
		ctx, namespace, patch, &client.PatchOptions{Force: &force, FieldManager: fieldManager},
	); err != nil {
		return fmt.Errorf("failed to ensure remote namespace exists: %w", err)
	}

	return nil
}

func (c *KymaSynchronizationContext) CreateOrUpdateCRD(ctx context.Context, plural string) error {
	crd := &v1extensions.CustomResourceDefinition{}
	crdFromRuntime := &v1extensions.CustomResourceDefinition{}
	var err error
	err = c.ControlPlaneClient.Get(ctx, client.ObjectKey{
		// this object name is derived from the plural and is the default kustomize value for crd namings, if the CRD
		// name changes, this also has to be adjusted here. We can think of making this configurable later
		Name: fmt.Sprintf("%s.%s", plural, v1beta1.GroupVersion.Group),
	}, crd)

	if err != nil {
		return err
	}

	err = c.RuntimeClient.Get(ctx, client.ObjectKey{
		Name: fmt.Sprintf("%s.%s", plural, v1beta1.GroupVersion.Group),
	}, crdFromRuntime)

	if k8serrors.IsNotFound(err) || !ContainsLatestVersion(crdFromRuntime, v1beta1.GroupVersion.Version) {
		return PatchCRD(ctx, c.RuntimeClient, crd)
	}

	if err != nil {
		return err
	}

	return nil
}

func (c *KymaSynchronizationContext) CreateOrFetchRemoteKyma(
	ctx context.Context, kyma *v1beta1.Kyma,
) (*v1beta1.Kyma, error) {
	recorder := adapter.RecorderFromContext(ctx)
	remoteKyma := &v1beta1.Kyma{}

	remoteKyma.Name = kyma.Name
	remoteKyma.Namespace = kyma.Namespace
	if kyma.Spec.Sync.Namespace != "" {
		remoteKyma.Namespace = kyma.Spec.Sync.Namespace
	}

	err := c.RuntimeClient.Get(ctx, client.ObjectKeyFromObject(remoteKyma), remoteKyma)

	if meta.IsNoMatchError(err) {
		recorder.Event(kyma, "Normal", err.Error(), "CRDs are missing in SKR and will be installed")

		if err := c.CreateOrUpdateCRD(ctx, v1beta1.KymaKind.Plural()); err != nil {
			return nil, err
		}

		recorder.Event(kyma, "Normal", "CRDInstallation", "CRDs were installed to SKR")
		// the NoMatch error we previously encountered is now fixed through the CRD installation
		err = nil
	}

	if k8serrors.IsNotFound(err) {
		kyma.Spec.DeepCopyInto(&remoteKyma.Spec)

		if kyma.Spec.Sync.NoModuleCopy {
			remoteKyma.Spec.Modules = []v1beta1.Module{}
		}

		err = c.RuntimeClient.Create(ctx, remoteKyma)
		if err != nil {
			recorder.Event(kyma, "Normal", "RemoteInstallation", "Kyma was installed to SKR")

			return nil, err
		}
	} else if err != nil {
		recorder.Event(kyma, "Warning", err.Error(), "Client could not fetch remote Kyma")

		return nil, err
	}

	return remoteKyma, err
}

func (c *KymaSynchronizationContext) SynchronizeRemoteKyma(ctx context.Context,
	controlPlaneKyma, remoteKyma *v1beta1.Kyma,
) error {
	recorder := adapter.RecorderFromContext(ctx)

	remoteKyma.Status = controlPlaneKyma.Status

	if err := c.RuntimeClient.Status().Update(ctx, remoteKyma); err != nil {
		recorder.Event(controlPlaneKyma, "Warning", err.Error(), "could not update runtime kyma status")
		return err
	}

	if !remoteKyma.GetDeletionTimestamp().IsZero() {
		return nil
	}

	c.InsertWatcherLabels(controlPlaneKyma, remoteKyma)

	if err := c.RuntimeClient.Update(ctx, remoteKyma.SetLastSync()); err != nil {
		recorder.Event(controlPlaneKyma, "Warning", err.Error(), "could not update runtime kyma last sync annotation")
		return err
	}

	return nil
}

// ReplaceWithVirtualKyma creates a virtual kyma instance from a control plane Kyma and N Remote Kymas,
// merging the module specification in the process.
func (c *KymaSynchronizationContext) ReplaceWithVirtualKyma(kyma *v1beta1.Kyma,
	remotes ...*v1beta1.Kyma,
) {
	totalModuleAmount := len(kyma.Spec.Modules)
	for _, remote := range remotes {
		totalModuleAmount += len(remote.Spec.Modules)
	}
	modules := make(map[string]v1beta1.Module, totalModuleAmount)

	for _, remote := range remotes {
		for _, m := range remote.Spec.Modules {
			modules[m.Name] = m
		}
	}
	for _, m := range kyma.Spec.Modules {
		modules[m.Name] = m
	}

	kyma.Spec.Modules = []v1beta1.Module{}
	for _, m := range modules {
		kyma.Spec.Modules = append(kyma.Spec.Modules, m)
	}
}

func GetRemoteObjectKey(kyma *v1beta1.Kyma) client.ObjectKey {
	name := kyma.Name
	namespace := kyma.Namespace
	if kyma.Spec.Sync.Namespace != "" {
		namespace = kyma.Spec.Sync.Namespace
	}
	return client.ObjectKey{Namespace: namespace, Name: name}
}

// InsertWatcherLabels inserts labels into the given KymaCR, which are needed to ensure
// a working e2e-flow for the runtime-watcher.
func (c *KymaSynchronizationContext) InsertWatcherLabels(controlPlaneKyma, remoteKyma *v1beta1.Kyma) {
	if remoteKyma.Labels == nil {
		remoteKyma.Labels = make(map[string]string)
	}

	remoteKyma.Labels[v1beta1.OwnedByLabel] = fmt.Sprintf(
		v1beta1.OwnedByFormat,
		controlPlaneKyma.Namespace, controlPlaneKyma.Name)
	remoteKyma.Labels[v1beta1.WatchedByLabel] = v1beta1.OperatorName
}
