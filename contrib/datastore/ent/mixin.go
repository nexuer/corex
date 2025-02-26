package ent

import (
	"time"

	"entgo.io/ent"
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
}

// Fields of the XID.
func (XIDMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("xid").
			GoType(xid.ID{}).
			DefaultFunc(xid.New).
			Immutable().
			Unique(),
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
