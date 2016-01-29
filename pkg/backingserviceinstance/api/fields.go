package api

import "k8s.io/kubernetes/pkg/fields"

// BackingServiceInstanceToSelectableFields returns a label set that represents the object
func BackingServiceInstanceToSelectableFields(backingService *BackingServiceInstance) fields.Set {
	return fields.Set{
		"metadata.name": backingService.Name,
	}
}
