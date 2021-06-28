package main

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
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

	app := NewHashServer(2*time.Second, context.Background())

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
	app := NewHashServer(2*time.Second, context.Background())

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
	app := NewHashServer(2*time.Second, context.Background())

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

	app := NewHashServer(2*time.Second, context.Background())
	app.hashCount = 10

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

// helper function
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

func TestGetHashNoID(t *testing.T) {
	app := NewHashServer(2*time.Second, context.Background())

	req, err := http.NewRequest("GET", "/hash/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.GetHash)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Was expecting http.StatusBadRequest, received: %v", rr.Code)
	}
}
func TestGetHashNonIntegerID(t *testing.T) {
	app := NewHashServer(2*time.Second, context.Background())

	req, err := http.NewRequest("GET", "/hash/bad", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.GetHash)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Was expecting http.StatusBadRequest, received: %v", rr.Code)
	}
}
func TestGetHashTooEarly(t *testing.T) {
	app := NewHashServer(2*time.Second, context.Background())
	app.hashCount = 10

	h1 := postToHash(app, "password=secret!", t)

	req, err := http.NewRequest("GET", fmt.Sprintf("/hash/%d", h1), nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.GetHash)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Was expecting http.StatusNotFound, received: %v", rr.Code)
	}
}

func TestGetHashSuccess(t *testing.T) {
	hs := NewHashServer(1*time.Second, context.Background())
	hs.hashCount = 10

	// send the post request:
	postToHash(hs, "password=secret!", t)

	// sleep for a little over 1 second
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

func getHash(hs *HashServer, id int, t *testing.T) string {

	req, err := http.NewRequest("GET", fmt.Sprintf("/hash/%d", id), nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(hs.GetHash)

	handler.ServeHTTP(rr, req)

	if rr.Code == http.StatusOK {
		return rr.Body.String()
	}

	return ""
}

func TestCancelInterruptPendingJobs(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	hs := NewHashServer(100*time.Microsecond, ctx)
	h1 := postToHash(hs, "password=superSecret1", t)
	h2 := postToHash(hs, "password=superSecret2", t)

	cancel()

	time.Sleep(hs.delay * 2)

	if hash1 := getHash(hs, h1, t); hash1 != "" {
		t.Errorf("was expecting an empty string: received[%v]", hash1)
	}

	if hash2 := getHash(hs, h2, t); hash2 != "" {
		t.Errorf("was expecting an empty string: received[%v]", hash2)
	}
}

func TestGetStats(t *testing.T) {

	app := NewHashServer(5*time.Second, context.Background())
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
