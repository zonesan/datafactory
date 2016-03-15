package controller

import (
	osclient "github.com/openshift/origin/pkg/client"
	backingserviceapi "github.com/openshift/origin/pkg/backingservice/api"
	servicebrokerapi "github.com/openshift/origin/pkg/servicebroker/api"
	"k8s.io/kubernetes/pkg/api/errors"
	"github.com/golang/glog"
)

func newBackingService(name string, spec backingserviceapi.BackingServiceSpec) *backingserviceapi.BackingService {
	bs := new(backingserviceapi.BackingService)
	bs.Spec = backingserviceapi.BackingServiceSpec(spec)
	bs.Annotations = make(map[string]string)
	bs.Name = spec.Name
	bs.GenerateName = name
	bs.Labels = map[string]string{
		servicebrokerapi.ServiceBrokerLabel: name,
	}

	return bs
}

func backingServiceHandler(client osclient.Interface, backingService *backingserviceapi.BackingService) error {
	_, err := client.BackingServices().Get(backingService.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			if _, err := client.BackingServices().Create(backingService); err != nil {
				glog.Errorln("servicebroker create backingservice err ", err)
				return err
			}
		}
	} else {
		if _, err := client.BackingServices().Update(backingService); err != nil {
			glog.Errorln("servicebroker update backingservice err ", err)
			return err
		}
	}

	return nil
}
