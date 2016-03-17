package controller

import (
	kclient "k8s.io/kubernetes/pkg/client/unversioned"

	"fmt"
	backingserviceapi "github.com/openshift/origin/pkg/backingservice/api"
	osclient "github.com/openshift/origin/pkg/client"
	servicebrokerapi "github.com/openshift/origin/pkg/servicebroker/api"
	servicebrokerclient "github.com/openshift/origin/pkg/servicebroker/client"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"strconv"
	"time"
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

const BSNS = "openshift"

type fatalError string

func (e fatalError) Error() string {
	return "fatal error handling ServiceBrokerController: " + string(e)
}

// Handle processes a namespace and deletes content in origin if its terminating
func (c *ServiceBrokerController) Handle(sb *servicebrokerapi.ServiceBroker) (err error) {

	if sb.Spec.Url == "" {
		return nil
	}

	switch sb.Status.Phase {
	case servicebrokerapi.ServiceBrokerNew:
		if getRetryTime(sb) <= 3 {
			if Ping(sb, 10) {
				setRetryTime(sb)

				services, err := c.ServiceBrokerClient.Catalog(sb.Spec.Url, sb.Spec.UserName, sb.Spec.Password)
				if err != nil {
					c.Client.ServiceBrokers().Update(sb)
					return err
				}

				errs := []error{}
				for _, v := range services.Services {

					if err := backingServiceHandler(c.Client, newBackingService(sb.Name, v)); err != nil {
						errs = append(errs, err)
					}
				}
				if len(errs) == 0 {
					removeRetryTime(sb)
					sb.Status.Phase = servicebrokerapi.ServiceBrokerActive
				}

				c.Client.ServiceBrokers().Update(sb)
				return nil
			}
		} else {
			sb.Status.Phase = servicebrokerapi.ServiceBrokerFailed
			c.Client.ServiceBrokers().Update(sb)

			c.inActiveBackingService(sb.Name)
			return nil
		}

	case servicebrokerapi.ServiceBrokerDeleting:
		c.inActiveBackingService(sb.Name)
		c.Client.ServiceBrokers().Delete(sb.Name)
		return nil
	case servicebrokerapi.ServiceBrokerActive:
		if Ping(sb, 60) {
			_, err := c.ServiceBrokerClient.Catalog(sb.Spec.Url, sb.Spec.UserName, sb.Spec.Password)
			if err != nil {
				sb.Status.Phase = servicebrokerapi.ServiceBrokerFailed
				c.Client.ServiceBrokers().Update(sb)

				c.inActiveBackingService(sb.Name)
				return err
			}

			c.Client.ServiceBrokers().Update(sb)
			return nil
		}
	case servicebrokerapi.ServiceBrokerFailed:
		if Ping(sb, 60) {
			_, err := c.ServiceBrokerClient.Catalog(sb.Spec.Url, sb.Spec.UserName, sb.Spec.Password)
			if err != nil {
				c.Client.ServiceBrokers().Update(sb)
				return err
			}

			sb.Status.Phase = servicebrokerapi.ServiceBrokerActive
			c.Client.ServiceBrokers().Update(sb)

			c.ActiveBackingService(sb.Name)
			return nil
		}

	}

	return nil
}

func (c *ServiceBrokerController) inActiveBackingService(serviceBrokerName string) {
	selector, _ := labels.Parse(servicebrokerapi.ServiceBrokerLabel + "=" + serviceBrokerName)

	bsList, err := c.Client.BackingServices(BSNS).List(selector, fields.Everything())
	if err == nil {
		for _, bsvc := range bsList.Items {
			if bsvc.Status.Phase != backingserviceapi.BackingServicePhaseInactive {
				bsvc.Status.Phase = backingserviceapi.BackingServicePhaseInactive
				c.Client.BackingServices(BSNS).Update(&bsvc)
			}
		}
	}
}

func (c *ServiceBrokerController) ActiveBackingService(serviceBrokerName string) {
	selector, _ := labels.Parse(servicebrokerapi.ServiceBrokerLabel + "=" + serviceBrokerName)

	bsList, err := c.Client.BackingServices(BSNS).List(selector, fields.Everything())
	if err == nil {
		for _, bsvc := range bsList.Items {
			if bsvc.Status.Phase != backingserviceapi.BackingServicePhaseActive {
				bsvc.Status.Phase = backingserviceapi.BackingServicePhaseActive
				c.Client.BackingServices(BSNS).Update(&bsvc)
			}
		}
	}
}

func Ping(sb *servicebrokerapi.ServiceBroker, pingSecond int64) bool {
	if sb.Annotations == nil {
		sb.Annotations = map[string]string{}
	}

	lastTimeStr := sb.Annotations[servicebrokerapi.ServiceBrokerLastPingTime]
	if len(lastTimeStr) == 0 {
		sb.Annotations[servicebrokerapi.ServiceBrokerLastPingTime] = fmt.Sprintf("%d", time.Now().UnixNano())
		return true
	}

	lastPing, err := strconv.Atoi(lastTimeStr)
	if err != nil {
		sb.Annotations[servicebrokerapi.ServiceBrokerLastPingTime] = fmt.Sprintf("%d", time.Now().UnixNano())
		return false
	}

	if (time.Now().UnixNano()-int64(lastPing))/1e9 < pingSecond {
		return false
	}

	sb.Annotations[servicebrokerapi.ServiceBrokerLastPingTime] = fmt.Sprintf("%d", time.Now().UnixNano())
	return true
}

func getRetryTime(sb *servicebrokerapi.ServiceBroker) int {
	retries := sb.Annotations[servicebrokerapi.ServiceBrokerNewRetryTimes]
	if len(retries) == 0 {
		return 0
	}

	i, err := strconv.Atoi(retries)
	if err != nil {
		return 0
	}

	return i
}

func setRetryTime(sb *servicebrokerapi.ServiceBroker) {
	if len(sb.Annotations[servicebrokerapi.ServiceBrokerNewRetryTimes]) == 0 {
		sb.Annotations[servicebrokerapi.ServiceBrokerNewRetryTimes] = fmt.Sprintf("%d", 0)
	}

	i, err := strconv.Atoi(sb.Annotations[servicebrokerapi.ServiceBrokerNewRetryTimes])
	if err != nil {
		sb.Annotations[servicebrokerapi.ServiceBrokerNewRetryTimes] = fmt.Sprintf("%d", 1)
		return
	}

	sb.Annotations[servicebrokerapi.ServiceBrokerNewRetryTimes] = fmt.Sprintf("%d", i+1)

	return
}

func removeRetryTime(sb *servicebrokerapi.ServiceBroker) {
	delete(sb.Annotations, servicebrokerapi.ServiceBrokerNewRetryTimes)
}
