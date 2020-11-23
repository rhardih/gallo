package lib

import (
	"context"
	"log"
	"time"
)

type WebhooksManager struct {
	cache  RedisProvider
	ctx    context.Context
	ticker *time.Ticker
}

func NewWebooksManager(cache RedisProvider, ctx context.Context, tickRate time.Duration) WebhooksManager {
	return WebhooksManager{cache, ctx, time.NewTicker(tickRate * time.Second)}
}

func (w WebhooksManager) getTokens() []string {
	tokens := make([]string, 0)
	var keys []string
	var cursor uint64
	var err error

	for {
		// count is an optional parameter for SCAN, with default value 10
		// https://redis.io/commands/scan
		keys, cursor, err = w.cache.Scan(w.ctx, cursor, "tokens:*", 10).Result()

		if err != nil {
			log.Println(err.Error())
			break
		}

		tokens = append(tokens, keys...)

		if cursor == 0 {
			break
		}
	}

	return tokens
}

func (w WebhooksManager) check() {
	// tokens := w.gettokens()
	// 1. In batches of 10, request webhooks for tokens
	// 2. If one is missing, create a new and move on
}

// 1. Start the time job in one goroutine
// 1.1. Scan all tokens currently in Redis
// 1.2. Check that each token has a webhook
// 1.3. If not create one
// 2. Start the redis psubscribe in another thread

func (w WebhooksManager) Run() {
	log.Println("WebhooksManager Running...")

	// defer w.ticker.Stop()
	w.ticker.Stop()

	for {
		println("WebhooksManager is working...")

		select {
		case <-w.ctx.Done():
			log.Println("WebhooksManager Stopped!")
			return
		case t := <-w.ticker.C:
			log.Println("Tick at", t)
		}
	}
}
