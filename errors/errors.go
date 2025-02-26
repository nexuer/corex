package errors

import (
	"github.com/go-kratos/kratos/v2/errors"
)

type Error = errors.Error

var (
	Code      = errors.Code
	Reason    = errors.Reason
	FromError = errors.FromError
	Clone     = errors.Clone
	New       = errors.New
	Newf      = errors.Newf
	Errorf    = errors.Errorf
)
