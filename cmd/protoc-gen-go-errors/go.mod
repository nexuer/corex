module github.com/nexuer/corex/cmd/protoc-gen-go-errors

go 1.23

replace github.com/nexuer/api => ../../../api

require (
	github.com/nexuer/api v0.0.0-00010101000000-000000000000
	golang.org/x/text v0.20.0
	google.golang.org/protobuf v1.35.1
)
