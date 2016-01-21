package client

import (
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"

	backingserviceapi "github.com/openshift/origin/pkg/backingservice/api"
)

// BackingServicesInterface has methods to work with BackingService resources in a namespace
type BackingServicesInterface interface {
	BackingServices() BackingServiceInterface
}

// BackingServiceInterface exposes methods on project resources.
type BackingServiceInterface interface {
	Create(p *backingserviceapi.BackingService) (*backingserviceapi.BackingService, error)
	Delete(name string) error
	Update(p *backingserviceapi.BackingService) (*backingserviceapi.BackingService, error)
	Get(name string) (*backingserviceapi.BackingService, error)
	List(label labels.Selector, field fields.Selector) (*backingserviceapi.BackingServiceList, error)
	Watch(label labels.Selector, field fields.Selector, resourceVersion string) (watch.Interface, error)
}

type backingservices struct {
	r *Client
}

// newUsers returns a project
func newBackingServices(c *Client) *backingservices {
	return &backingservices{
		r: c,
	}
}

// Get returns information about a particular project or an error
func (c *backingservices) Get(name string) (result *backingserviceapi.BackingService, err error) {
	result = &backingserviceapi.BackingService{}
	err = c.r.Get().Resource("backingservices").Name(name).Do().Into(result)
	return
}

// List returns all backingservices matching the label selector
func (c *backingservices) List(label labels.Selector, field fields.Selector) (result *backingserviceapi.BackingServiceList, err error) {
	result = &backingserviceapi.BackingServiceList{}
	err = c.r.Get().
		Resource("backingservices").
		LabelsSelectorParam(label).
		FieldsSelectorParam(field).
		Do().
		Into(result)
	return
}

// Create creates a new BackingService
func (c *backingservices) Create(p *backingserviceapi.BackingService) (result *backingserviceapi.BackingService, err error) {
	result = &backingserviceapi.BackingService{}
	err = c.r.Post().Resource("backingservices").Body(p).Do().Into(result)
	return
}

// Update updates the project on server
func (c *backingservices) Update(p *backingserviceapi.BackingService) (result *backingserviceapi.BackingService, err error) {
	result = &backingserviceapi.BackingService{}
	err = c.r.Put().Resource("backingservices").Name(p.Name).Body(p).Do().Into(result)
	return
}

// Delete removes the project on server
func (c *backingservices) Delete(name string) (err error) {
	err = c.r.Delete().Resource("backingservices").Name(name).Do().Error()
	return
}

// Watch returns a watch.Interface that watches the requested backingservices
func (c *backingservices) Watch(label labels.Selector, field fields.Selector, resourceVersion string) (watch.Interface, error) {
	return c.r.Get().
		Prefix("watch").
		Resource("backingservices").
		Param("resourceVersion", resourceVersion).
		LabelsSelectorParam(label).
		FieldsSelectorParam(field).
		Watch()
}
