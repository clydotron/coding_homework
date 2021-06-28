package main

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/clydotron/jumpcloud/utils"
)

/*
	Test coverage:
		requests to /hash:
			GET requests rejected
			POST requests are accepted
			must have include a form value of 'password'
			POST request returns an incremented count
*/
func TestHashGetRequest_BadRequest(t *testing.T) {

	app := NewHashServer(2 * time.Second)

	req, err := http.NewRequest("GET", "/hash", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.Hash)

	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestHashNoParameters(t *testing.T) {
	app := NewHashServer(2 * time.Second)

	req, err := http.NewRequest("POST", "/hash", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.Hash)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusPreconditionFailed)
	}
}

func TestHashNoPassword(t *testing.T) {
	app := NewHashServer(2 * time.Second)

	req, err := http.NewRequest("POST", "/hash", strings.NewReader("wrong=value"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.Hash)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusPreconditionFailed {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusPreconditionFailed)
	}
}

func TestHashPostRequest(t *testing.T) {

	app := NewHashServer(2 * time.Second)
	app.count = 10

	req, err := http.NewRequest("POST", "/hash", strings.NewReader("password=boogabooga"))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.Hash)

	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "11" //count + 1
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got [%v] want [%v]",
			rr.Body.String(), expected)
	}

}

// helper functions
func postToHash(app *HashServer, payload string, t *testing.T) int {

	req, err := http.NewRequest("POST", "/hash", strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.Hash)

	handler.ServeHTTP(rr, req)

	id, err := strconv.Atoi(rr.Body.String())
	if err != nil {
		t.Fatal(err)
	}
	return id
}

/*
	Test Coverage: requests to /hash/:id
		request missing the id
		requesting a hash before the delay has elapsed -> NotFound ?
		requesting an unknown hash ID returns notFound
		requesting an invalid hash ID (string) returns
		requesting a valid hash ID after the delay returns StatusOK and the hash.

*/
func TestGetHashTooEarly(t *testing.T) {
	app := NewHashServer(2 * time.Second)
	app.count = 10

	postToHash(app, "password=secret!", t)

	{
		req, err := http.NewRequest("GET", "/hash/11", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(app.GetHash)

		handler.ServeHTTP(rr, req)

		// validate the return: -- should be notFound?

		if rr.Code != http.StatusNotFound {
			t.Errorf("Was expecting http.StatusNotFound, received: %v", rr.Code)
		}
	}

}
func TestGetHashSuccess(t *testing.T) {
	hs := NewHashServer(1 * time.Second)
	hs.count = 10

	// send the post request:
	postToHash(hs, "password=secret!", t)

	// sleep for 1 second (might want to )
	time.Sleep(1050 * time.Millisecond)

	req, err := http.NewRequest("GET", "/hash/11", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(hs.GetHash)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Was expecting http.StatusOK, received: %v", rr.Code)
	}
	hasher := sha512.New()
	expected := base64.URLEncoding.EncodeToString(hasher.Sum([]byte("secret!")))
	if expected != rr.Body.String() {
		t.Error("Hashes do not match!")
	}

}

func TestPostHashAfterShutdown(t *testing.T) {

	app := NewHashServer(2 * time.Second)
	app.Shutdown()

	req, err := http.NewRequest("POST", "/hash", strings.NewReader("password=boogabooga"))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.Hash)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusGone {
		t.Errorf("Was expecting http.StatusGone, received: %v", rr.Code)
	}
}

func TestGetHashAfterShutdown(t *testing.T) {

	hs := NewHashServer(1 * time.Second)
	postToHash(hs, "password=superSecret1", t)
	hs.Shutdown()

	req, err := http.NewRequest("GET", "/hash/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(hs.GetHash)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusGone {
		t.Errorf("Was expecting http.StatusGone, received: %v", rr.Code)
	}
}

func TestShutdownInterruptPendingJobs(t *testing.T) {
	//how do i actually test this?

	hs := NewHashServer(5 * time.Second)
	postToHash(hs, "password=superSecret1", t)
	postToHash(hs, "password=superSecret2", t)

	// set up a 5 second timer:
	// we should be done almost immediately..
	done := make(chan bool)
	timeout := make(chan bool)

	go func() {
		time.Sleep(hs.delay)
		timeout <- true
	}()

	go func() {
		hs.Shutdown()
		done <- true
	}()

	select {
	case <-timeout:
		t.Error("unexpected timeout")
	case <-done:
	}

}

func TestGetStats(t *testing.T) {

	app := NewHashServer(5 * time.Second)
	postToHash(app, "password=superSecret1", t)
	postToHash(app, "password=superSecret2", t)
	postToHash(app, "password=superSecret3", t)

	req, err := http.NewRequest("GET", "/stats", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.Stats)

	handler.ServeHTTP(rr, req)

	tsr := &utils.TimeStatsReport{}
	err = json.Unmarshal(rr.Body.Bytes(), tsr)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

}
