package auth

import (
	"crypto/sha256"
	"fmt"
	"github.com/JMURv/sso/internal/dto"
	md "github.com/JMURv/sso/internal/models"
	"github.com/mssola/useragent"
)

// TODO: Make better device recognition
func GenerateDevice(d *dto.DeviceRequest) md.Device {
	hash := sha256.Sum256([]byte(d.IP + d.UA))
	ua := useragent.New(d.UA)
	dt := parseDeviceType(ua)
	bName, _ := ua.Browser()

	return md.Device{
		ID:         fmt.Sprintf("%x", hash[:8]),
		Name:       "My " + dt,
		DeviceType: dt,
		OS:         ua.OS(),
		Browser:    bName,
		UA:         d.UA,
		IP:         d.IP,
	}
}

func parseDeviceType(ua *useragent.UserAgent) string {
	switch {
	case ua.Mobile():
		return "mobile"
	case ua.Bot():
		return "bot"
	case ua.Mozilla() != "":
		return "desktop"
	case !ua.Mobile() && !ua.Bot() && ua.Mozilla() == "":
		return "tablet"
	default:
		return "unknown"
	}
}
