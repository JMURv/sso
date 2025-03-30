package db

import (
	md "github.com/JMURv/sso/internal/models"
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
