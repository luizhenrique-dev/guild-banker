package user

import "errors"

var ErrNotFound = errors.New("user not found")

var ErrDisabled = errors.New("user disabled")
var ErrEmailConflict = errors.New("user email conflict")
