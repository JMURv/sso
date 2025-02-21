package db

import (
	md "github.com/JMURv/sso/internal/models"
	"strconv"
	"strings"
)

func ScanPermissions(perms []string) ([]md.Permission, error) {
	permissions := make([]md.Permission, 0, len(perms))
	for _, perm := range perms {
		parts := strings.Split(perm, "|")
		if len(parts) != 3 {
			continue
		}

		id, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return nil, err
		}

		value, err := strconv.ParseBool(parts[2])
		if err != nil {
			return nil, err
		}

		permissions = append(
			permissions, md.Permission{
				ID:    id,
				Name:  parts[1],
				Value: value,
			},
		)
	}
	return permissions, nil
}
