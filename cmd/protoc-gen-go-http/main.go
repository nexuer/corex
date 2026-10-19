package main

import (
	"flag"
	"fmt"

	"google.golang.org/protobuf/types/pluginpb"

	"google.golang.org/protobuf/compiler/protogen"
)

const version = "v0.0.1"

var (
	showVersion = flag.Bool("version", false, "print the version and exit")
	prefix      = flag.String("prefix", "", "add prefix to google.api.http path")
)

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-go-http %s\n", version)
		return
	}
	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			generateFile(gen, f, *prefix)
		}
		return nil
	})
}
