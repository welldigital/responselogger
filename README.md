# responselogger

Server middleware to log HTTP server response status codes and their times.

## Usage

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/welldigital/responselogger"
)

func main() {
	// Create a mux to store routes.
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world!"))
	})

	mux.HandleFunc("/other", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	// Wrap the mux inside the responselogger.
	loggedHandler := responselogger.NewHandler(mux)

	// Start serving.
	fmt.Println("Listening on :1234...")
	http.ListenAndServe(":1234", loggedHandler)
	fmt.Println("Exited")
}
```

## Output

### Example output from JSON logging

```json
{"time":"2018-02-01T18:41:31Z","src":"rl","status":404,"http_4xx":1,"len":19,"ms":2,"path":"/other"}
{"time":"2018-02-01T18:41:39Z","src":"rl","status":200,"http_2xx":1,"len":12,"ms":4,"path":"/"}
```

## Configuring AWS CloudWatch metrics extraction with Terraform

The JSON logs can be converted into CloudWatch metrics for monitoring with the following Terraform configuration.

```hcl
resource "aws_cloudwatch_log_metric_filter" "http_2xx" {
  name = "HTTP 2xx status"
  pattern = "{ ($.src = \"rl\") && ($.http_2xx = *) }"
  log_group_name = "logs"

  metric_transformation {
    name = "http_2xx"
    namespace = "HTTPMetrics"
    value = "$.http_2xx"
  }
}

resource "aws_cloudwatch_log_metric_filter" "http_3xx" {
  name = "HTTP 3xx status"
  pattern = "{ ($.src = \"rl\") && ($.http_3xx = *) }"
  log_group_name = "logs"

  metric_transformation {
    name = "http_3xx"
    namespace = "HTTPMetrics"
    value = "$.http_3xx"
  }
}

resource "aws_cloudwatch_log_metric_filter" "http_4xx" {
  name = "HTTP 4xx status"
  pattern = "{ ($.src = \"rl\") && ($.http_4xx = *) }"
  log_group_name = "logs"

  metric_transformation {
    name = "http_4xx"
    namespace = "HTTPMetrics"
    value = "$.http_4xx"
  }
}

resource "aws_cloudwatch_log_metric_filter" "http_5xx" {
  name = "HTTP 5xx status"
  pattern = "{ ($.src = \"rl\") && ($.http_5xx = *) }"
  log_group_name = "logs"

  metric_transformation {
    name = "http_5xx"
    namespace = "HTTPMetrics"
    value = "$.http_5xx"
  }
}

resource "aws_cloudwatch_log_metric_filter" "http_duration_ms" {
  name = "HTTP duration ms"
  pattern = "{ ($.src = \"rl\") && ($.ms = *) }"
  log_group_name = "logs"

  metric_transformation {
    name = "http_duration_ms"
    namespace = "HTTPMetrics"
    value = "$.ms"
  }
}
```
