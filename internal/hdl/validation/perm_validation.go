package validation

import (
	"github.com/JMURv/sso/internal/models"
)

func PermValidation(u *models.Permission) error {
	if u.Name == "" {
		return ErrMissingName
	}

	return nil
}
