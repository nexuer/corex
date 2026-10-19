package validate

import (
	"buf.build/go/protovalidate"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/middleware/validate"
	"google.golang.org/protobuf/proto"
)

var (
	protoValidator protovalidate.Validator
)

func init() {
	var err error
	protoValidator, err = protovalidate.New()
	if err != nil {
		panic(err)
	}
}

func Message(msg proto.Message) error {
	return protoValidator.Validate(msg)
}

var Validator = validate.Validator

// Protobuf validates protobuf requests using their declared validation rules.
// Requests that do not implement proto.Message are passed through unchanged.
func Protobuf() middleware.Middleware {
	return validate.Validator(func(v any) error {
		if msg, ok := v.(proto.Message); ok {
			return Message(msg)
		}
		return nil
	})
}
