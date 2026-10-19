package entx

import (
	"context"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/rs/xid"
)

// Mixin definition

// XIDMixin Package xid is a globally unique ID generator library that uses the Mongo Object ID algorithm to
// generate a 12 byte, 20 character ID with no configuration.
// The xid package comes with database/sql sql.Scanner and driver.Valuer interfaces required by Ent for serialization.
type XIDMixin struct {
	mixin.Schema

	StorageKey string
}

// Fields of the XID.
func (x XIDMixin) Fields() []ent.Field {
	return StringIDMixin{
		StorageKey: x.StorageKey,
		DefaultFunc: func() string {
			return xid.New().String()
		},
	}.Fields()
}

type StringIDMixin struct {
	mixin.Schema
	StorageKey  string
	DefaultFunc func() string
}

func (s StringIDMixin) Fields() []ent.Field {
	if s.StorageKey == "" {
		s.StorageKey = "id"
	}
	f := field.String(s.StorageKey).
		Immutable().
		NotEmpty().
		Unique()

	if s.DefaultFunc != nil {
		f.DefaultFunc(s.DefaultFunc)
	}

	return []ent.Field{
		f,
	}
}

// TimeMixin implements the ent.Mixin for sharing
// time fields with package schemas.
type TimeMixin struct {
	mixin.Schema
}

func (TimeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Immutable().
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// SoftDeleteMixin implements the soft delete pattern for schemas.
type SoftDeleteMixin struct {
	mixin.Schema
}

// Fields of the SoftDeleteMixin.
func (SoftDeleteMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

type softDeleteKey struct{}

// SoftDeleteAnnotation marks schemas whose generated clients support soft deletion.
type SoftDeleteAnnotation struct{}

func (SoftDeleteAnnotation) Name() string { return "SoftDelete" }

func (SoftDeleteMixin) Annotations() []schema.Annotation {
	return []schema.Annotation{SoftDeleteAnnotation{}}
}

// SkipSoftDelete returns a new context that skips the soft-delete interceptor/mutators.
func SkipSoftDelete(parent context.Context) context.Context {
	return context.WithValue(parent, softDeleteKey{}, true)
}

// SoftDeleteSkipped reports whether filtering and deletion conversion are disabled.
func SoftDeleteSkipped(ctx context.Context) bool {
	skip, _ := ctx.Value(softDeleteKey{}).(bool)
	return skip
}
