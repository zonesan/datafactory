package application

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/registry/generic"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/fielderrors"

	api "github.com/openshift/origin/pkg/application/api"
	applicationvalidation "github.com/openshift/origin/pkg/application/api/validation"
)

type Strategy struct {
	runtime.ObjectTyper
}

// Strategy is the default logic that applies when creating and updating HostSubnet
// objects via the REST API.
var AppStrategy = Strategy{kapi.Scheme}

func (Strategy) PrepareForUpdate(obj, old runtime.Object) {}

func (Strategy) NamespaceScoped() bool {
	return true
}

func (Strategy) GenerateName(base string) string {
	return base
}

func (Strategy) PrepareForCreate(obj runtime.Object) {
}

func (Strategy) Validate(ctx kapi.Context, obj runtime.Object) fielderrors.ValidationErrorList {
	application := obj.(*api.Application)
	return applicationvalidation.ValidateApplication(application)
}

func (Strategy) AllowCreateOnUpdate() bool {
	return false
}

func (Strategy) AllowUnconditionalUpdate() bool {
	return false
}

// ValidateUpdate is the default update validation for a HostSubnet
func (Strategy) ValidateUpdate(ctx kapi.Context, obj, old runtime.Object) fielderrors.ValidationErrorList {
	return fielderrors.ValidationErrorList{}
}

// Matcher returns a generic matcher for a given label and field selector.
func Matcher(label labels.Selector, field fields.Selector) generic.Matcher {
	return &generic.SelectionPredicate{Label: label, Field: field, GetAttrs: getAttrs}
}

func getAttrs(obj runtime.Object) (objLabels labels.Set, objFields fields.Set, err error) {
	app := obj.(*api.Application)
	return labels.Set(app.Labels), api.ApplicationToSelectableFields(app), nil
}
