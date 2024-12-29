package discovery

import "errors"

var ErrFailedToRegister = errors.New("failed to register service")
var ErrFailedToDeregister = errors.New("failed to deregister service")
var ErrFailedToFindService = errors.New("failed to find service")
