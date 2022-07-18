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

type Semaphore chan struct{}

// func emote_gifs_urls_by_id(id string) (string, string, string) {
// 	return fmt.Sprintf("https://cdn.betterttv.net/emote/{id}/1x"),
// 		fmt.Sprintf("https://cdn.betterttv.net/emote/{id}/2x"),
// 		fmt.Sprintf("https://cdn.betterttv.net/emote/{id}/3x")
// }

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

// TODO: remove sync.Map, return maps, then gather them
func find_emotes(query string, mp *sync.Map, semaphore Semaphore) {
	offset := uint(0)
	for {
		response, data, err := doRequest(semaphore, query, offset)
		if err != nil {
			// TODO: fix too fast gathering
			log.Println(query, err)
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
				lst, ok := mp.Load(x.Code)
				if !ok {
					lst = any(make([]string, 0))
				}
				mp.Store(x.Code, append(lst.([]string), x.ID))
			}
			if len(v) < bttvPageSize {
				break
			}
			offset += bttvPageSize
		} else if response.StatusCode == http.StatusTooManyRequests {
			time.Sleep(time.Second * 2)
		} else {
			log.Println(response)
		}
	}
}

func main() {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789'"
	var (
		mp sync.Map
		wg sync.WaitGroup
	)
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
					find_emotes(query, &mp, semaphore)
				}()
			}
		}
	}
	wg.Wait()

	mp.Range(func(k, v any) bool {
		fmt.Println(k, strings.Join(v.([]string), " "))
		return true
	})
}
