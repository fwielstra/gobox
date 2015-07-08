package main

import (
	"encoding/json"
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"io"
	"log"
	"math"
	"net/http"
	"time"
)

type poast struct {
	Username string
	Poasted  time.Time
	Poast    string
}

var poasts []poast

func parsePoast(r io.Reader) (poast, error) {
	var rawPoast poast
	parseError := json.NewDecoder(r).Decode(&rawPoast)

	// sanitize input. TODO: see if there's a way to sanitize during json decoding instead of
	// creating two poasts.
	if parseError == nil {
		sanitizer := bluemonday.UGCPolicy()
		sanitizedPoast := poast{Username: sanitizer.Sanitize(rawPoast.Username), Poast: sanitizer.Sanitize(rawPoast.Poast), Poasted: time.Now()}
		return sanitizedPoast, nil
	}

	return rawPoast, parseError
}

func writePoasts(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(poasts)
}

func writeError(code int, error string, w http.ResponseWriter) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(error)
}

func main() {
	poasts = []poast{}

	http.HandleFunc("/poast", func(w http.ResponseWriter, r *http.Request) {

		if r.Method == "GET" {
			writePoasts(w)
		}

		if r.Method == "POST" {
			newPoast, error := parsePoast(r.Body)
			if error != nil {
				fmt.Printf("Error parsing poast: " + error.Error())
				writeError(http.StatusBadRequest, "Unable to parse JSON: "+error.Error(), w)
			} else {
				fmt.Printf("new poast: %+v\n", newPoast)
				newPoasts := []poast{newPoast}
				poasts = append(newPoasts, poasts...)
				// save max 100 messages. Arbitrary limit.
				max := int(math.Min(float64(len(poasts)), 100))
				poasts = poasts[:max]
				writePoasts(w)
			}
		}
	})

	// serve client.html file
	http.HandleFunc("/client.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/client.html")
	})

	hoast := "localhost:8080"
	log.Print("Starting gobox server at " + hoast)
	log.Fatal(http.ListenAndServe(hoast, nil))
}
