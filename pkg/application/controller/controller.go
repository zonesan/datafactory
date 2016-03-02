package controller

import (
	"errors"
	"fmt"
	api "github.com/openshift/origin/pkg/application/api"
	osclient "github.com/openshift/origin/pkg/client"
	kerrors "k8s.io/kubernetes/pkg/api/errors"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
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

	if err := c.itemsHandler(application); err != nil {
		application.Status.Phase = api.ApplicationInActive
		return err
	}

	switch application.Status.Phase {

	case api.ApplicationNew:

		application.Status.Phase = api.ApplicationActive
		fallthrough

	case api.ApplicationInActive:

		fallthrough

	case api.ApplicationActive:

		c.Client.Applications(application.Namespace).Update(application)

	case api.ApplicationTerminatingLabel:

		fallthrough

	case api.ApplicationTerminating:

		c.Client.Applications(application.Namespace).Delete(application.Name)

	}

	return nil
}

func (c *ApplicationController) itemsHandler(app *api.Application) error {
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

			toDeleteResource, toUpdateResource := false, false
			toUpdateItemOK := false
			switch app.Status.Phase {

			case api.ApplicationNew:
				if resource.Labels == nil {
					resource.Labels = make(map[string]string)
				}

				toUpdateResource = optLabelByItemStatus(resource.Labels, item.Status, labelSelectorStr, app.Name)
				toUpdateItemOK = true

			case api.ApplicationTerminatingLabel:
				if resource.Labels != nil {
					delete(resource.Labels, labelSelectorStr)
					toUpdateResource, toUpdateItemOK = true, true

				}

			case api.ApplicationTerminating:
				if containsOtherApplicationLabel(resource.Labels, labelSelectorStr) {
					delete(resource.Labels, labelSelectorStr)
					toUpdateResource = true
				} else {
					toDeleteResource = true
				}

			case api.ApplicationInActive:
				fallthrough
			case api.ApplicationActive:
				toUpdateResource = optLabelByItemStatus(resource.Labels, item.Status, labelSelectorStr, item.Name)
				toUpdateItemOK = true

			}

			if toUpdateResource {
				_, err := client.Update(resource)
				if err != nil {
					app.Spec.Items[i].Status = api.ApplicationItemStatusErr
					errs = append(errs, err)
				}

				if err == nil && toUpdateItemOK {
					switch app.Spec.Items[i].Status {
					case api.ApplicationItemStatusAdd:
						app.Spec.Items[i].Status = api.ApplicationItemStatusOk
					case api.ApplicationItemStatusDelete:
						app.Spec.Items = append(app.Spec.Items[:i], app.Spec.Items[i+1:]...)
					}
				}
			}

			if toDeleteResource {
				if err := client.Delete(item.Name); err != nil {
					errs = append(errs, err)
				}
			}

		case "BackingService":

			client := c.Client.BackingServices()
			resource, err := client.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			toDeleteResource, toUpdateResource := false, false
			toUpdateItemOK := false
			switch app.Status.Phase {

			case api.ApplicationNew:
				if resource.Labels == nil {
					resource.Labels = make(map[string]string)
				}

				toUpdateResource = optLabelByItemStatus(resource.Labels, item.Status, labelSelectorStr, app.Name)
				toUpdateItemOK = true

			case api.ApplicationTerminatingLabel:
				if resource.Labels != nil {
					delete(resource.Labels, labelSelectorStr)
					toUpdateResource, toUpdateItemOK = true, true

				}

			case api.ApplicationTerminating:
				if containsOtherApplicationLabel(resource.Labels, labelSelectorStr) {
					delete(resource.Labels, labelSelectorStr)
					toUpdateResource = true
				} else {
					toDeleteResource = true
				}

			case api.ApplicationInActive:
				fallthrough
			case api.ApplicationActive:
				toUpdateResource = optLabelByItemStatus(resource.Labels, item.Status, labelSelectorStr, item.Name)
				toUpdateItemOK = true

			}

			if toUpdateResource {
				_, err := client.Update(resource)
				if err != nil {
					app.Spec.Items[i].Status = api.ApplicationItemStatusErr
					errs = append(errs, err)
				}

				if err == nil && toUpdateItemOK {
					switch app.Spec.Items[i].Status {
					case api.ApplicationItemStatusAdd:
						app.Spec.Items[i].Status = api.ApplicationItemStatusOk
					case api.ApplicationItemStatusDelete:
						app.Spec.Items = append(app.Spec.Items[:i], app.Spec.Items[i+1:]...)
					}
				}
			}

			if toDeleteResource {
				if err := client.Delete(item.Name); err != nil {
					errs = append(errs, err)
				}
			}

		case "BackingServiceInstance":

			client := c.Client.BackingServiceInstances(app.Namespace)
			resource, err := client.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			toDeleteResource, toUpdateResource := false, false
			toUpdateItemOK := false
			switch app.Status.Phase {

			case api.ApplicationNew:
				if resource.Labels == nil {
					resource.Labels = make(map[string]string)
				}

				toUpdateResource = optLabelByItemStatus(resource.Labels, item.Status, labelSelectorStr, app.Name)
				toUpdateItemOK = true

			case api.ApplicationTerminatingLabel:
				if resource.Labels != nil {
					delete(resource.Labels, labelSelectorStr)
					toUpdateResource, toUpdateItemOK = true, true

				}

			case api.ApplicationTerminating:
				if containsOtherApplicationLabel(resource.Labels, labelSelectorStr) {
					delete(resource.Labels, labelSelectorStr)
					toUpdateResource = true
				} else {
					toDeleteResource = true
				}

			case api.ApplicationInActive:
				fallthrough
			case api.ApplicationActive:
				toUpdateResource = optLabelByItemStatus(resource.Labels, item.Status, labelSelectorStr, item.Name)
				toUpdateItemOK = true

			}

			if toUpdateResource {
				_, err := client.Update(resource)
				if err != nil {
					app.Spec.Items[i].Status = api.ApplicationItemStatusErr
					errs = append(errs, err)
				}

				if err == nil && toUpdateItemOK {
					switch app.Spec.Items[i].Status {
					case api.ApplicationItemStatusAdd:
						app.Spec.Items[i].Status = api.ApplicationItemStatusOk
					case api.ApplicationItemStatusDelete:
						app.Spec.Items = append(app.Spec.Items[:i], app.Spec.Items[i+1:]...)
					}
				}
			}

			if toDeleteResource {
				if err := client.Delete(item.Name); err != nil {
					errs = append(errs, err)
				}
			}

		case "Build":

			client := c.Client.Builds(app.Namespace)
			resource, err := client.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			toDeleteResource, toUpdateResource := false, false
			toUpdateItemOK := false
			switch app.Status.Phase {

			case api.ApplicationNew:
				if resource.Labels == nil {
					resource.Labels = make(map[string]string)
				}

				toUpdateResource = optLabelByItemStatus(resource.Labels, item.Status, labelSelectorStr, app.Name)
				toUpdateItemOK = true

			case api.ApplicationTerminatingLabel:
				if resource.Labels != nil {
					delete(resource.Labels, labelSelectorStr)
					toUpdateResource, toUpdateItemOK = true, true

				}

			case api.ApplicationTerminating:
				if containsOtherApplicationLabel(resource.Labels, labelSelectorStr) {
					delete(resource.Labels, labelSelectorStr)
					toUpdateResource = true
				} else {
					toDeleteResource = true
				}

			case api.ApplicationInActive:
				fallthrough
			case api.ApplicationActive:
				toUpdateResource = optLabelByItemStatus(resource.Labels, item.Status, labelSelectorStr, item.Name)
				toUpdateItemOK = true

			}

			if toUpdateResource {
				_, err := client.Update(resource)
				if err != nil {
					app.Spec.Items[i].Status = api.ApplicationItemStatusErr
					errs = append(errs, err)
				}

				if err == nil && toUpdateItemOK {
					switch app.Spec.Items[i].Status {
					case api.ApplicationItemStatusAdd:
						app.Spec.Items[i].Status = api.ApplicationItemStatusOk
					case api.ApplicationItemStatusDelete:
						app.Spec.Items = append(app.Spec.Items[:i], app.Spec.Items[i+1:]...)
					}
				}
			}

			if toDeleteResource {
				if err := client.Delete(item.Name); err != nil {
					errs = append(errs, err)
				}
			}
		case "BuildConfig":

			client := c.Client.BuildConfigs(app.Namespace)
			resource, err := client.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			toDeleteResource, toUpdateResource := false, false
			toUpdateItemOK := false
			switch app.Status.Phase {

			case api.ApplicationNew:
				if resource.Labels == nil {
					resource.Labels = make(map[string]string)
				}

				toUpdateResource = optLabelByItemStatus(resource.Labels, item.Status, labelSelectorStr, app.Name)
				toUpdateItemOK = true

			case api.ApplicationTerminatingLabel:
				if resource.Labels != nil {
					delete(resource.Labels, labelSelectorStr)
					toUpdateResource, toUpdateItemOK = true, true

				}

			case api.ApplicationTerminating:
				if containsOtherApplicationLabel(resource.Labels, labelSelectorStr) {
					delete(resource.Labels, labelSelectorStr)
					toUpdateResource = true
				} else {
					toDeleteResource = true
				}

			case api.ApplicationInActive:
				fallthrough
			case api.ApplicationActive:
				toUpdateResource = optLabelByItemStatus(resource.Labels, item.Status, labelSelectorStr, item.Name)
				toUpdateItemOK = true

			}

			if toUpdateResource {
				_, err := client.Update(resource)
				if err != nil {
					app.Spec.Items[i].Status = api.ApplicationItemStatusErr
					errs = append(errs, err)
				}

				if err == nil && toUpdateItemOK {
					switch app.Spec.Items[i].Status {
					case api.ApplicationItemStatusAdd:
						app.Spec.Items[i].Status = api.ApplicationItemStatusOk
					case api.ApplicationItemStatusDelete:
						app.Spec.Items = append(app.Spec.Items[:i], app.Spec.Items[i+1:]...)
					}
				}
			}

			if toDeleteResource {
				if err := client.Delete(item.Name); err != nil {
					errs = append(errs, err)
				}
			}
		case "DeploymentConfig":

			client := c.Client.DeploymentConfigs(app.Namespace)
			resource, err := client.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			toDeleteResource, toUpdateResource := false, false
			toUpdateItemOK := false
			switch app.Status.Phase {

			case api.ApplicationNew:
				if resource.Labels == nil {
					resource.Labels = make(map[string]string)
				}

				toUpdateResource = optLabelByItemStatus(resource.Labels, item.Status, labelSelectorStr, app.Name)
				toUpdateItemOK = true

			case api.ApplicationTerminatingLabel:
				if resource.Labels != nil {
					delete(resource.Labels, labelSelectorStr)
					toUpdateResource, toUpdateItemOK = true, true

				}

			case api.ApplicationTerminating:
				if containsOtherApplicationLabel(resource.Labels, labelSelectorStr) {
					delete(resource.Labels, labelSelectorStr)
					toUpdateResource = true
				} else {
					toDeleteResource = true
				}

			case api.ApplicationInActive:
				fallthrough
			case api.ApplicationActive:
				toUpdateResource = optLabelByItemStatus(resource.Labels, item.Status, labelSelectorStr, item.Name)
				toUpdateItemOK = true

			}

			if toUpdateResource {
				_, err := client.Update(resource)
				if err != nil {
					app.Spec.Items[i].Status = api.ApplicationItemStatusErr
					errs = append(errs, err)
				}

				if err == nil && toUpdateItemOK {
					switch app.Spec.Items[i].Status {
					case api.ApplicationItemStatusAdd:
						app.Spec.Items[i].Status = api.ApplicationItemStatusOk
					case api.ApplicationItemStatusDelete:
						app.Spec.Items = append(app.Spec.Items[:i], app.Spec.Items[i+1:]...)
					}
				}
			}

			if toDeleteResource {
				if err := client.Delete(item.Name); err != nil {
					errs = append(errs, err)
				}
			}

		case "ReplicationController":

			client := c.KubeClient.ReplicationControllers(app.Namespace)
			resource, err := client.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			toDeleteResource, toUpdateResource := false, false
			toUpdateItemOK := false
			switch app.Status.Phase {

			case api.ApplicationNew:
				if resource.Labels == nil {
					resource.Labels = make(map[string]string)
				}

				toUpdateResource = optLabelByItemStatus(resource.Labels, item.Status, labelSelectorStr, app.Name)
				toUpdateItemOK = true

			case api.ApplicationTerminatingLabel:
				if resource.Labels != nil {
					delete(resource.Labels, labelSelectorStr)
					toUpdateResource, toUpdateItemOK = true, true

				}

			case api.ApplicationTerminating:
				if containsOtherApplicationLabel(resource.Labels, labelSelectorStr) {
					delete(resource.Labels, labelSelectorStr)
					toUpdateResource = true
				} else {
					toDeleteResource = true
				}

			case api.ApplicationInActive:
				fallthrough
			case api.ApplicationActive:
				toUpdateResource = optLabelByItemStatus(resource.Labels, item.Status, labelSelectorStr, item.Name)
				toUpdateItemOK = true

			}

			if toUpdateResource {
				_, err := client.Update(resource)
				if err != nil {
					app.Spec.Items[i].Status = api.ApplicationItemStatusErr
					errs = append(errs, err)
				}

				if err == nil && toUpdateItemOK {
					switch app.Spec.Items[i].Status {
					case api.ApplicationItemStatusAdd:
						app.Spec.Items[i].Status = api.ApplicationItemStatusOk
					case api.ApplicationItemStatusDelete:
						app.Spec.Items = append(app.Spec.Items[:i], app.Spec.Items[i+1:]...)
					}
				}
			}

			if toDeleteResource {
				if err := client.Delete(item.Name); err != nil {
					errs = append(errs, err)
				}
			}

		case "Node":

			client := c.KubeClient.Nodes()
			resource, err := client.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			toDeleteResource, toUpdateResource := false, false
			toUpdateItemOK := false
			switch app.Status.Phase {

			case api.ApplicationNew:
				if resource.Labels == nil {
					resource.Labels = make(map[string]string)
				}

				toUpdateResource = optLabelByItemStatus(resource.Labels, item.Status, labelSelectorStr, app.Name)
				toUpdateItemOK = true

			case api.ApplicationTerminatingLabel:
				if resource.Labels != nil {
					delete(resource.Labels, labelSelectorStr)
					toUpdateResource, toUpdateItemOK = true, true

				}

			case api.ApplicationTerminating:
				if containsOtherApplicationLabel(resource.Labels, labelSelectorStr) {
					delete(resource.Labels, labelSelectorStr)
					toUpdateResource = true
				} else {
					toDeleteResource = true
				}

			case api.ApplicationInActive:
				fallthrough
			case api.ApplicationActive:
				toUpdateResource = optLabelByItemStatus(resource.Labels, item.Status, labelSelectorStr, item.Name)
				toUpdateItemOK = true

			}

			if toUpdateResource {
				_, err := client.Update(resource)
				if err != nil {
					app.Spec.Items[i].Status = api.ApplicationItemStatusErr
					errs = append(errs, err)
				}

				if err == nil && toUpdateItemOK {
					switch app.Spec.Items[i].Status {
					case api.ApplicationItemStatusAdd:
						app.Spec.Items[i].Status = api.ApplicationItemStatusOk
					case api.ApplicationItemStatusDelete:
						app.Spec.Items = append(app.Spec.Items[:i], app.Spec.Items[i+1:]...)
					}
				}
			}

			if toDeleteResource {
				if err := client.Delete(item.Name); err != nil {
					errs = append(errs, err)
				}
			}

		case "Pod":

			client := c.KubeClient.Pods(app.Namespace)
			resource, err := client.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			toDeleteResource, toUpdateResource := false, false
			toUpdateItemOK := false
			switch app.Status.Phase {

			case api.ApplicationNew:
				if resource.Labels == nil {
					resource.Labels = make(map[string]string)
				}

				toUpdateResource = optLabelByItemStatus(resource.Labels, item.Status, labelSelectorStr, app.Name)
				toUpdateItemOK = true

			case api.ApplicationTerminatingLabel:
				if resource.Labels != nil {
					delete(resource.Labels, labelSelectorStr)
					toUpdateResource, toUpdateItemOK = true, true

				}

			case api.ApplicationTerminating:
				if containsOtherApplicationLabel(resource.Labels, labelSelectorStr) {
					delete(resource.Labels, labelSelectorStr)
					toUpdateResource = true
				} else {
					toDeleteResource = true
				}

			case api.ApplicationInActive:
				fallthrough
			case api.ApplicationActive:
				toUpdateResource = optLabelByItemStatus(resource.Labels, item.Status, labelSelectorStr, item.Name)
				toUpdateItemOK = true

			}

			if toUpdateResource {
				_, err := client.Update(resource)
				if err != nil {
					app.Spec.Items[i].Status = api.ApplicationItemStatusErr
					errs = append(errs, err)
				}

				if err == nil && toUpdateItemOK {
					switch app.Spec.Items[i].Status {
					case api.ApplicationItemStatusAdd:
						app.Spec.Items[i].Status = api.ApplicationItemStatusOk
					case api.ApplicationItemStatusDelete:
						app.Spec.Items = append(app.Spec.Items[:i], app.Spec.Items[i+1:]...)
					}
				}
			}

			if toDeleteResource {
				if err := client.Delete(item.Name, nil); err != nil {
					errs = append(errs, err)
				}
			}

		case "Service":

			client := c.KubeClient.Services(app.Namespace)
			resource, err := client.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			toDeleteResource, toUpdateResource := false, false
			toUpdateItemOK := false
			switch app.Status.Phase {

			case api.ApplicationNew:
				if resource.Labels == nil {
					resource.Labels = make(map[string]string)
				}

				toUpdateResource = optLabelByItemStatus(resource.Labels, item.Status, labelSelectorStr, app.Name)
				toUpdateItemOK = true

			case api.ApplicationTerminatingLabel:
				if resource.Labels != nil {
					delete(resource.Labels, labelSelectorStr)
					toUpdateResource, toUpdateItemOK = true, true

				}

			case api.ApplicationTerminating:
				if containsOtherApplicationLabel(resource.Labels, labelSelectorStr) {
					delete(resource.Labels, labelSelectorStr)
					toUpdateResource = true
				} else {
					toDeleteResource = true
				}

			case api.ApplicationInActive:
				fallthrough
			case api.ApplicationActive:
				toUpdateResource = optLabelByItemStatus(resource.Labels, item.Status, labelSelectorStr, item.Name)
				toUpdateItemOK = true

			}

			if toUpdateResource {
				_, err := client.Update(resource)
				if err != nil {
					app.Spec.Items[i].Status = api.ApplicationItemStatusErr
					errs = append(errs, err)
				}

				if err == nil && toUpdateItemOK {
					switch app.Spec.Items[i].Status {
					case api.ApplicationItemStatusAdd:
						app.Spec.Items[i].Status = api.ApplicationItemStatusOk
					case api.ApplicationItemStatusDelete:
						app.Spec.Items = append(app.Spec.Items[:i], app.Spec.Items[i+1:]...)
					}
				}
			}

			if toDeleteResource {
				if err := client.Delete(item.Name); err != nil {
					errs = append(errs, err)
				}
			}
		default:
			errs = append(errs, errors.New("unknown resource "+item.Kind+"="+item.Name))
		}
	}

	return errutil.NewAggregate(errs)
}

func optLabelByItemStatus(label map[string]string, status, labelKey, labelValue string) bool {

	switch status {
	case api.ApplicationItemStatusAdd:
		label[labelKey] = labelValue
		return true
	case api.ApplicationItemStatusDelete:
		delete(label, labelKey)
		return true
	}

	return false
}

func containsOtherApplicationLabel(label map[string]string, labelStr string) bool {
	list := getApplicationLabels(label)
	if len(list) > 1 {
		for _, v := range list {
			if v == labelStr {
				return true
			}
		}
	}

	return false
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
