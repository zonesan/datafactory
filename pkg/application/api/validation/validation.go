package validation

import (
	"reflect"

	"k8s.io/kubernetes/pkg/api/validation"
	"k8s.io/kubernetes/pkg/util/fielderrors"

	"fmt"
	oapi "github.com/openshift/origin/pkg/api"
	applicationapi "github.com/openshift/origin/pkg/application/api"
	applicationutil "github.com/openshift/origin/pkg/application/util"
)

func ValidateApplicationName(name string, prefix bool) (bool, string) {
	if ok, reason := oapi.MinimalNameRequirements(name, prefix); !ok {
		return ok, reason
	}

	if len(name) < 2 {
		return false, "must be at least 2 characters long"
	}

	return true, ""
}

func ValidationApplicationItems(items applicationapi.ItemList) (bool, string) {
	for _, item := range items {
		if !applicationutil.Contains(applicationapi.ApplicationItemSupportKinds, item.Kind) {
			return false, fmt.Sprintf("item unsupport selected kind %s", item.Kind)
		}

		if len(item.Name) < 2 {
			return false, "item name must be at least 2 characters long"
		}

		if ok, reason := oapi.MinimalNameRequirements(item.Name, false); !ok {
			return ok, reason
		}
	}
	return true, ""
}

// ValidateApplication tests required fields for a Application.
// This should only be called when creating a application (not on update),
// since its name validation is more restrictive than default namespace name validation
func ValidateApplication(application *applicationapi.Application) fielderrors.ValidationErrorList {
	result := fielderrors.ValidationErrorList{}
	result = append(result, validation.ValidateObjectMeta(&application.ObjectMeta, false, ValidateApplicationName).Prefix("metadata")...)

	if ok, err := ValidationApplicationItems(application.Spec.Items); !ok {
		result = append(result, fielderrors.NewFieldInvalid("items", application.Spec.Items, err))
	}

	return result
}

// ValidateApplicationUpdate tests to make sure a application update can be applied.  Modifies newApplication with immutable fields.
func ValidateApplicationUpdate(newApplication *applicationapi.Application, oldApplication *applicationapi.Application) fielderrors.ValidationErrorList {
	allErrs := fielderrors.ValidationErrorList{}
	allErrs = append(allErrs, validation.ValidateObjectMetaUpdate(&newApplication.ObjectMeta, &oldApplication.ObjectMeta).Prefix("metadata")...)
	allErrs = append(allErrs, ValidateApplication(newApplication)...)

	if !reflect.DeepEqual(newApplication.Spec.Finalizers, oldApplication.Spec.Finalizers) {
		allErrs = append(allErrs, fielderrors.NewFieldInvalid("spec.finalizers", oldApplication.Spec.Finalizers, "field is immutable"))
	}
	if !reflect.DeepEqual(newApplication.Status, oldApplication.Status) {
		allErrs = append(allErrs, fielderrors.NewFieldInvalid("status", oldApplication.Spec.Finalizers, "field is immutable"))
	}

	for name, value := range newApplication.Labels {
		if value != oldApplication.Labels[name] {
			allErrs = append(allErrs, fielderrors.NewFieldInvalid("metadata.labels["+name+"]", value, "field is immutable, , try updating the namespace"))
		}
	}
	for name, value := range oldApplication.Labels {
		if _, inNew := newApplication.Labels[name]; !inNew {
			allErrs = append(allErrs, fielderrors.NewFieldInvalid("metadata.labels["+name+"]", value, "field is immutable, try updating the namespace"))
		}
	}

	return allErrs
}
