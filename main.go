package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	_bttvPageSize = 100
	_ulimit       = 500 // max number of open file descriptors

	_alphabet = "abcdefghijklmnopqrstuvwxyz0123456789'"
)

type Semaphore chan struct{}

func newSemaphore(size uint) Semaphore {
	res := make(Semaphore, _ulimit)
	for range [_ulimit]struct{}{} {
		res <- struct{}{}
	}
	return res
}

func (self Semaphore) acquire() {
	<-self
}

func (self Semaphore) release() {
	self <- struct{}{}
}

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

func doRequest(semaphore Semaphore, query string, offset uint) (*http.Response, []byte, error) {
	semaphore.acquire()
	defer semaphore.release()

	request, err := http.NewRequest(http.MethodGet, "https://api.betterttv.net/3/emotes/shared/search", nil)
	if err != nil {
		return nil, nil, err
	}

	q := request.URL.Query()
	q.Set("query", query)
	q.Set("offset", strconv.FormatUint(uint64(offset), 10))
	q.Set("limit", strconv.Itoa(_bttvPageSize))
	request.URL.RawQuery = q.Encode()

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, nil, err
	}

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
			if len(v) < _bttvPageSize {
				break
			}
			offset += _bttvPageSize
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
	var wg sync.WaitGroup
	semaphore := newSemaphore(_ulimit)
	wg.Add(len(_alphabet) * len(_alphabet) * len(_alphabet))
	for i := range _alphabet {
		for j := range _alphabet {
			for k := range _alphabet {
				query := _alphabet[i:i+1] + _alphabet[j:j+1] + _alphabet[k:k+1]
				go func() {
					defer wg.Done()
					find_emotes(query, semaphore)
				}()
			}
		}
	}
	wg.Wait()
}
