package client

import (
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"

	backingserviceinstanceapi "github.com/openshift/origin/pkg/backingserviceinstance/api"
)

// BackingServiceInstancesInterface has methods to work with BackingServiceInstance resources in a namespace
type BackingServiceInstancesInterface interface {
	BackingServiceInstances() BackingServiceInstanceInterface
}

// BackingServiceInstanceInterface exposes methods on project resources.
type BackingServiceInstanceInterface interface {
	Create(p *backingserviceinstanceapi.BackingServiceInstance) (*backingserviceinstanceapi.BackingServiceInstance, error)
	Delete(name string) error
	Update(p *backingserviceinstanceapi.BackingServiceInstance) (*backingserviceinstanceapi.BackingServiceInstance, error)
	Get(name string) (*backingserviceinstanceapi.BackingServiceInstance, error)
	List(label labels.Selector, field fields.Selector) (*backingserviceinstanceapi.BackingServiceInstanceList, error)
	Watch(label labels.Selector, field fields.Selector, resourceVersion string) (watch.Interface, error)
}

type backingserviceinstances struct {
	r *Client
}

// newUsers returns a project
func newBackingServiceInstances(c *Client) *backingserviceinstances {
	return &backingserviceinstances{
		r: c,
	}
}

// Get returns information about a particular project or an error
func (c *backingserviceinstances) Get(name string) (result *backingserviceinstanceapi.BackingServiceInstance, err error) {
	result = &backingserviceinstanceapi.BackingServiceInstance{}
	err = c.r.Get().Resource("backingserviceinstances").Name(name).Do().Into(result)
	return
}

// List returns all backingserviceinstances matching the label selector
func (c *backingserviceinstances) List(label labels.Selector, field fields.Selector) (result *backingserviceinstanceapi.BackingServiceInstanceList, err error) {
	result = &backingserviceinstanceapi.BackingServiceInstanceList{}
	err = c.r.Get().
		Resource("backingserviceinstances").
		LabelsSelectorParam(label).
		FieldsSelectorParam(field).
		Do().
		Into(result)
	return
}

// Create creates a new BackingServiceInstance
func (c *backingserviceinstances) Create(p *backingserviceinstanceapi.BackingServiceInstance) (result *backingserviceinstanceapi.BackingServiceInstance, err error) {
	result = &backingserviceinstanceapi.BackingServiceInstance{}
	err = c.r.Post().Resource("backingserviceinstances").Body(p).Do().Into(result)
	return
}

// Update updates the project on server
func (c *backingserviceinstances) Update(p *backingserviceinstanceapi.BackingServiceInstance) (result *backingserviceinstanceapi.BackingServiceInstance, err error) {
	result = &backingserviceinstanceapi.BackingServiceInstance{}
	err = c.r.Put().Resource("backingserviceinstances").Name(p.Name).Body(p).Do().Into(result)
	return
}

// Delete removes the project on server
func (c *backingserviceinstances) Delete(name string) (err error) {
	err = c.r.Delete().Resource("backingserviceinstances").Name(name).Do().Error()
	return
}

// Watch returns a watch.Interface that watches the requested backingserviceinstances
func (c *backingserviceinstances) Watch(label labels.Selector, field fields.Selector, resourceVersion string) (watch.Interface, error) {
	return c.r.Get().
		Prefix("watch").
		Resource("backingserviceinstances").
		Param("resourceVersion", resourceVersion).
		LabelsSelectorParam(label).
		FieldsSelectorParam(field).
		Watch()
}