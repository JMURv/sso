package validation

import (
	"github.com/JMURv/sso/pkg/model"
)

func PermValidation(u *model.Permission) error {
	if u.Name == "" {
		return ErrMissingName
	}

	return nil
}
