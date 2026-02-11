package validation

import (
	"net/url"

	"k8s.io/apimachinery/pkg/util/validation/field"

	config "github.com/metal-stack/gardener-extension-csi-driver-synology/pkg/apis/config"
)

func ValidateConfiguration(cfg *config.ControllerConfiguration) field.ErrorList {
	var allErrs field.ErrorList
	fldPath := field.NewPath("controllerConfiguration")

	if cfg.SynologyURL == "" {
		allErrs = append(allErrs,
			field.Required(fldPath.Child("synologyURL"), "must be set"),
		)
	} else {
		if _, err := url.ParseRequestURI(cfg.SynologyURL); err != nil {
			allErrs = append(allErrs,
				field.Invalid(fldPath.Child("synologyURL"), cfg.SynologyURL, "must be a valid URL"),
			)
		}
	}

	if cfg.AdminUsername == "" {
		allErrs = append(allErrs,
			field.Required(fldPath.Child("adminUsername"), "must be set"),
		)
	}

	if cfg.AdminPassword == "" {
		allErrs = append(allErrs,
			field.Required(fldPath.Child("adminPassword"), "must be set"),
		)
	}

	return allErrs
}
