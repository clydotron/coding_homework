# jumpcloud
JumpCloud homework assignment: Simple HTTP server for generating base64 encoded Sha512 hashes. Includes statistics and shutdown functionality. Written in Go.

## Usage
Make local copy of the repo, open a terminal, cd into the root folder (jumpcloud)

To run unit tests:
`go test`

To start the server:
`go run .`

additional flags: 
  * --port _integer_ port to listen on, **default is 8080**
  * --delay _integer_ delay (in seconds) before generating the sha512 hash, **default is 5.**

`go run . --port 4000 --delay 3` 
Listens on port 4000, 3 second delay between the POST hash request and the hash being available via GET

Server can be shutdown either by hitting ctrl-c in the terminal window, or sending a GET request to /shutdown

## Requests:
#### hash
  `POST /hash` 
  `curl --data "password=secret" http://localhost:8080/hash`
  
  Code | Message | Description
  StatusOK|Hash ID|ID of the new hash.
  
 #### hash/id
  `GET /hash/:id`
  
 #### stats
  `GET /stats`
  
 #### shutdown
  `GET /shutdown`
