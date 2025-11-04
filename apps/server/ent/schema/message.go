package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Message struct {
	ent.Schema
}

func (Message) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").MaxLen(32).NotEmpty(),
		field.String("message").MaxLen(512).NotEmpty(),
		field.Time("created_at").Default(time.Now),
	}
}
