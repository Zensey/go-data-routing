package jobs

import (
	"errors"
	"fmt"
	"net/http"

	res "github.com/Zensey/go-data-routing/examples/crawler/res"
	"github.com/Zensey/go-data-routing/examples/crawler/util"
	"golang.org/x/net/html"
)

type HtmlDocChecker struct {
	Url   string
	Hrefs []string

	*res.Resources
}

func (r *HtmlDocChecker) Run() {
	r.Hrefs = make([]string, 0)
	err := HtmlDocLinks(r.Url, &r.Hrefs, r.Client)

	if err != nil {
		fmt.Println("Runnable err:", err)
	}
}

func HtmlDocLinks(baseUrl string, urls *[]string, client *http.Client) error {
	res, err := client.Get(baseUrl)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		err = errors.New("bad status")
		return err
	}
	b := res.Body
	defer b.Close()

	z := html.NewTokenizer(b)
	for {
		tokenType := z.Next()
		switch tokenType {

		case html.ErrorToken:
			return errors.New("html.ErrorToken")

		case html.StartTagToken:
			t := z.Token()
			href := util.GetHrefFromToken(t)
			if t.Data == "a" && href != "" {
				url, err := util.GetFullUrl(baseUrl, href)
				if err != nil {
					continue
				}
				if url != "" {
					*urls = append(*urls, url)
				}
			}

		case html.TextToken:
		case html.EndTagToken:
		}
	}

	return nil
}
