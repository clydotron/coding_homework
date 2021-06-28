package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type application struct {
	hs     *HashServer
	cancel context.CancelFunc
}

func (app *application) shutdown(w http.ResponseWriter, req *http.Request) {

	w.Write([]byte("Shutting down."))

	// call the cancel function for the context
	app.cancel()
}

func main() {

	// allow the user to specify the port and delay duration
	port := flag.Int("port", 8080, "port to listen on")
	delay := flag.Int("delay", 15, "delay (in seconds) for hash")
	flag.Parse()

	// create a context with cancel - cancel will be called from two places:
	// 1. if an os.interrupt is received
	// 2. if /shutdown is called
	ctx, cancel := context.WithCancel(context.Background())

	app := &application{
		hs:     NewHashServer(time.Duration(*delay)*time.Second, ctx),
		cancel: cancel,
	}

	// listen for an interrupt from the OS (ctrl-c)
	osSignalCh := make(chan os.Signal, 1)
	signal.Notify(osSignalCh, os.Interrupt)

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-osSignalCh:
			cancel()
		}
	}()

	// set up the routes
	mux := http.NewServeMux()
	mux.HandleFunc("/hash", app.hs.Hash)
	mux.HandleFunc("/hash/", app.hs.GetHash)
	mux.HandleFunc("/stats", app.hs.Stats)
	mux.HandleFunc("/shutdown", app.shutdown)

	// create the server: call ListenAndServe from its own goroutine
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: mux,
	}

	go func() {
		fmt.Printf("Listening on: %v delay: %v\n", *port, *delay)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe:%+s\n", err)
		}
	}()

	// wait until someone calls the cancel function for the context
	<-ctx.Done()

	ctxShutDown, cancelFcn := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancelFcn()
	}()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("Server Shutdown Failed:%+s", err)
	}

	fmt.Println("All done.")
}
