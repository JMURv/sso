package dto

type UpdateDeviceRequest struct {
	Name string `json:"name" validate:"required"`
}
