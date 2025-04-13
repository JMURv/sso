package db

import (
	md "github.com/JMURv/sso/internal/models"
	"strconv"
	"strings"
)

func ScanPerms(req []string) ([]md.Permission, error) {
	res := make([]md.Permission, 0, len(req))
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
			res, md.Permission{
				ID:          id,
				Name:        parts[1],
				Description: parts[2],
			},
		)
	}
	return res, nil
}
