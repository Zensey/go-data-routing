package util

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func GetFullUrl(docUrl, href string) (string, error) {
	h, err := url.Parse(href)
	if err != nil {
		return "", err
	}
	h.Fragment = ""

	if h.IsAbs() {
		return h.String(), nil
	}
	doc, err := url.Parse(docUrl)
	if err != nil {
		return "", err
	}

	doc.Fragment = ""
	doc.RawQuery = ""
	doc.Path = ""

	if strings.HasPrefix(href, "//") {
		doc.Host = ""
		return doc.String() + h.String(), nil
	}

	if !strings.HasPrefix(href, "/") {
		h.Path = "/" + h.Path
	}
	return doc.String() + h.RequestURI(), nil
}

func GetHrefFromToken(t html.Token) string {
	for _, a := range t.Attr {
		if a.Key == "href" {
			return a.Val
		}
	}
	return ""
}

func SplitFragment(s string) string {
	i := strings.Index(s, "#")
	if i < 0 {
		return s
	}
	return s[:i]
}
