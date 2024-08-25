package main

import (
	"fmt"
	"github.com/fiatjaf/eventstore/postgresql"
	"github.com/fiatjaf/khatru"
	"github.com/kelseyhightower/envconfig"
	"net/http"
)

type settings struct {
	Relay struct {
		Name        string `envconfig:"RELAY_NAME"`
		PubKey      string `envconfig:"RELAY_PUBKEY"`
		Description string `envconfig:"RELAY_DESCRIPTION"`
		IconURL     string `envconfig:"RELAY_ICON_URL"`
	}
	DatabaseURL string `envconfig:"DATABASE_URL" default:"postgres://postgres:example@localhost:5432/postgres?sslmode=disable"`
}

func main() {
	var s settings
	envconfig.MustProcess("", &s)

	db := postgresql.PostgresBackend{DatabaseURL: s.DatabaseURL}
	if err := db.Init(); err != nil {
		panic(err)
	}

	relay := khatru.NewRelay()

	relay.Info.Name = s.Relay.Name
	relay.Info.PubKey = s.Relay.PubKey
	relay.Info.Description = s.Relay.Description
	relay.Info.Icon = s.Relay.IconURL
	relay.StoreEvent = append(relay.StoreEvent, db.SaveEvent)
	relay.QueryEvents = append(relay.QueryEvents, db.QueryEvents)
	relay.CountEvents = append(relay.CountEvents, db.CountEvents)
	relay.DeleteEvent = append(relay.DeleteEvent, db.DeleteEvent)

	relay.Router().Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))

	fmt.Println("Starting relay on :3334 ...")
	if err := http.ListenAndServe(":3334", relay); err != nil {
		panic(err)
	}
}
