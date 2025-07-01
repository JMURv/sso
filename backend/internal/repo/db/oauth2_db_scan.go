package db

import (
	"strconv"
	"strings"

	md "github.com/JMURv/sso/internal/models"
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

func ScanRoles(req []string) ([]md.Role, error) {
	res := make([]md.Role, 0, len(req))
	for _, v := range req {
		parts := strings.Split(v, "|")
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
