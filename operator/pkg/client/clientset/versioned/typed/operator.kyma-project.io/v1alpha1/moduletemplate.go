/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/kyma-project/lifecycle-manager/operator/api/v1alpha1"
	scheme "github.com/kyma-project/lifecycle-manager/operator/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ModuleTemplatesGetter has a method to return a ModuleTemplateInterface.
// A group's client should implement this interface.
type ModuleTemplatesGetter interface {
	ModuleTemplates(namespace string) ModuleTemplateInterface
}

// ModuleTemplateInterface has methods to work with ModuleTemplate resources.
type ModuleTemplateInterface interface {
	Create(ctx context.Context, moduleTemplate *v1alpha1.ModuleTemplate, opts v1.CreateOptions) (*v1alpha1.ModuleTemplate, error)
	Update(ctx context.Context, moduleTemplate *v1alpha1.ModuleTemplate, opts v1.UpdateOptions) (*v1alpha1.ModuleTemplate, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.ModuleTemplate, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.ModuleTemplateList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ModuleTemplate, err error)
	ModuleTemplateExpansion
}

// moduleTemplates implements ModuleTemplateInterface
type moduleTemplates struct {
	client rest.Interface
	ns     string
}

// newModuleTemplates returns a ModuleTemplates
func newModuleTemplates(c *OperatorV1alpha1Client, namespace string) *moduleTemplates {
	return &moduleTemplates{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the moduleTemplate, and returns the corresponding moduleTemplate object, and an error if there is any.
func (c *moduleTemplates) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ModuleTemplate, err error) {
	result = &v1alpha1.ModuleTemplate{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("moduletemplates").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ModuleTemplates that match those selectors.
func (c *moduleTemplates) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ModuleTemplateList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.ModuleTemplateList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("moduletemplates").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested moduleTemplates.
func (c *moduleTemplates) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("moduletemplates").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a moduleTemplate and creates it.  Returns the server's representation of the moduleTemplate, and an error, if there is any.
func (c *moduleTemplates) Create(ctx context.Context, moduleTemplate *v1alpha1.ModuleTemplate, opts v1.CreateOptions) (result *v1alpha1.ModuleTemplate, err error) {
	result = &v1alpha1.ModuleTemplate{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("moduletemplates").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(moduleTemplate).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a moduleTemplate and updates it. Returns the server's representation of the moduleTemplate, and an error, if there is any.
func (c *moduleTemplates) Update(ctx context.Context, moduleTemplate *v1alpha1.ModuleTemplate, opts v1.UpdateOptions) (result *v1alpha1.ModuleTemplate, err error) {
	result = &v1alpha1.ModuleTemplate{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("moduletemplates").
		Name(moduleTemplate.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(moduleTemplate).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the moduleTemplate and deletes it. Returns an error if one occurs.
func (c *moduleTemplates) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("moduletemplates").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *moduleTemplates) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("moduletemplates").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patches and returns the patched moduleTemplate.
func (c *moduleTemplates) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ModuleTemplate, err error) {
	result = &v1alpha1.ModuleTemplate{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("moduletemplates").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
