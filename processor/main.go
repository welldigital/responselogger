package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/welldigital/responselogger/processor/stats"

	"github.com/welldigital/responselogger/processor/urlpattern"
)

func main() {
	f, err := os.Open("logs.json")
	if err != nil {
		log.Fatalf("could not open logs.json: %v", err)
	}
	defer f.Close()
	r := bufio.NewScanner(f)
	r.Split(bufio.ScanLines)
	lines := []logLine{}
	var lineIndex int
	for r.Scan() {
		var ll logLine
		err := json.Unmarshal(r.Bytes(), &ll)
		if err != nil {
			log.Fatalf("error unmarshalling line: %v", err)
		}
		if ll.Src != "rl" {
			continue
		}
		lines = append(lines, ll)
		lineIndex++
	}
	if r.Err() != nil {
		log.Fatalf("failed to read data from logs: %v", r.Err())
	}

	w := csv.NewWriter(os.Stdout)
	w.Write([]string{"URL", "Count", "Sum", "Avg"})

	urlPatternToLines := map[string][]logLine{}
	for _, l := range lines {
		method := l.Method
		if method == "" {
			method = "HTTP"
		}
		pattern := fmt.Sprintf("%v %v", method, urlpattern.Extract(l.Path))
		urlPatternToLines[pattern] = append(urlPatternToLines[pattern], l)
	}
	for urlPattern, urlLines := range urlPatternToLines {
		responseTime := byMilliseconds(urlLines)
		count := fmt.Sprintf("%d", len(urlLines))
		sum := fmt.Sprintf("%d", stats.Sum(responseTime))
		avg := fmt.Sprintf("%.2f", stats.Average(responseTime))
		w.Write([]string{urlPattern, count, sum, avg})
	}
	w.Flush()
}

func byMilliseconds(lines []logLine) []stats.Value {
	op := make([]stats.Value, len(lines))
	for i := range lines {
		op[i] = stats.Value(logLineByMilliseconds(lines[i]))
	}
	return op
}

type logLineByMilliseconds logLine

func (ll logLineByMilliseconds) Value() int {
	return ll.Milliseconds
}

type logLine struct {
	Time         time.Time `json:"time"`
	Package      string    `json:"pkg"`
	Function     string    `json:"fn"`
	Src          string    `json:"src"`
	Level        string    `json:"level"`
	Status       int       `json:"status"`
	Length       int       `json:"len"`
	Milliseconds int       `json:"ms"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
}
