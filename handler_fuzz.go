// +build gofuzz

package responselogger

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

// Fuzz the JSONLogMessage.
func Fuzz(data []byte) int {
	u := &url.URL{
		Path: string(data),
	}
	j := JSONLogMessage(time.Now, u, http.StatusOK, 10, time.Millisecond*100)
	if isJSON(j) {
		return 1
	}
	return 0
}

func isJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}
