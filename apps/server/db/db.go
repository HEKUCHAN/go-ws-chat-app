package db

import (
	"context"
	"log"
	"server/ent"

	_ "github.com/mattn/go-sqlite3"
)

var Client *ent.Client

func Init() {
	var err error
	Client, err = ent.Open("sqlite3", "file:chat.db?_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	if err := Client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
}
