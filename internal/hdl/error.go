package hdl

import "errors"

var ErrInternal = errors.New("internal error")
var ErrDecodeRequest = errors.New("decode request")
var ErrNoDeviceInfo = errors.New("no device info provided")

var ErrFailedToGetUUID = errors.New("failed to get uid from context")
var ErrFailedToParseUUID = errors.New("failed to parse uid")
