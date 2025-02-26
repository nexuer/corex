package config

import (
	"reflect"

	"dario.cat/mergo"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type File struct {
	path []string
	cfg  config.Config
}

func NewFile(path ...string) *File {
	source := make([]config.Source, len(path))
	for i, v := range path {
		source[i] = file.NewSource(v)
	}

	return &File{
		path: path,
		cfg:  config.New(config.WithSource(source...)),
	}
}

func (f *File) Scan(val any) error {
	if err := f.cfg.Load(); err != nil {
		return err
	}
	if err := f.cfg.Scan(val); err != nil {
		return err
	}
	return nil
}

func Merge(dst any, src any) error {
	return mergo.Merge(dst, src, mergo.WithOverride, mergo.WithTransformers(transformer{}))
}

var protoEnumType = reflect.TypeOf(new(protoreflect.Enum)).Elem()

type transformer struct {
}

func (t transformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	// proto enum always sets
	if typ.Implements(protoEnumType) {
		if typ.Kind() == reflect.Int32 && typ.Name() != "" {
			return func(dst, src reflect.Value) error {
				if src.Int() > 0 {
					dst.Set(src)
				}
				return nil
			}
		}
	}

	return nil
}
