package client

import (
	servicebrokerapi "github.com/openshift/origin/pkg/servicebroker/api"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
)

// ServiceBrokersInterface has methods to work with ServiceBroker resources in a namespace
type ServiceBrokersInterface interface {
	ServiceBrokers() ServiceBrokerInterface
}

// ServiceBrokerInterface exposes methods on project resources.
type ServiceBrokerInterface interface {
	Create(p *servicebrokerapi.ServiceBroker) (*servicebrokerapi.ServiceBroker, error)
	Delete(name string) error
	Get(name string) (*servicebrokerapi.ServiceBroker, error)
	List(label labels.Selector, field fields.Selector) (*servicebrokerapi.ServiceBrokerList, error)
}

type serviceBrokers struct {
	r *Client
}

// newUsers returns a project
func newServiceBrokers(c *Client) *serviceBrokers {
	return &serviceBrokers{
		r: c,
	}
}

// Get returns information about a particular project or an error
func (c *serviceBrokers) Get(name string) (result *servicebrokerapi.ServiceBroker, err error) {
	result = &servicebrokerapi.ServiceBroker{}
	err = c.r.Get().Resource("serviceBrokers").Name(name).Do().Into(result)
	return
}

// List returns all serviceBrokers matching the label selector
func (c *serviceBrokers) List(label labels.Selector, field fields.Selector) (result *servicebrokerapi.ServiceBrokerList, err error) {
	result = &servicebrokerapi.ServiceBrokerList{}
	err = c.r.Get().
	Resource("serviceBrokers").
	LabelsSelectorParam(label).
	FieldsSelectorParam(field).
	Do().
	Into(result)
	return
}

// Create creates a new ServiceBroker
func (c *serviceBrokers) Create(p *servicebrokerapi.ServiceBroker) (result *servicebrokerapi.ServiceBroker, err error) {
	result = &servicebrokerapi.ServiceBroker{}
	err = c.r.Post().Resource("serviceBrokers").Body(p).Do().Into(result)
	return
}

// Update updates the project on server
func (c *serviceBrokers) Update(p *servicebrokerapi.ServiceBroker) (result *servicebrokerapi.ServiceBroker, err error) {
	result = &servicebrokerapi.ServiceBroker{}
	err = c.r.Put().Resource("serviceBrokers").Name(p.Name).Body(p).Do().Into(result)
	return
}

// Delete removes the project on server
func (c *serviceBrokers) Delete(name string) (err error) {
	err = c.r.Delete().Resource("serviceBrokers").Name(name).Do().Error()
	return
}
