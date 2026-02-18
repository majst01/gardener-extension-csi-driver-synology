package validation

import (
	"net/url"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"

	config "github.com/metal-stack/gardener-extension-csi-driver-synology/pkg/apis/config"
)

func ValidateConfiguration(cfg *config.ControllerConfiguration) field.ErrorList {
	var allErrs field.ErrorList
	fldPath := field.NewPath("controllerConfiguration")

	// synology
	synPath := fldPath.Child("synology")

	if cfg.SynologyConfig.URL == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("synologyURL"), "must be set"))
	} else {
		if _, err := url.ParseRequestURI(cfg.SynologyConfig.URL); err != nil {
			allErrs = append(allErrs, field.Invalid(synPath.Child("synologyURL"), cfg.SynologyConfig.URL, "must be a valid URL"))
		}
	}
	// secret ref required (name of a Secret that holds credentials)
	if strings.TrimSpace(cfg.SynologyConfig.SecretRef) == "" {
		allErrs = append(allErrs, field.Required(synPath.Child("secretRef"), "must be set"))
	}

	// storageClasses.iscsi.parameters
	scPath := synPath.Child("storageClasses").Child("iscsi").Child("parameters")
	params := cfg.SynologyConfig.StorageClasses.ISCSI.Parameters
	if len(params) == 0 {
		allErrs = append(allErrs, field.Required(scPath, "must be set"))
		return allErrs
	}

	return allErrs
}
