package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	_bttvPageSize = 100
	_workers      = 100 // aiohttp default connection pool size
	_retryDelay   = 10 * time.Second

	_apiURL = "https://api.betterttv.net/3/emotes/shared/search"

	_alphabet = "abcdefghijklmnopqrstuvwxyz0123456789'"
)

type emote struct {
	Code string `json:"code"`
	ID   string `json:"id"`
}

func emote_query_url(query string, offset int) string {
	q := url.Values{}
	q.Set("query", query)
	q.Set("offset", strconv.Itoa(offset))
	q.Set("limit", strconv.Itoa(_bttvPageSize))
	return _apiURL + "?" + q.Encode()
}

func get(client *http.Client, query string, offset int) (*http.Response, []byte, error) {
	response, err := client.Get(emote_query_url(query, offset))
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = response.Body.Close() }()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, err
	}
	return response, data, nil
}

// find_emotes collects emote ids for one query into `_<query>.json`,
// skipping the query if its file already exists. Any non-JSON response
// is retried forever after _retryDelay, like collect_raw.py.
func find_emotes(client *http.Client, query string) error {
	filename := "_" + query + ".json"
	if _, err := os.Stat(filename); err == nil {
		return nil
	}
	result := map[string][]string{}
	offset := 0
	for {
		response, data, err := get(client, query, offset)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s (offset %d): %v\n", query, offset, err)
			time.Sleep(_retryDelay)
			continue
		}
		if response.Header.Get("Content-Type") != "application/json; charset=utf-8" {
			fmt.Fprintf(os.Stderr, "%s (offset %d): not json response: %s\n", query, offset, response.Status)
			time.Sleep(_retryDelay)
			continue
		}
		var emotes []emote
		if err := json.Unmarshal(data, &emotes); err != nil {
			fmt.Fprintf(os.Stderr, "%s (offset %d): bad json: %v\n", query, offset, err)
			time.Sleep(_retryDelay)
			continue
		}
		for _, x := range emotes {
			result[x.Code] = append(result[x.Code], x.ID)
		}
		if len(emotes) < _bttvPageSize {
			break
		}
		offset += _bttvPageSize
	}
	out, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, out, 0o644)
}

func main() {
	total := len(_alphabet) * len(_alphabet) * len(_alphabet)
	var completed atomic.Int64

	queries := make(chan string)
	var wg sync.WaitGroup
	for range _workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{}
			for query := range queries {
				if err := find_emotes(client, query); err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
				done := completed.Add(1)
				fmt.Printf("%.2f%%\r", float64(done)/float64(total)*100)
			}
		}()
	}
	for i := range _alphabet {
		for j := range _alphabet {
			for k := range _alphabet {
				queries <- _alphabet[i:i+1] + _alphabet[j:j+1] + _alphabet[k:k+1]
			}
		}
	}
	close(queries)
	wg.Wait()
	fmt.Println()
}
