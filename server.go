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
var policy *bluemonday.Policy

func initPolicy() {
	policy = bluemonday.UGCPolicy()
}

func parsePoast(r io.Reader) (poast, error) {
	var rawPoast poast
	parseError := json.NewDecoder(r).Decode(&rawPoast)

	// sanitize input. TODO: see if there's a way to sanitize during json decoding instead of
	// creating two poasts.
	if parseError == nil {
		sanitizedPoast := poast{Username: policy.Sanitize(rawPoast.Username), Poast: policy.Sanitize(rawPoast.Poast), Poasted: time.Now()}
		return sanitizedPoast, nil
	}

	return rawPoast, parseError
}

var jsonCache []byte

func writePoasts(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	if len(jsonCache) == 0 {
		println("JSON cache is empty, refreshing")
		marshald, err := json.Marshal(poasts)
		if err != nil {
			writeError(1, "Unable to marshal json: "+err.Error(), w)
			return
		}

		jsonCache = marshald
	}

	w.Write(jsonCache)
}

func writeError(code int, error string, w http.ResponseWriter) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(error)
}

func main() {
	initPolicy()
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

				// clear json cache
				jsonCache = []byte{}
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
