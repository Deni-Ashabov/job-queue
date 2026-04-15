package storage

import "errors"

var (
	ErrJobNotFound = errors.New("job not found")
	ErrJobExists   = errors.New("job exists")
)
