package util

import (
	"testing"
)

func Test_UrlParser(t *testing.T) {
	docUrl := "https://aaa.b:80/a/b/c?a=1#aa"

	u, _ := GetFullUrl(docUrl, "http://aaa.b:80/d/e/f/j?a=1#aaa")
	if u != "http://aaa.b:80/d/e/f/j?a=1" {
		t.Fail()
	}

	u, _ = GetFullUrl(docUrl, "//bbb.b:9999/d/e/f/j?b=1#aaa")
	if u != "https://bbb.b:9999/d/e/f/j?b=1" {
		t.Fail()
	}

	u, _ = GetFullUrl(docUrl, "/d/e/f/j?a=1#aaa")
	if u != "https://aaa.b:80/d/e/f/j?a=1" {
		t.Fail()
	}

	u, _ = GetFullUrl(docUrl, "admin/")
	if u != "https://aaa.b:80/admin/" {
		t.Fail()
	}

}
