package types

import "github.com/cockroachdb/errors"

var (
	ErrInvalidCapacity = errors.New("invalid capacity")
	ErrInvalidVolume   = errors.New("invalid volume")
	ErrInvalidStorage  = errors.New("invalid storage")
	ErrInvalidVolumes  = errors.New("invalid volumes")
	ErrInvalidParams   = errors.New("invalid io parameters")
)
