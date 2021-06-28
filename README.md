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

Server can be shutdown either by via ctrl-c from inside the terminal window, or sending a GET request to /shutdown

## Requests:
#### hash
  `curl --data "password=secret" http://localhost:8080/hash`
  
  
  Code | Text | Message | Description
  --- | --- | --- | ---
  200|StatusOK|Hash ID|ID of the new hash: the nth request will return n.
  400|StatusBadRequest|Only POST accepted.|Non POST request received.
  400|StatusBadRequest|required key password missing|Required value 'password' not present.
  500|StatusInternalServerError|ParseForm failed|Unable to parse the form data: bad format or empty.
  
 #### hash/id  
  `curl http://localhost:8080/hash/id`
  
  
  Code | Text | Message | Description
  --- | --- | --- | ---
  200|StatusOK|Hash|Base64 encoded Sha512 hash of user supplied 'password'
  400|StatusBadRequest|Only GET accepted.|Non GET request received.
  400|StatusBadRequest|Must provide valid id|id missing or invalid.
  400|StatusBadRequest|ID must be a number|id not a number
   404|StatusNotFound|Not found|Hashed value not found
  
 #### stats
  `curl http://localhost:8080/stats`
  
  
  Code | Text | Message | Description
  --- | --- | --- | ---
  200|StatusOK|json object|json containing 2 fields
  500|StatusInternalServerError|Failed to JSON encode time statistics|Something went wrong while encoding the JSON
  
  returns json object with two fields:
  * total: total number of calls to POST /hash
  * average: average time, in microseconds, or each call
  
 #### shutdown
  `curl http://localhost:8080/shutdown`
  
  
  Code | Text | Message | Description
  --- | --- | --- | ---
  200|StatusOK|Shutting down.|Server is shutting down.
  
  
  Gracefully shutdown the server
