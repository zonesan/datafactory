package controller

import (
	kclient "k8s.io/kubernetes/pkg/client/unversioned"

	osclient "github.com/openshift/origin/pkg/client"
	servicebrokerapi "github.com/openshift/origin/pkg/servicebroker/api"

	"github.com/golang/glog"
	backingservice "github.com/openshift/origin/pkg/backingservice/api"
	servicebrokerclient "github.com/openshift/origin/pkg/servicebroker/client"
)

// NamespaceController is responsible for participating in Kubernetes Namespace termination
// Use the NamespaceControllerFactory to create this controller.
type ServiceBrokerController struct {
	// Client is an OpenShift client.
	Client osclient.Interface
	// KubeClient is a Kubernetes client.
	KubeClient kclient.Interface
	//ServiceBrokerClient is a ServiceBroker client
	ServiceBrokerClient servicebrokerclient.Interface
}

type fatalError string

func (e fatalError) Error() string {
	return "fatal error handling ServiceBrokerController: " + string(e)
}

// Handle processes a namespace and deletes content in origin if its terminating
func (c *ServiceBrokerController) Handle(sb *servicebrokerapi.ServiceBroker) (err error) {

	if sb.Spec.Url == "" {
		return nil
	}

	services, err := c.ServiceBrokerClient.Catalog(sb.Spec.Url, sb.Spec.UserName, sb.Spec.Password)
	if err != nil {
		glog.Infof("ServiceBroker %s catalog err %s\n", sb.Name, err.Error())
		//time.Sleep(time.Minute * 5)
		if sb.Status.Phase != servicebrokerapi.ServiceBrokerFailed {
			sb.Status.Phase = servicebrokerapi.ServiceBrokerFailed
			c.Client.ServiceBrokers().Update(sb)
		}

		return nil
	}

	finish := false
	for _, v := range services.Services {
		backingService := &backingservice.BackingService{}
		backingService.Spec = backingservice.BackingServiceSpec(v)
		backingService.Annotations = make(map[string]string)
		backingService.Name = v.Name
		backingService.GenerateName = sb.Name
		backingService.Labels = map[string]string{
			servicebrokerapi.ServiceBrokerLabel: sb.Name,
		}
		glog.Info("create backingservice")
		if _, err := c.Client.BackingServices("openshift").Create(backingService); err == nil {
			glog.Info("create backingservice successfuly!", backingService)
			finish = true
		} else {
			glog.Info(err)
			return err
		}
	}
	if finish == true {
		//TODO
	}

	if sb.Status.Phase != servicebrokerapi.ServiceBrokerActive {
		sb.Status.Phase = servicebrokerapi.ServiceBrokerActive
		c.Client.ServiceBrokers().Update(sb)
	}

	return nil
}
