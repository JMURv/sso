package db

import (
	"strings"
	"time"

	md "github.com/JMURv/sso/internal/models"
)

func ScanDevices(req []string) ([]md.Device, error) {
	res := make([]md.Device, 0, len(req))
	for _, v := range req {
		parts := strings.Split(v, "|")
		if len(parts) != 8 {
			continue
		}

		layout := "2006-01-02 15:04:05.999999-07"
		lastActive, err := time.Parse(layout, parts[7])
		if err != nil {
			return nil, err
		}

		res = append(
			res, md.Device{
				ID:         parts[0],
				Name:       parts[1],
				DeviceType: parts[2],
				OS:         parts[3],
				Browser:    parts[4],
				UA:         parts[5],
				IP:         parts[6],
				LastActive: lastActive,
			},
		)
	}
	return res, nil
}
