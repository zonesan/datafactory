package etcd

import (
	"errors"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/registry/generic"
	etcdgeneric "k8s.io/kubernetes/pkg/registry/generic/etcd"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/storage"
	"k8s.io/kubernetes/pkg/watch"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api/unversioned"

	"github.com/openshift/origin/pkg/backingserviceinstance/api"
	backingserviceinstance "github.com/openshift/origin/pkg/backingserviceinstance/registry/backingserviceinstance"
	//backingserviceinstanceapi "github.com/openshift/origin/pkg/backingserviceinstance/registry/backingserviceinstance/api"
)

type BackingServiceInstanceStorage struct {
	BackingServiceInstance *REST
	Binding                *BindingREST
}

const BackingServiceInstancePath = "/backingserviceinstances"

type REST struct {
	store *etcdgeneric.Etcd
}

func NewREST(s storage.Interface) BackingServiceInstanceStorage {
	store := &etcdgeneric.Etcd{
		NewFunc:     func() runtime.Object { return &api.BackingServiceInstance{} },
		NewListFunc: func() runtime.Object { return &api.BackingServiceInstanceList{} },
		KeyRootFunc: func(ctx kapi.Context) string {
			return etcdgeneric.NamespaceKeyRootFunc(ctx, BackingServiceInstancePath)
		},
		KeyFunc: func(ctx kapi.Context, id string) (string, error) {
			return etcdgeneric.NamespaceKeyFunc(ctx, BackingServiceInstancePath, id)
		},
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*api.BackingServiceInstance).Name, nil
		},
		PredicateFunc: func(label labels.Selector, field fields.Selector) generic.Matcher {
			return backingserviceinstance.Matcher(label, field)
		},
		EndpointName: "backingserviceinstance",

		CreateStrategy: backingserviceinstance.BsiStrategy,
		UpdateStrategy: backingserviceinstance.BsiStrategy,

		ReturnDeletedObject: false,

		Storage: s,
	}

	return BackingServiceInstanceStorage{
		BackingServiceInstance: &REST{store: store},
		Binding:                NewBindingREST(),
	}
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
	bsi := bsiObj.(*api.BackingServiceInstance)

	if bsi.DeletionTimestamp.IsZero() {
		now := unversioned.Now()
		bsi.DeletionTimestamp = &now
		bsi.Status.Phase = api.BackingServiceInstancePhaseInactive
		result, _, err := r.store.Update(ctx, bsi)
		return result, err
	}

	return r.store.Delete(ctx, name, options)
}

func (r *REST) Watch(ctx kapi.Context, label labels.Selector, field fields.Selector, resourceVersion string) (watch.Interface, error) {
	return r.store.Watch(ctx, label, field, resourceVersion)
}

//============================================

type BindingREST struct {
	// todo:
}

func NewBindingREST() *BindingREST {
	return &BindingREST{}
}

func (r *BindingREST) New() runtime.Object {
	return &api.BindingRequest{}
}

func (r *BindingREST) Create(ctx kapi.Context, obj runtime.Object) (runtime.Object, error) {
	glog.Infoln("to create a bsi binding.")

	// todo
	// request := obj.(*api.BindingRequest)
	//
	// return BackingServiceInstance

	return nil, errors.New("not implenmented yet")
}

func (r *BindingREST) Delete(ctx kapi.Context, name string, options *kapi.DeleteOptions) (runtime.Object, error) {
	glog.Infoln("to delete a bsi binding")

	// return BackingServiceInstance

	return nil, errors.New("not implenmented yet")
}
