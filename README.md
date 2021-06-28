# jumpcloud

## Usage
Make local copy of the repo, CD into the root folder (jumpcloud)

To start the server:
```go run .```

additional flags: 
  * --port <int> port to listen on, default is 8080
  * --delay <int> delay (in seconds) before generating the sha512 hash, default is 5.

## Requests:
  * POST /hash
  * GET /hash/:id
  * GET /stats
  * GET /shutdown
