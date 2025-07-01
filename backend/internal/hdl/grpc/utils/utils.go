package utils

import (
	"context"

	"github.com/JMURv/sso/internal/dto"
)

func ParseDeviceFromContext(ctx context.Context) dto.DeviceRequest {
	return dto.DeviceRequest{
		UA: ctx.Value("ua").(string),
		IP: ctx.Value("ip").(string),
	}
}
