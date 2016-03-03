package controller

import (
	"errors"
	"fmt"
	api "github.com/openshift/origin/pkg/application/api"
	osclient "github.com/openshift/origin/pkg/client"
	kerrors "k8s.io/kubernetes/pkg/api/errors"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	errutil "k8s.io/kubernetes/pkg/util/errors"
	"strings"
)

// NamespaceController is responsible for participating in Kubernetes Namespace termination
// Use the NamespaceControllerFactory to create this controller.
type ApplicationController struct {
	// Client is an OpenShift client.
	Client osclient.Interface
	// KubeClient is a Kubernetes client.
	KubeClient kclient.Interface
}

type fatalError string

func (e fatalError) Error() string {
	return "fatal error handling ApplicationController: " + string(e)
}

// Handle processes a namespace and deletes content in origin if its terminating
func (c *ApplicationController) Handle(application *api.Application) (err error) {

	switch application.Status.Phase {
	case api.ApplicationActive:
		return nil
	case api.ApplicationActiveUpdate:
		if err := c.deleteOldLabel(application); err != nil {
			return err
		}

		fallthrough
	default:
		if err := c.handleLabel(application); err != nil {

			return err
		}

		application.Status.Phase = api.ApplicationActive
		c.Client.Applications(application.Namespace).Update(application)

	}

	return nil
}

func (c *ApplicationController) deleteOldLabel(application *api.Application) error {
	errs := []error{}

	if err := deleteLabelServiceBrokers(c.Client, application); err != nil {
		errs = append(errs, err)
	}

	return errutil.NewAggregate(errs)
}

func deleteLabelServiceBrokers(client osclient.Interface, application *api.Application) error {

	selectorStr := fmt.Sprintf("%s.application.%s=%s", application.Namespace, application.Name, application.Name)
	selector, err := labels.Parse(selectorStr)
	if err != nil {
		return err
	}

	items, _ := client.ServiceBrokers().List(selector, fields.Everything())
	errs := []error{}
	for _, oldItem := range items.Items {
		if !hasItem(application.Spec.Items, api.Item{Kind: "ServiceBroker", Name: oldItem.Name}) {
			delete(oldItem.Labels, fmt.Sprintf("%s.application.%s", application.Namespace, application.Name))
			if _, err := client.ServiceBrokers().Update(&oldItem); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return nil
}

func (c *ApplicationController) handleLabel(app *api.Application) error {
	errs := []error{}
	labelSelectorStr := fmt.Sprintf("%s.application.%s", app.Namespace, app.Name)

	for _, item := range app.Spec.Items {
		switch item.Kind {
		case "ServiceBroker":

			client := c.Client.ServiceBrokers()

			resource, err := client.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			switch app.Status.Phase {
			case api.ApplicationActiveUpdate:
				if _, exists := resource.Labels[labelSelectorStr]; exists { //Active正常状态,当有新的更新时,如果这个label不存在,则新建
					continue
				}
				fallthrough
			case api.ApplicationNew:
				if resource.Labels == nil {
					resource.Labels == make(map[string]string)
				}

				resource.Labels[labelSelectorStr] = app.Name
				if _, err := client.Update(resource); err != nil {
					errs = append(errs, err)
				}

			case api.ApplicationTerminating:
				if err := client.Delete(item.Name); err != nil {
					errs = append(errs, err)
				}

			case api.ApplicationTerminatingLabel:
				delete(resource.Labels, app.Name)
				if _, err := client.Update(resource); err != nil {
					errs = append(errs, err)
				}
			}

		default:
			errs = append(errs, errors.New("unknown resource "+item.Kind+"="+item.Name))
		}
	}

	return errutil.NewAggregate(errs)
}

func (c *ApplicationController) deleteAllContentLabel(app *api.Application) error {
	errs := []error{}
	labelSelectorStr := fmt.Sprintf("%s.application.%s", app.Namespace, app.Name)

	for i, item := range app.Spec.Items {
		switch item.Kind {
		case "ServiceBroker":

			client := c.Client.ServiceBrokers()
			resource, err := client.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			if app.Status.Phase == api.ApplicationActive { //Post New
				if _, exists := resource.Labels[labelSelectorStr]; !exists {
					app.Spec.Items[i].Status = "user deleted"
				}
			}

			resource.Labels[labelSelectorStr] = app.Name

			if _, err := client.Update(resource); err != nil {
				errs = append(errs, err)
			}
			return nil

		default:
			errs = append(errs, errors.New("unknown resource "+item.Kind+"="+item.Name))
		}
	}

	return errutil.NewAggregate(errs)
}

func getApplicationLabels(label map[string]string) []string {
	arr := []string{}

	if label != nil {
		for key := range label {
			if strings.Contains(key, ".application.") {
				arr = append(arr, key)
			}
		}
	}

	return arr
}

func hasItem(items api.ItemList, item api.Item) bool {
	for i := range items {
		if items[i].Kind == item.Kind && items[i].Name == item.Name {
			return true
		}
	}

	return false
}
