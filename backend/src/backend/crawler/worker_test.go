package crawler

import (
	"net/url"
	"testing"
)

func urlParse(raw string) (res *url.URL) {
	res, _ = url.Parse(raw)
	return
}

func TestBaseTagExpand(t *testing.T) {
	cases := []struct {
		url      *url.URL
		baseHref string
		aHref    string
		excepted string
	}{
		{urlParse("http://www.example.com"), "base/", "test.html", "http://www.example.com/base/test.html"},
		{urlParse("http://www.example.com/base"), "../", "test.html", "http://www.example.com/test.html"},
		{urlParse("http://www.example.com/base/"), "./subfolder/", "test.html", "http://www.example.com/base/subfolder/test.html"},
		{urlParse("http://www.example.com/base/"), "/subfolder/", "test.html", "http://www.example.com/subfolder/test.html"},
	}

	for _, test := range cases {
		if res := handleBaseTag(test.url, test.baseHref, test.aHref); res != test.excepted {
			t.Errorf("handleBaseTag(\"%s\", \"%s\", \"%s\"):\n\tExcepted: %s\n\tActual: %s", test.url, test.baseHref, test.aHref, test.excepted, res)
		}
	}
}
