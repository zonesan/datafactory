package controller

import (
	kclient "k8s.io/kubernetes/pkg/client/unversioned"

	"errors"
	"github.com/golang/glog"
	applicationapi "github.com/openshift/origin/pkg/application/api"
	osclient "github.com/openshift/origin/pkg/client"
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

	if err := c.HandleAppItems(application); err != nil {
		application.Status.Phase = applicationapi.ApplicationFailed
		c.Client.Applications(application.Namespace).Update(application)
		return err
	}

	application.Status.Phase = applicationapi.ApplicationActive

	return nil
}

func (c *ApplicationController) HandleAppItems(app *applicationapi.Application) (err error) {
	var globalErr error
	for i, item := range app.Spec.Items {
		if item.Status == applicationapi.ApplicationItemStatusAdd || item.Status == applicationapi.ApplicationItemStatusDelete {
			switch item.Kind {
			case "Build":
				build := c.Client.Builds(app.Namespace)
				b, err := build.Get(item.Name)
				if err != nil {
					glog.Error(err)
					continue
				}

				updateLabelByItem(b.Labels, item)
				_, globalErr = build.Update(b)
				if globalErr == nil {
					app.Spec.Items[i].Status = ""
				}

			case "BuildConfig":
				buildConfig := c.Client.BuildConfigs(app.Namespace)
				bc, err := buildConfig.Get(item.Name)
				if err != nil {
					glog.Error(err)
					continue
				}

				updateLabelByItem(bc.Labels, item)
				_, globalErr = buildConfig.Update(bc)
				if globalErr == nil {
					app.Spec.Items[i].Status = ""
				}

			case "DeploymentConfig":
				deploymentConfig := c.Client.DeploymentConfigs(app.Namespace)
				dc, err := deploymentConfig.Get(item.Name)
				if err != nil {
					glog.Error(err)
					continue
				}

				updateLabelByItem(dc.Labels, item)
				_, globalErr = deploymentConfig.Update(dc)
				if globalErr == nil {
					app.Spec.Items[i].Status = ""
				}

			case "ImageStream":
				imageStream := c.Client.ImageStreams(app.Namespace)
				is, err := imageStream.Get(item.Name)
				if err != nil {
					glog.Error(err)
					continue
				}

				updateLabelByItem(is.Labels, item)
				_, globalErr = imageStream.Update(is)
				if globalErr == nil {
					app.Spec.Items[i].Status = ""
				}

			case "ImageStreamTag":
				//if ist, err := c.Client.ImageStreamTags(app.Namespace).Get(item.Name); ist != nil {
				//	return err
				//} else {
				//	ist.Labels[applicationapi.ApplicationSelector] = app.Name
				//}

			case "ImageStreamImage":
				//if isi, err := c.Client.ImageStreamImages(app.Namespace).Get(item.Name); isi != nil {
				//	return err
				//} else {
				//	isi.Labels[applicationapi.ApplicationSelector] = app.Name
				//}

			case "Event":
				event := c.KubeClient.Events(app.Namespace)
				e, err := event.Get(item.Name)
				if err != nil {
					glog.Error(err)
					continue
				}

				updateLabelByItem(e.Labels, item)
				_, globalErr = event.Update(e)
				if globalErr == nil {
					app.Spec.Items[i].Status = ""
				}

			case "Node":
				node := c.KubeClient.Nodes()
				n, err := node.Get(item.Name)
				if err != nil {
					glog.Error(err)
					continue
				}

				updateLabelByItem(n.Labels, item)
				_, globalErr = node.Update(n)
				if globalErr == nil {
					app.Spec.Items[i].Status = ""
				}

			case "Job":

			case "Pod":
				pod := c.KubeClient.Pods(app.Namespace)
				p, err := pod.Get(item.Name)
				if err != nil {
					glog.Error(err)
					continue
				}

				updateLabelByItem(p.Labels, item)
				_, globalErr = pod.Update(p)
				if globalErr == nil {
					app.Spec.Items[i].Status = ""
				}

			case "ReplicationController":
				replicationController := c.KubeClient.ReplicationControllers(app.Namespace)
				rc, err := replicationController.Get(item.Name)
				if err != nil {
					glog.Error(err)
					continue
				}

				updateLabelByItem(rc.Labels, item)
				_, globalErr = replicationController.Update(rc)
				if globalErr == nil {
					app.Spec.Items[i].Status = ""
				}

			case "Service":
				servce := c.KubeClient.Services(app.Namespace)
				s, err := servce.Get(item.Name)
				if err != nil {
					glog.Error(err)
					continue
				}

				updateLabelByItem(s.Labels, item)
				_, globalErr = servce.Update(s)
				if globalErr == nil {
					app.Spec.Items[i].Status = ""
				}

			case "PersistentVolume":
				persistentVolume := c.KubeClient.PersistentVolumes()
				pv, err := persistentVolume.Get(item.Name)
				if err != nil {
					glog.Error(err)
					continue
				}

				updateLabelByItem(pv.Labels, item)
				_, globalErr = persistentVolume.Update(pv)
				if globalErr == nil {
					app.Spec.Items[i].Status = ""
				}

			case "PersistentVolumeClaim":
				persistentVolumeClaim := c.KubeClient.PersistentVolumeClaims(app.Namespace)
				pvc, err := persistentVolumeClaim.Get(item.Name)
				if err != nil {
					glog.Error(err)
					continue
				}

				updateLabelByItem(pvc.Labels, item)
				_, globalErr = persistentVolumeClaim.Update(pvc)
				if globalErr == nil {
					app.Spec.Items[i].Status = ""
				}

			case "ServiceBroker":
				serviceBroker := c.Client.ServiceBrokers()
				sb, err := serviceBroker.Get(item.Name)
				if err != nil {
					glog.Error(err)
					continue
				}

				updateLabelByItem(sb.Labels, item)
				_, globalErr = serviceBroker.Update(sb)
				if globalErr == nil {
					app.Spec.Items[i].Status = ""
				}

			case "BackingService":
				backingService := c.Client.BackingServices()
				bs, err := backingService.Get(item.Name)
				if err != nil {
					glog.Error(err)
					continue
				}

				updateLabelByItem(bs.Labels, item)
				_, globalErr = backingService.Update(bs)
				if globalErr == nil {
					app.Spec.Items[i].Status = ""
				}

			case "BackingServiceInstance":
				backingServiceInstance := c.Client.BackingServiceInstances(app.Namespace)
				bsi, err := backingServiceInstance.Get(item.Name)
				if err != nil {
					glog.Error(err)
					continue
				}

				updateLabelByItem(bsi.Labels, item)
				_, globalErr = backingServiceInstance.Update(bsi)
				if globalErr == nil {
					app.Spec.Items[i].Status = ""
				}

			default:
				globalErr = errors.New("unknown resource " + item.Kind + "=" + item.Name)
			}
		}
	}



	return globalErr
}

func updateLabelByItem(label map[string]string, item applicationapi.Item) bool {
	switch item.Status {
	case applicationapi.ApplicationItemStatusAdd:
		label[applicationapi.ApplicationSelector] = item.Name
		return true
	case applicationapi.ApplicationItemStatusDelete:
		delete(label, applicationapi.ApplicationSelector)
		return true
	}

	return false
}

