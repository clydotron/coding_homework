package main

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/clydotron/jumpcloud/utils"
)

type HashServer struct {
	doneCh chan bool
	count  int32 // atomic
	hashes map[int]string
	delay  time.Duration
	wg     sync.WaitGroup
	ts     utils.TimeStats
}

func NewHashServer(delay time.Duration) *HashServer {

	app := &HashServer{
		delay:  delay,
		hashes: make(map[int]string),
		doneCh: make(chan bool),
	}
	return app
}

func (hs *HashServer) Hash(w http.ResponseWriter, req *http.Request) {

	// move this later?
	startTime := time.Now()
	defer hs.ts.Record(startTime)

	// validate that this is a post:
	if req.Method != "POST" {
		http.Error(w, "Only POST supported.", http.StatusBadRequest)
		return
	}

	if hs.isShuttingDown() {
		w.WriteHeader(http.StatusGone)
		fmt.Fprint(w, "Shutting down.") //@todo use an error code?
		return
	}

	// parse the form to get the data
	if err := req.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintln("ParseForm failed.", err), http.StatusInternalServerError)
		return
	}

	// get the value for the key 'password' - make sure it has a valid vale
	password := req.FormValue("password")
	if len(password) == 0 {
		http.Error(w, "required key password missing", http.StatusPreconditionFailed)
	}

	// increment the hash count, and return the new value
	hashID := int(atomic.AddInt32(&hs.count, 1))
	fmt.Fprint(w, hashID)

	// Two things happen next:
	// 1. The ID of the pending hash is returned to the caller
	// 2. A goroutine is used to concurrently execute a function to wait the specified delay then
	//  generate the new hash and store it in the map. If the doneCh is closed during this wait (via Shutdown),
	//  the goroutine will exit immediately and return without generating the hash

	// increment the wait group (this is used on shutdown)
	hs.wg.Add(1)

	// check if we are shutting down...

	go func(i int, pw string) {
		defer hs.wg.Done()

		// sleep in a separate goroutine so that we can
		timeout := make(chan bool)
		go func() {
			time.Sleep(hs.delay)
			timeout <- true
		}()

		select {
		case <-hs.doneCh:
			return
		case <-timeout:
		}

		// create the sha512 hash function, hash the password, base64 encode it, store in map
		hasher := sha512.New()
		hs.hashes[i] = base64.URLEncoding.EncodeToString(hasher.Sum([]byte(pw)))
	}(hashID, password)
}

func (hs *HashServer) GetHash(w http.ResponseWriter, req *http.Request) {

	// validate that this is a GET:
	if req.Method != "GET" {
		http.Error(w, "Only GET supported.", http.StatusBadRequest)
		return
	}

	if hs.isShuttingDown() {
		w.WriteHeader(http.StatusGone)
		fmt.Fprint(w, "Shutting down.")
		return
	}

	// extract the POST request id -- must be a number
	id := strings.TrimPrefix(req.URL.Path, "/hash/")
	i, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Must be a number", http.StatusBadRequest)
		return
	}

	// check to see if the hash exists: if it does, return it, otherwise report not found.
	if val, ok := hs.hashes[i]; ok {
		fmt.Fprintf(w, val)
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func (hs *HashServer) Stats(w http.ResponseWriter, req *http.Request) {

	// validate that this is a GET:
	if req.Method != "GET" {
		http.Error(w, "Only GET supported.", http.StatusBadRequest)
		return
	}

	// get the stats report and return the json data:
	json, err := json.Marshal(hs.ts.GetReport())
	if err != nil {
		http.Error(w, "Failed to JSON encode time statistics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func (hs *HashServer) Shutdown() {

	// check if already been called:
	if hs.isShuttingDown() {
		return
	}

	// close the done channel
	// will cause any active hash requests to exit
	// prevents new calls to /hash, /hash/ /stats from being processed
	close(hs.doneCh)

	// wait for any pending hash requests to exit
	hs.wg.Wait()
}

// determine if Shutdown as been called: the done channel will return ok=false if so.
func (a *HashServer) isShuttingDown() bool {
	ok := true
	select {
	case _, ok = <-a.doneCh:
	default:
	}
	return !ok
}
