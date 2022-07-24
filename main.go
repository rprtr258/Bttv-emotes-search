package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	bttvPageSize = 100
	ulimit       = 500 // max number of open file descriptors
)

type HTTPResponse struct {
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	Args    map[string][]string `json:"args"`
	Data    string              `json:"data"`
	Headers map[string]string   `json:"headers"`
}

type requestError struct {
	Query        string       `json:"query"`
	HTTPResponse HTTPResponse `json:"http_response"`
	Error        string       `json:"error"`
}

type Semaphore chan struct{}

func emote_query_url(query string, offset uint) string {
	return fmt.Sprintf("https://api.betterttv.net/3/emotes/shared/search?query=%s&offset=%d&limit=%d", query, offset, bttvPageSize)
}

func doRequest(semaphore Semaphore, query string, offset uint) (*http.Response, []byte, error) {
	<-semaphore // acquire
	defer func() {
		semaphore <- struct{}{} // release
	}()
	response, err := http.Get(emote_query_url(query, offset))
	if err != nil {
		return nil, nil, err
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, nil, err
	}
	response.Body.Close()
	return response, data, nil
}

func safeJSONMarshal[A any](v A) string {
	res, err := json.Marshal(v)
	if err != nil {
		// TODO: escape json string
		panic(fmt.Sprintf("error marshaling %+v: %s", v, err.Error()))
	}
	return string(res)
}

func find_emotes(query string, semaphore Semaphore) {
	offset := uint(0)
	for {
		response, data, err := doRequest(semaphore, query, offset)
		if err != nil {
			// TODO: fix too fast gathering
			headers := make(map[string]string, len(response.Header))
			for header, values := range response.Header {
				headers[header] = strings.Join(values, "\n")
			}
			errJSON := safeJSONMarshal(requestError{
				Error: err.Error(),
				Query: query,
				HTTPResponse: HTTPResponse{
					Method:  response.Request.Method,
					Path:    response.Request.URL.RawPath,
					Args:    response.Request.Form,
					Headers: headers,
				},
			})
			log.Println(errJSON)
			time.Sleep(time.Second * 2)
		}

		if response.Header.Get("content-type") == "application/json; charset=utf-8" {
			v := []struct {
				Code      string `json:"code"`
				ID        string `json:"id"`
				ImageType string `json:"imageType"`
				User      struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					DisplayName string `json:"displayName"`
					ProviderID  string `json:"providerId"`
				} `json:"user"`
			}{}
			json.Unmarshal(data, &v)
			for _, x := range v {
				out := safeJSONMarshal(x)
				fmt.Println(out)
			}
			if len(v) < bttvPageSize {
				break
			}
			offset += bttvPageSize
		} else if response.StatusCode == http.StatusTooManyRequests {
			time.Sleep(time.Second * 2)
		} else {
			headers := make(map[string]string, len(response.Header))
			for header, values := range response.Header {
				headers[header] = strings.Join(values, "\n")
			}
			errJSON := safeJSONMarshal(requestError{
				Error: "not json response",
				Query: query,
				HTTPResponse: HTTPResponse{
					Method:  response.Request.Method,
					Path:    response.Request.URL.RawPath,
					Args:    response.Request.Form,
					Headers: headers,
				},
			})
			// TODO: replace log with print to stderr
			log.Println(errJSON)
		}
	}
}

func main() {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789'"
	var wg sync.WaitGroup
	semaphore := make(Semaphore, ulimit)
	for range [ulimit]struct{}{} {
		semaphore <- struct{}{}
	}
	wg.Add(len(alphabet) * len(alphabet) * len(alphabet))
	for i := range alphabet {
		for j := range alphabet {
			for k := range alphabet {
				query := alphabet[i:i+1] + alphabet[j:j+1] + alphabet[k:k+1]
				go func() {
					defer wg.Done()
					find_emotes(query, semaphore)
				}()
			}
		}
	}
	wg.Wait()
}
