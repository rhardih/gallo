package gallo

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gallo/app/controllers"
	"gallo/lib"

  "github.com/go-redis/redis/v8"
)

type Application struct {
	Addr string
}

func server(ctx context.Context, wg *sync.WaitGroup) {
	// tell the caller that we've stopped
	defer wg.Done()

	// create a new mux and handler
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("server: received request")
		time.Sleep(5 * time.Second)
		io.WriteString(w, "Finished!\n")
		fmt.Println("server: request finished")
	}))

	// create a server
	srv := &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil {
			fmt.Printf("Listen : %s\n", err)
		}
	}()

	<-ctx.Done()
	fmt.Println("server: caller has told us to stop")

	// shut down gracefully, but wait no longer than 5 seconds before halting
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ignore error since it will be "Err shutting down server : context canceled"
	srv.Shutdown(shutdownCtx)

	fmt.Println("server gracefully stopped")
}

func tick(ctx context.Context, wg *sync.WaitGroup) {
	// tell the caller we've stopped
	defer wg.Done()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			fmt.Printf("tick: tick %s\n", now.UTC().Format("20060102-150405.000000000"))
		case <-ctx.Done():
			fmt.Println("tick: caller has told us to stop")
			return
		}
	}
}

func tock(ctx context.Context, wg *sync.WaitGroup) {
	// tell the caller we've stopped
	defer wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			fmt.Printf("tock: tock %s\n", now.UTC().Format("20060102-150405.000000000"))
		case <-ctx.Done():
			fmt.Println("tock: caller has told us to stop")
			return
		}
	}
}

func (app Application) Run() {
	// // create a context that we can cancel
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// // a WaitGroup for the goroutines to tell us they've stopped
	// wg := sync.WaitGroup{}

	// // a channel for `tick()` to tell us they've stopped
	// wg.Add(1)
	// go tick(ctx, &wg)

	// // a channel for `tock()` to tell us they've stopped
	// wg.Add(1)
	// go tock(ctx, &wg)

	// // run `server` in it's own goroutine
	// wg.Add(1)
	// go server(ctx, &wg)

	// // listen for C-c
	// c := make(chan os.Signal, 1)
	// // signal.Notify(c, os.Interrupt)
	// signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	// <-c
	// fmt.Println("main: received C-c - shutting down")

	// // tell the goroutines to stop
	// fmt.Println("main: telling goroutines to stop")
	// cancel()

	// // and wait for them both to reply back
	// wg.Wait()
	// fmt.Println("main: all goroutines have told us they've finished")

	cancelCtx, cancel := context.WithCancel(context.Background())
	wm := lib.NewWebooksManager(
    redis.NewClient(&redis.Options{
      Addr: lib.MustGetEnv("REDIS_ADDR"),
    }),
		cancelCtx,
		10,
	)
	srv := &http.Server{
		Handler:      controllers.NewRouter(),
		Addr:         app.Addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := srv.ListenAndServe(); err != nil {
			fmt.Printf("ListenAndServe : %s\n", err)
		}
	}()

	log.Println("Server started.")

	wg.Add(1)
	go func() {
		defer wg.Done()

		wm.Run()
	}()

	<-done

	log.Println("Server stopping...")

	timeoutCtx, _ := context.WithTimeout(cancelCtx, 5*time.Second)
	if err := srv.Shutdown(timeoutCtx); err != nil {
		fmt.Printf("Server Shutdown Failed:%+v", err)
	}

	cancel()

	wg.Wait()

	log.Println("Server stopped!")
}
