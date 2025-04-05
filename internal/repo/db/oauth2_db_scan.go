package db

import (
	md "github.com/JMURv/sso/internal/models"
	"strconv"
	"strings"
)

func ScanOauth2Connections(conns []string) ([]md.Oauth2Connection, error) {
	res := make([]md.Oauth2Connection, 0, len(conns))
	for _, conn := range conns {
		parts := strings.Split(conn, "|")
		if len(parts) != 2 {
			continue
		}

		res = append(
			res, md.Oauth2Connection{
				Provider:   parts[0],
				ProviderID: parts[1],
			},
		)
	}
	return res, nil
}

func ScanRoles(roles []string) ([]md.Role, error) {
	res := make([]md.Role, 0, len(roles))
	for _, role := range roles {
		parts := strings.Split(role, "|")
		if len(parts) != 3 {
			continue
		}

		id, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return nil, err
		}

		res = append(
			res, md.Role{
				ID:          id,
				Name:        parts[1],
				Description: parts[2],
			},
		)
	}
	return res, nil
}
