package service

import (
	"net/url"
	"strings"
	"sync"
	"time"
)

type State struct {
	url     string
	visited time.Time
}

type VisitorService struct {
	urls []string

	sync.RWMutex

	prefix  string
	host    string
	visited map[string]State
}

func NewVisitTrackerService() *VisitorService {
	return &VisitorService{
		urls:    make([]string, 0),
		visited: make(map[string]State),
	}
}

func (v *VisitorService) IsVisited(u string) bool {
	v.Lock()
	defer v.Unlock()

	_, ok := v.visited[u]
	return ok
}

func (v *VisitorService) SetUrlFilterPrefix(u string) {
	v.prefix = u
}

func (v *VisitorService) AddUrl(u string) {
	v.urls = append(v.urls, u)
}

func (v *VisitorService) DequeUrl() <-chan string {
	out := make(chan string)
	go func() {
		for {
			time.Sleep(1000 * time.Millisecond)

			if len(v.urls) > 0 {
				u := v.urls[0]
				v.urls = v.urls[1:]
				out <- u
			} else {
				close(out)
				return
			}
		}
	}()
	return out
}

func (v *VisitorService) FilterNewUrls(uu []string, sourceDoc string) []string {
	v.Lock()
	defer v.Unlock()

	newUrls := make([]string, 0)
	for _, u := range uu {
		_, err := url.Parse(u)
		if err != nil {
			continue
		}

		_, ok := v.visited[u]
		if !ok && strings.HasPrefix(u, v.prefix) {
			v.visited[sourceDoc] = State{
				url:     sourceDoc,
				visited: time.Now(),
			}
			v.urls = append(v.urls, u)
			newUrls = append(newUrls, u)
		}
	}
	return newUrls
}
