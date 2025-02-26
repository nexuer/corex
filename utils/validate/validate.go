package validate

import (
	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/protobuf/proto"
)

var (
	protoValidator *protovalidate.Validator
)

func init() {
	var err error
	protoValidator, err = protovalidate.New()
	if err != nil {
		panic(err)
	}
}

func Message(message proto.Message) error {
	return protoValidator.Validate(message)
}
