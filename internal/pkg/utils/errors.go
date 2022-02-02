package utils

import "errors"

//ErrTooBusy indicates too many request to service
var ErrTooBusy error = errors.New("too busy")
