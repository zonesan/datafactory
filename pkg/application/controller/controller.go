package controller

import (
	"errors"
	"fmt"
	applicationapi "github.com/openshift/origin/pkg/application/api"
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
func (c *ApplicationController) Handle(application *applicationapi.Application) (err error) {

	switch application.Status.Phase {

	case applicationapi.ApplicationNew:

		if err := c.UpsertHandle(application); err != nil {
			application.Status.Phase = applicationapi.ApplicationInActive
			c.Client.Applications(application.Namespace).Update(application)

			return err
		}

		application.Status.Phase = applicationapi.ApplicationActive
		c.Client.Applications(application.Namespace).Update(application)

	case applicationapi.ApplicationTerminatingLabel:

		if err := c.UpsertHandle(application); err != nil {
			application.Status.Phase = applicationapi.ApplicationTerminatingErr
			c.Client.Applications(application.Namespace).Update(application)

			return err
		}

		c.Client.Applications(application.Namespace).Delete(application.Name)

	case applicationapi.ApplicationTerminating:

		if err := c.DeleteHandle(application); err != nil {

			application.Status.Phase = applicationapi.ApplicationTerminatingErr
			c.Client.Applications(application.Namespace).Update(application)

			return err
		}

		c.Client.Applications(application.Namespace).Delete(application.Name)
	}

	return nil
}

func (c *ApplicationController) UpsertHandle(app *applicationapi.Application) error {
	errs := []error{}
	applicationSelector := fmt.Sprintf("%s/Application", app.Namespace)
	for i, item := range app.Spec.Items {
		switch item.Kind {
		case "Build":
			build := c.Client.Builds(app.Namespace)
			b, err := build.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			whetherUpdate := false
			switch app.Status.Phase {
			case applicationapi.ApplicationTerminatingLabel:
				if b.Labels != nil {
					delete(b.Labels, applicationSelector)
					whetherUpdate = true
				}

			default:
				if b.Labels == nil {
					b.Labels = make(map[string]string)
				}

				whetherUpdate = optLabelByItemStatus(b.Labels, item.Status, applicationSelector, item.Name)

			}

			if whetherUpdate {
				if _, err := build.Update(b); err != nil {
					app.Spec.Items[i].Status = ""
					errs = append(errs, err)
				}
			}

		case "BuildConfig":
			buildConfig := c.Client.BuildConfigs(app.Namespace)
			bc, err := buildConfig.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			whetherUpdate := false
			switch app.Status.Phase {
			case applicationapi.ApplicationTerminatingLabel:
				if bc.Labels != nil {
					delete(bc.Labels, applicationSelector)
					whetherUpdate = true
				}

			default:
				if bc.Labels == nil {
					bc.Labels = make(map[string]string)
				}

				whetherUpdate = optLabelByItemStatus(bc.Labels, item.Status, applicationSelector, item.Name)

			}

			if whetherUpdate {
				if _, err := buildConfig.Update(bc); err != nil {
					app.Spec.Items[i].Status = ""
					errs = append(errs, err)
				}
			}

		case "DeploymentConfig":
			deploymentConfig := c.Client.DeploymentConfigs(app.Namespace)
			dc, err := deploymentConfig.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			whetherUpdate := false
			switch app.Status.Phase {
			case applicationapi.ApplicationTerminatingLabel:
				if dc.Labels != nil {
					delete(dc.Labels, applicationSelector)
					whetherUpdate = true
				}

			default:
				if dc.Labels == nil {
					dc.Labels = make(map[string]string)
				}

				whetherUpdate = optLabelByItemStatus(dc.Labels, item.Status, applicationSelector, item.Name)
			}

			if whetherUpdate {
				if _, err := deploymentConfig.Update(dc); err != nil {
					app.Spec.Items[i].Status = ""
					errs = append(errs, err)
				}
			}

		case "ImageStream":
			imageStream := c.Client.ImageStreams(app.Namespace)
			is, err := imageStream.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			whetherUpdate := false
			switch app.Status.Phase {
			case applicationapi.ApplicationTerminatingLabel:
				if is.Labels != nil {
					delete(is.Labels, applicationSelector)
					whetherUpdate = true
				}

			default:
				if is.Labels == nil {
					is.Labels = make(map[string]string)
				}

				whetherUpdate = optLabelByItemStatus(is.Labels, item.Status, applicationSelector, item.Name)
			}

			if whetherUpdate {
				if _, err := imageStream.Update(is); err != nil {
					app.Spec.Items[i].Status = ""
					errs = append(errs, err)
				}
			}

		case "ImageStreamTag":
			//if ist, err := c.Client.ImageStreamTags(app.Namespace).Get(item.Name); ist != nil {
			//	return err
			//} else {
			//	ist.Labels[applicationSelector] = app.Name
			//}

		case "ImageStreamImage":
			//if isi, err := c.Client.ImageStreamImages(app.Namespace).Get(item.Name); isi != nil {
			//	return err
			//} else {
			//	isi.Labels[applicationSelector] = app.Name
			//}

		case "Event":
			event := c.KubeClient.Events(app.Namespace)
			e, err := event.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			whetherUpdate := false
			switch app.Status.Phase {
			case applicationapi.ApplicationTerminatingLabel:
				if e.Labels != nil {
					delete(e.Labels, applicationSelector)
					whetherUpdate = true
				}

			default:
				if e.Labels == nil {
					e.Labels = make(map[string]string)
				}

				whetherUpdate = optLabelByItemStatus(e.Labels, item.Status, applicationSelector, item.Name)
			}

			if whetherUpdate {
				if _, err := event.Update(e); err != nil {
					app.Spec.Items[i].Status = ""
					errs = append(errs, err)
				}
			}

		case "Node":
			node := c.KubeClient.Nodes()
			n, err := node.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			whetherUpdate := false
			switch app.Status.Phase {
			case applicationapi.ApplicationTerminatingLabel:
				if n.Labels != nil {
					delete(n.Labels, applicationSelector)
					whetherUpdate = true
				}

			default:
				if n.Labels == nil {
					n.Labels = make(map[string]string)
				}

				whetherUpdate = optLabelByItemStatus(n.Labels, item.Status, applicationSelector, item.Name)
			}

			if whetherUpdate {
				if _, err := node.Update(n); err != nil {
					app.Spec.Items[i].Status = ""
					errs = append(errs, err)
				}
			}

		case "Job":

		case "Pod":
			pod := c.KubeClient.Pods(app.Namespace)
			p, err := pod.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			whetherUpdate := false
			switch app.Status.Phase {
			case applicationapi.ApplicationTerminatingLabel:
				if p.Labels != nil {
					delete(p.Labels, applicationSelector)
					whetherUpdate = true
				}

			default:
				if p.Labels == nil {
					p.Labels = make(map[string]string)
				}

				whetherUpdate = optLabelByItemStatus(p.Labels, item.Status, applicationSelector, item.Name)
			}

			if whetherUpdate {
				if _, err := pod.Update(p); err != nil {
					app.Spec.Items[i].Status = ""
					errs = append(errs, err)
				}
			}

		case "ReplicationController":
			replicationController := c.KubeClient.ReplicationControllers(app.Namespace)
			rc, err := replicationController.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			whetherUpdate := false
			switch app.Status.Phase {
			case applicationapi.ApplicationTerminatingLabel:
				if rc.Labels != nil {
					delete(rc.Labels, applicationSelector)
					whetherUpdate = true
				}

			default:
				if rc.Labels == nil {
					rc.Labels = make(map[string]string)
				}

				whetherUpdate = optLabelByItemStatus(rc.Labels, item.Status, applicationSelector, item.Name)
			}

			if whetherUpdate {
				if _, err := replicationController.Update(rc); err != nil {
					app.Spec.Items[i].Status = ""
					errs = append(errs, err)
				}
			}

		case "Service":
			servce := c.KubeClient.Services(app.Namespace)
			s, err := servce.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			whetherUpdate := false
			switch app.Status.Phase {
			case applicationapi.ApplicationTerminatingLabel:
				if s.Labels != nil {
					delete(s.Labels, applicationSelector)
					whetherUpdate = true
				}

			default:
				if s.Labels == nil {
					s.Labels = make(map[string]string)
				}

				whetherUpdate = optLabelByItemStatus(s.Labels, item.Status, applicationSelector, item.Name)
			}

			if whetherUpdate {
				if _, err := servce.Update(s); err != nil {
					app.Spec.Items[i].Status = ""
					errs = append(errs, err)
				}
			}

		case "PersistentVolume":
			persistentVolume := c.KubeClient.PersistentVolumes()
			pv, err := persistentVolume.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			whetherUpdate := false
			switch app.Status.Phase {
			case applicationapi.ApplicationTerminatingLabel:
				if pv.Labels != nil {
					delete(pv.Labels, applicationSelector)
					whetherUpdate = true
				}

			default:
				if pv.Labels == nil {
					pv.Labels = make(map[string]string)
				}

				whetherUpdate = optLabelByItemStatus(pv.Labels, item.Status, applicationSelector, item.Name)
			}

			if whetherUpdate {
				if _, err := persistentVolume.Update(pv); err != nil {
					app.Spec.Items[i].Status = ""
					errs = append(errs, err)
				}
			}

		case "PersistentVolumeClaim":
			persistentVolumeClaim := c.KubeClient.PersistentVolumeClaims(app.Namespace)
			pvc, err := persistentVolumeClaim.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			whetherUpdate := false
			switch app.Status.Phase {
			case applicationapi.ApplicationTerminatingLabel:
				if pvc.Labels != nil {
					delete(pvc.Labels, applicationSelector)
					whetherUpdate = true
				}

			default:
				if pvc.Labels == nil {
					pvc.Labels = make(map[string]string)
				}

				whetherUpdate = optLabelByItemStatus(pvc.Labels, item.Status, applicationSelector, item.Name)
			}

			if whetherUpdate {
				if _, err := persistentVolumeClaim.Update(pvc); err != nil {
					app.Spec.Items[i].Status = ""
					errs = append(errs, err)
				}
			}

		case "ServiceBroker":
			serviceBroker := c.Client.ServiceBrokers()
			sb, err := serviceBroker.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			whetherUpdate := false
			switch app.Status.Phase {
			case applicationapi.ApplicationTerminatingLabel:
				if sb.Labels != nil {
					delete(sb.Labels, applicationSelector)
					whetherUpdate = true
				}

			default:
				if sb.Labels == nil {
					sb.Labels = make(map[string]string)
				}

				whetherUpdate = optLabelByItemStatus(sb.Labels, item.Status, applicationSelector, item.Name)
			}

			if whetherUpdate {
				if _, err := serviceBroker.Update(sb); err != nil {
					app.Spec.Items[i].Status = ""
					errs = append(errs, err)
				}
			}

		case "BackingService":
			backingService := c.Client.BackingServices()
			bs, err := backingService.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			whetherUpdate := false
			switch app.Status.Phase {
			case applicationapi.ApplicationTerminatingLabel:
				if bs.Labels != nil {
					delete(bs.Labels, applicationSelector)
					whetherUpdate = true
				}

			default:
				if bs.Labels == nil {
					bs.Labels = make(map[string]string)
				}

				whetherUpdate = optLabelByItemStatus(bs.Labels, item.Status, applicationSelector, item.Name)
			}

			if whetherUpdate {
				if _, err := backingService.Update(bs); err != nil {
					app.Spec.Items[i].Status = ""
					errs = append(errs, err)
				}
			}

		case "BackingServiceInstance":
			backingServiceInstance := c.Client.BackingServiceInstances(app.Namespace)
			bsi, err := backingServiceInstance.Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			whetherUpdate := false
			switch app.Status.Phase {
			case applicationapi.ApplicationTerminatingLabel:
				if bsi.Labels != nil {
					delete(bsi.Labels, applicationSelector)
					whetherUpdate = true
				}

			default:
				if bsi.Labels == nil {
					bsi.Labels = make(map[string]string)
				}

				whetherUpdate = optLabelByItemStatus(bsi.Labels, item.Status, applicationSelector, item.Name)
			}

			if whetherUpdate {
				if _, err := backingServiceInstance.Update(bsi); err != nil {
					app.Spec.Items[i].Status = ""
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
	case applicationapi.ApplicationItemStatusAdd:
		label[labelKey] = labelValue
		return true
	case applicationapi.ApplicationItemStatusDelete:
		delete(label, labelKey)
		return true
	}

	return false
}

func (c *ApplicationController) DeleteHandle(app *applicationapi.Application) error {
	return deleteAllContent(c.Client, c.KubeClient, app)
}

func deleteAllContent(c osclient.Interface, kc kclient.Interface, app *applicationapi.Application) error {
	errs := []error{}
	for _, item := range app.Spec.Items {
		labelString := fmt.Sprintf("%s/Application=%s", app.Namespace, item.Name)

		switch item.Kind {
		case "Build":
			itemResource, err := c.Builds(app.Namespace).Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			if containsOtherApplicationLabel(itemResource.Labels, labelString) {
				delete(itemResource.Labels, labelString)
				if _, err := c.Builds(app.Namespace).Update(itemResource); err != nil {
					errs = append(errs, err)
				}
			} else {
				err := c.Builds(app.Namespace).Delete(item.Name)
				if err != nil && !kerrors.IsNotFound(err) {
					errs = append(errs, err)
				}
			}

		case "BuildConfig":
			itemResource, err := c.BuildConfigs(app.Namespace).Get(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}

			if containsOtherApplicationLabel(itemResource.Labels, labelString) {
				delete(itemResource.Labels, labelString)
				if _, err := c.BuildConfigs(app.Namespace).Update(itemResource); err != nil {
					errs = append(errs, err)
				}
			} else {
				err := c.BuildConfigs(app.Namespace).Delete(item.Name)
				if err != nil && !kerrors.IsNotFound(err) {
					errs = append(errs, err)
				}
			}

		case "DeploymentConfig":
			err := c.DeploymentConfigs(app.Namespace).Delete(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "ImageStream":
			err := c.ImageStreams(app.Namespace).Delete(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "ImageStreamTag":

		case "ImageStreamImage":

		case "Event":
			err := kc.Events(app.Namespace).Delete(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "Node":
			err := kc.Nodes().Delete(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "Job":

		case "Pod":
			// todo make sure deleteOption
			err := kc.Pods(app.Namespace).Delete(item.Name, nil)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "ReplicationController":
			err := kc.ReplicationControllers(app.Namespace).Delete(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "Service":
			err := kc.Services(app.Namespace).Delete(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "PersistentVolume":
			err := kc.PersistentVolumes().Delete(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "PersistentVolumeClaim":
			err := kc.PersistentVolumeClaims(app.Namespace).Delete(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "ServiceBroker":
			err := c.ServiceBrokers().Delete(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "BackingService":
			err := c.BackingServices().Delete(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "BackingServiceInstance":
			err := c.BackingServiceInstances(app.Namespace).Delete(item.Name)
			if err != nil && !kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		default:
			err := errors.New("unknown resource " + item.Kind + "=" + item.Name)
			errs = append(errs, err)
		}
	}

	return errutil.NewAggregate(errs)
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
		for k, v := range label {
			if strings.HasSuffix(k, "/Application") {
				arr = append(arr, fmt.Sprintf("%s/Application=%s", k, v))
			}
		}
	}

	return arr
}
