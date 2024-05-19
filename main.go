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
	"sync"
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

	tgClient := tgclient.New(tgBotHost, mustToken())

	processor := tgprocessor.New(
		tgClient,
		st,
		yandex.New(),
	)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		server := server.NewServer(st, tgClient)
		if err := server.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		consumer := event_consumer.New(processor, processor, batchSize)
		if err := consumer.Start(); err != nil {
			log.Fatal("service is stopped", err)
		}
	}()

	wg.Wait()
}

func mustToken() string {
	token := flag.String("token", "", "telegramm api token")

	flag.Parse()

	if *token == "" {
		log.Fatal("token is unset")
	}

	return *token
}