package etcd

import (
	"errors"
	"fmt"
	
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/registry/generic"
	etcdgeneric "k8s.io/kubernetes/pkg/registry/generic/etcd"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/storage"
	"k8s.io/kubernetes/pkg/watch"
	"k8s.io/kubernetes/pkg/util"
	
	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api/unversioned"

	//backingserviceregistry "github.com/openshift/origin/pkg/backingservice/registry"
	backingserviceinstanceapi "github.com/openshift/origin/pkg/backingserviceinstance/api"
	backingserviceinstanceregistry "github.com/openshift/origin/pkg/backingserviceinstance/registry/backingserviceinstance"
	//backingserviceinstancecontroller "github.com/openshift/origin/pkg/backingserviceinstance/controller"
	deployconfigregistry "github.com/openshift/origin/pkg/deploy/registry/deployconfig"
)

const BackingServiceInstancePath = "/backingserviceinstances"

type REST struct {
	store *etcdgeneric.Etcd
}

func NewREST(s storage.Interface) *REST {
	store := &etcdgeneric.Etcd{
		NewFunc:     func() runtime.Object { return &backingserviceinstanceapi.BackingServiceInstance{} },
		NewListFunc: func() runtime.Object { return &backingserviceinstanceapi.BackingServiceInstanceList{} },
		KeyRootFunc: func(ctx kapi.Context) string {
			return etcdgeneric.NamespaceKeyRootFunc(ctx, BackingServiceInstancePath)
		},
		KeyFunc: func(ctx kapi.Context, id string) (string, error) {
			return etcdgeneric.NamespaceKeyFunc(ctx, BackingServiceInstancePath, id)
		},
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*backingserviceinstanceapi.BackingServiceInstance).Name, nil
		},
		PredicateFunc: func(label labels.Selector, field fields.Selector) generic.Matcher {
			return backingserviceinstanceregistry.Matcher(label, field)
		},
		EndpointName: "backingserviceinstance",

		CreateStrategy: backingserviceinstanceregistry.BsiStrategy,
		UpdateStrategy: backingserviceinstanceregistry.BsiStrategy,

		ReturnDeletedObject: false,

		Storage: s,
	}
	
	return &REST{store: store}
}

func (r *REST) New() runtime.Object {
	return r.store.NewFunc()
}

func (r *REST) NewList() runtime.Object {
	return r.store.NewListFunc()
}

func (r *REST) Get(ctx kapi.Context, name string) (runtime.Object, error) {
	return r.store.Get(ctx, name)
}

func (r *REST) List(ctx kapi.Context, label labels.Selector, field fields.Selector) (runtime.Object, error) {
	return r.store.List(ctx, label, field)
}

func (r *REST) Create(ctx kapi.Context, obj runtime.Object) (runtime.Object, error) {
	return r.store.Create(ctx, obj)
}

func (r *REST) Update(ctx kapi.Context, obj runtime.Object) (runtime.Object, bool, error) {
	return r.store.Update(ctx, obj)
}

func (r *REST) Delete(ctx kapi.Context, name string, options *kapi.DeleteOptions) (runtime.Object, error) {
	bsiObj, err := r.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	bsi := bsiObj.(*backingserviceinstanceapi.BackingServiceInstance)

	if bsi.DeletionTimestamp.IsZero() {
		now := unversioned.Now()
		bsi.DeletionTimestamp = &now
		//bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseInactive
		bsi.Spec.InstanceID = "" // notify controller to delete
		result, _, err := r.store.Update(ctx, bsi)
		return result, err
	}

	return r.store.Delete(ctx, name, options)
}

func (r *REST) Watch(ctx kapi.Context, label labels.Selector, field fields.Selector, resourceVersion string) (watch.Interface, error) {
	return r.store.Watch(ctx, label, field, resourceVersion)
}

//============================================

func NewBindingREST(bsir backingserviceinstanceregistry.Registry, dcr deployconfigregistry.Registry) *BindingREST {
	return &BindingREST{
		backingServiceInstanceRegistry: bsir,
		deployConfigRegistry: dcr,
	}
}

type BindingREST struct {
	backingServiceInstanceRegistry backingserviceinstanceregistry.Registry
	deployConfigRegistry deployconfigregistry.Registry
}

func (r *BindingREST) New() runtime.Object {
	return &backingserviceinstanceapi.BindingRequestOptions{}
}

func (r *BindingREST) Create(ctx kapi.Context, obj runtime.Object) (runtime.Object, error) {
	glog.Infoln("to create a bsi binding.")
	
	bro := obj.(*backingserviceinstanceapi.BindingRequestOptions)
	if bro.BindKind != backingserviceinstanceapi.BindKind_DeploymentConfig {
		return nil, fmt.Errorf("unsupported bind type: %s", bro.BindKind)
	}
	// todo: check bro.BindResourceVersion
	
	bsi, err := r.backingServiceInstanceRegistry.GetBackingServiceInstance(ctx, bro.Name)
	if err != nil {
		return nil, err
	}
	
	if bsi.Status.Phase != backingserviceinstanceapi.BackingServiceInstancePhaseActive {
		return nil, errors.New("back service instance is not in active phase")
	}
	
	if bsi.Spec.Bound {
		return nil, errors.New("back service instance already bound")
	}
	
	//bs, err := r.backingServiceRegistry.GetBackingService(ctx, bsi.Spec.BackingServiceName)
	//if err != nil {
	//	return nil, err
	//}
	
	dc, err := r.deployConfigRegistry.GetDeploymentConfig(ctx, bro.ResourceName)
	if err != nil {
		return nil, err
	}
	_ = dc
	
	// update bsi
	
	bsi.Spec.BindUuid = string(util.NewUUID())
	bsi.Spec.BindDeploymentConfig = bro.ResourceName // dc.Name
	
	bsi, err = r.backingServiceInstanceRegistry.UpdateBackingServiceInstance(ctx, bsi)
	if err != nil {
		return nil, err
	}
	
	// ...
	
	return bsi, nil
}

func (r *BindingREST) Delete(ctx kapi.Context, name string, options *kapi.DeleteOptions) (runtime.Object, error) {
	glog.Infoln("to delete a bsi binding")
	
	bsi, err := r.backingServiceInstanceRegistry.GetBackingServiceInstance(ctx, name)
	if err != nil {
		return nil, err
	}
	
	bsi.Spec.BindUuid = "" // notify controller to unbind
	
	bsi, err = r.backingServiceInstanceRegistry.UpdateBackingServiceInstance(ctx, bsi)
	if err != nil {
		return nil, err
	}

	return bsi, nil
}

