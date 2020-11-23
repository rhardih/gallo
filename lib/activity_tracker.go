package lib

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/adlio/trello"
)

type WebhookManager struct {
	client *trello.Client
	token  string
}

func (w WebhookManager) start(done chan bool) {
	ticker := time.NewTicker(5 * time.Minute)

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			fmt.Println("Tick at", t)
		}
	}
}

func NewWebhookManager(token string, key string) WebhookManager {
	wm := WebhookManager{
		trello.NewClient(key, token),
		token,
	}

	done := make(chan bool)

	go wm.start(done)

	return wm
}

// The activity tracker is used as a means to keep a record of when a specific
// user was last "active". It's meant to be used by calling .Track() on every
// incoming request by a logged in user. At a later stage, .ActiveTokens() can
// return a list of currently active tokens, which was all seen active within
// some time limit.
type ActivityTracker struct {
	tokensMutex          sync.RWMutex
	tokensChan           chan string
	tokensLastActive     map[string]time.Time
	webhookManagers      map[string]WebhookManager
	webhookManagersMutex sync.Mutex
}

func NewActivityTracker(ctx context.Context) *ActivityTracker {
	activityTracker := &ActivityTracker{
		tokensMutex:          sync.RWMutex{},
		tokensChan:           make(chan string),
		tokensLastActive:     make(map[string]time.Time),
		webhookManagersMutex: sync.Mutex{},
	}

	go activityTracker.start(ctx)

	return activityTracker
}

func (a ActivityTracker) start(ctx context.Context) {
	for {
		select {
		case token := <-a.tokensChan:
			a.updateTokenTimestamp(token)

			a.webhookManagersMutex.Lock()

			// User with token is active
			// 1. Check if there's a WebhookManager for the token
			// 2. If not, create a new one

			if manager, ok := a.webhookManagers[token]; ok {
			} else {
				a.webhookManagers[token] = NewWebhookManager(MustGetEnv("TRELLO_KEY"), token)

				log.Println(manager)
			}

			a.webhookManagersMutex.Unlock()

		case <-ctx.Done():
			close(a.tokensChan)
			log.Println("ActivityTracker closed tokensChan on context cancellation")
			return
		}
	}
}

func (a ActivityTracker) updateTokenTimestamp(token string) {
	a.tokensMutex.Lock()
	a.tokensLastActive[token] = time.Now()
	a.tokensMutex.Unlock()
}

// Logs a new request by token
func (a ActivityTracker) Track(token string) {
	log.Println("activity from", token)
	a.tokensChan <- token
}

// Returns a slice of the tokens, that have been seen within some number of
// hours
func (a ActivityTracker) ActiveTokens(withinMinutes int) []string {
	a.tokensMutex.RLock()
	defer a.tokensMutex.RUnlock()

	tokens := []string{}

	for token, t := range a.tokensLastActive {
		if int(time.Since(t).Minutes()) < withinMinutes {
			tokens = append(tokens, token)
		}
	}

	return tokens
}
