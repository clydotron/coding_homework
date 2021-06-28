package main

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/clydotron/jumpcloud/utils"
)

type HashServer struct {
	hashCount int32
	hashes    map[int]string
	delay     time.Duration
	ts        utils.TimeStats
	ctx       context.Context
}

func NewHashServer(delay time.Duration, ctx context.Context) *HashServer {

	app := &HashServer{
		delay:  delay,
		hashes: make(map[int]string),
		ctx:    ctx,
	}
	return app
}

func (hs *HashServer) Hash(w http.ResponseWriter, req *http.Request) {

	// validate that this is a post:
	if req.Method != "POST" {
		http.Error(w, "Only POST supported.", http.StatusBadRequest)
		return
	}

	// Record how long this call takes. Collect the data within TimeStats
	defer hs.ts.Record(time.Now())

	// parse the form to get the data
	if err := req.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintln("ParseForm failed.", err), http.StatusInternalServerError)
		return
	}

	// get the value for the key 'password' - make sure it has a valid vale
	password := req.FormValue("password")
	if len(password) == 0 {
		http.Error(w, "required key password missing", http.StatusBadRequest)
	}

	// increment the hash count, and return the new value
	hashID := int(atomic.AddInt32(&hs.hashCount, 1))
	fmt.Fprint(w, hashID)

	// Two things happen next:
	// 1. The ID of the pending hash is returned to the caller
	// 2. A goroutine is used to concurrently execute a function to wait the specified delay then
	//  generate the new hash and store it in the map. If the context is cancelled during this wait,
	//  the goroutine will exit immediately and return without generating the hash

	go func(i int, pw string) {

		// sleep in a separate goroutine so that we can also monitor the status of the context
		// (see select below)
		timeout := make(chan bool)
		go func() {
			time.Sleep(hs.delay)
			timeout <- true
		}()

		select {
		case <-hs.ctx.Done():
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

	// extract the POST request id -- must be a number
	id := strings.TrimPrefix(req.URL.Path, "/hash/")
	if len(id) == 0 {
		http.Error(w, "Must provide valid id", http.StatusBadRequest)
	}

	i, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "ID must be a number", http.StatusBadRequest)
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
