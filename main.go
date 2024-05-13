package main

import (
	"context"
	"flag"
	"log"
	tgclient "scaling-bot/client/telegram"
	"scaling-bot/consumer/event_consumer.go"
	tgprocessor "scaling-bot/event_processor/telegram"
	"scaling-bot/scaler/yandex"
	"scaling-bot/server"
	"scaling-bot/storage/sqlite"
)

const (
	tgBotHost         = "api.telegram.org"
	sqliteStoragePath = "data/sqlite/storage.db"
	fileStoragePath   = "data/user_storage"
	batchSize         = 100
)

func main() {
	st, err := sqlite.New(context.TODO(), sqliteStoragePath)
	if err != nil {
		log.Fatalf("can't connect to storage: ", err)
	}

	if err := st.Init(); err != nil {
		log.Fatalf("can't init storage: ", err)
	}

	server := server.NewServer()
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}

	processor := tgprocessor.New(
		tgclient.New(tgBotHost, mustToken()),
		st,
		yandex.New(),
	)

	consumer := event_consumer.New(processor, processor, batchSize)
	if err := consumer.Start(); err != nil {
		log.Fatal("service is stopped", err)
	}
}

func mustToken() string {
	token := flag.String("token", "", "telegramm api token")

	flag.Parse()

	if *token == "" {
		log.Fatal("token is unset")
	}

	return *token
}