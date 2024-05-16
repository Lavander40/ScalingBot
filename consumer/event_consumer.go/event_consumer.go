package event_consumer

import (
	"log"
	"time"
	ep "scaling-bot/event_processor"
)

type Consumer struct {
	fetcher   ep.Fetcher
	processor ep.Processor
	batchSize int
}

func New(fetcher ep.Fetcher, processor ep.Processor, batchSize int) Consumer {
	return Consumer{
		fetcher:   fetcher,
		processor: processor,
		batchSize: batchSize,
	}
}

func (c *Consumer) Start() error {
	for {
		newEvents, err := c.fetcher.Fetch(c.batchSize)
		if err != nil {
			log.Printf("error during consumer start: %s", err.Error())
			continue
		}

		if len(newEvents) == 0 {
			time.Sleep(2 * time.Second)
			continue
		}

		if err := c.handleEvents(newEvents); err != nil {
			log.Print(err)
			continue
		}
	}
}

func (c *Consumer) handleEvents(events []ep.Event) error {
	for _, event := range events {
		if err := c.processor.Process(event); err != nil {
			log.Printf("can't handle event %s", err.Error())
			continue
		}
	}

	return nil
}
