package auth

import (
	"crypto/sha256"
	"fmt"
)

type DeviceInfo struct {
	UserAgent  string `json:"userAgent"`
	IP         string `json:"ip"`
	OS         string `json:"os"`
	Browser    string `json:"browser"`
	DeviceType string `json:"deviceType"`
}

func NewDeviceInfo(ip, ua string) DeviceInfo {
	return DeviceInfo{
		IP:         ip,
		UserAgent:  ua,
		OS:         parseOS(ua),
		Browser:    parseBrowser(ua),
		DeviceType: parseDeviceType(ua),
	}
}

func (d *DeviceInfo) GenerateDeviceID(ip, ua string) string {
	hash := sha256.Sum256([]byte(ip + ua))
	return fmt.Sprintf("%x", hash[:8])
}

func parseOS(ua string) string {
	return ""
}

func parseBrowser(ua string) string {
	return ""
}

func parseDeviceType(ua string) string {
	return ""
}
