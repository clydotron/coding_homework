# jumpcloud
JumpCloud homework assignment: Simple HTTP server for generating base64 encoded Sha512 hashes. Includes statistics and shutdown functionality. Written in Go.

## Usage
Make local copy of the repo, open a terminal, cd into the root folder (jumpcloud)

To start the server:
```go run .```

additional flags: 
  * --port _integer_ port to listen on, **default is 8080**
  * --delay _integer_ delay (in seconds) before generating the sha512 hash, **default is 5.**

`go run . --port 4000 --delay 3` Listens on port 4000, 3 second delay between the hash POST request and the hash being available.

## Requests:
#### hash
  `POST /hash` 
  
  * GET /hash/:id
  * GET /stats
  * GET /shutdown
