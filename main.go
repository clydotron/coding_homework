package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

type application struct {
	hs *HashServer
}

func (app *application) shutdown(w http.ResponseWriter, req *http.Request) {
	//prevent duplicate calls...

	// this will block until the hash engine is done:
	app.hs.Shutdown()

	w.Write([]byte("Shutting down."))

	// buy some time: give this call the chance to return before the app
	// is terminated
	go func() {
		time.Sleep(50 * time.Millisecond)
		os.Exit(0)
	}()
}

func main() {

	// allow the user to specify the port and delay duration
	port := flag.Int("port", 8080, "port to listen on")
	delay := flag.Int("delay", 5, "delay (in seconds) for hash")
	flag.Parse()

	app := &application{
		hs: NewHashServer(time.Duration(*delay) * time.Second),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/hash", app.hs.Hash)
	mux.HandleFunc("/hash/", app.hs.GetHash)
	mux.HandleFunc("/stats", app.hs.Stats)
	mux.HandleFunc("/shutdown", app.shutdown)

	fmt.Printf("Listening on: %v delay: %v\n", *port, *delay)

	http.ListenAndServe(fmt.Sprintf(":%d", *port), mux)
}
