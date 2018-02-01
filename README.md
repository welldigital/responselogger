# responselogger

Server middleware to log HTTP server response status codes and their times.

## Usage

```go
http.
h := responselogger.NewHandler()
```

## Output

### Example output from JSON logging

```json
{ "time": "2000-01-02T03:04:05Z", "src": "rl", "status": 200, "http_2xx": 1, "len": 454, "ms": 200, "path": "/test" }
{ "time": "2000-01-02T03:04:05Z", "src": "rl", "status": 300, "http_3xx": 1, "len": 656, "ms": 200, "path": "/test" }
{ "time": "2000-01-02T03:04:05Z", "src": "rl", "status": 404, "http_4xx": 1, "len": 757, "ms": 200, "path": "/test" }
```