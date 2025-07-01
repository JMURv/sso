package hdl

import "errors"

var (
	ErrInternal      = errors.New("internal error")
	ErrDecodeRequest = errors.New("decode request")
	ErrNoDeviceInfo  = errors.New("no device info provided")
	ErrFileTooLarge  = errors.New("file too large")
)

var (
	ErrToRetrievePathArg  = errors.New("error to retrieve path argument")
	ErrFailedToGetUUID    = errors.New("failed to get uid from context")
	ErrFailedToParseUUID  = errors.New("failed to parse uid")
	ErrFailedToParseRoles = errors.New("failed to parse roles")
)
