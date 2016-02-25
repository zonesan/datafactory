package validation

import (
	"k8s.io/kubernetes/pkg/api/validation"
	"k8s.io/kubernetes/pkg/util/fielderrors"
	oapi "github.com/openshift/origin/pkg/api"
	backingserviceinstanceapi "github.com/openshift/origin/pkg/backingserviceinstance/api"
)

func BackingServiceInstanceName(name string, prefix bool) (bool, string) {
	if ok, reason := oapi.MinimalNameRequirements(name, prefix); !ok {
		return ok, reason
	}

	if len(name) < 2 {
		return false, "must be at least 2 characters long"
	}


	return true, ""
}
// ValidateBackingServiceInstance tests required fields for a BackingServiceInstance.
func ValidateBackingServiceInstance(bsi *backingserviceinstanceapi.BackingServiceInstance) fielderrors.ValidationErrorList {
	allErrs := fielderrors.ValidationErrorList{}
	allErrs = append(allErrs, validation.ValidateObjectMeta(&bsi.ObjectMeta, false, BackingServiceInstanceName).Prefix("metadata")...)
	//allErrs = append(allErrs, validateBuildSpec(&build.Spec).Prefix("spec")...)
	return allErrs
}

// ValidateBuildRequest validates a BuildRequest object
func ValidateBackingServiceInstanceUpdate(bsi *backingserviceinstanceapi.BackingServiceInstance, older *backingserviceinstanceapi.BackingServiceInstance) fielderrors.ValidationErrorList {
	allErrs := fielderrors.ValidationErrorList{}
	allErrs = append(allErrs, validation.ValidateObjectMetaUpdate(&bsi.ObjectMeta, &older.ObjectMeta).Prefix("metadata")...)

	allErrs = append(allErrs, ValidateBackingServiceInstance(bsi)...)

	// todo: 
	//if older.Status != bsi.Status {
	//	allErrs = append(allErrs, fielderrors.NewFieldInvalid("Status", bsi.Status.Phase, "status cannot be updated from a terminal state"))
	//}
	return allErrs
}