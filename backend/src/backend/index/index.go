package index

import (
	"backend/crawler"
	"backend/redis"
	"backend/utils"
	"bytes"
	"strings"
	"unicode"

	"golang.org/x/net/html"

	"github.com/PuerkitoBio/goquery"
	"github.com/golang/glog"
	"github.com/kljensen/snowball/english"
)

const (
	TermPrefix = "term"
	UrlPrefix  = "url"
)

type Indexer struct {
	incoming *utils.PopChannel
}

type Document goquery.Document

// Get the specified node's text content.
// https://github.com/PuerkitoBio/goquery/blob/master/property.go#L195
func getNodeText(node *html.Node) string {
	if node.Type == html.TextNode {
		// Keep newlines and spaces, like jQuery
		return node.Data
	} else if node.Type == html.ElementNode && (node.Data == "script" || node.Data == "style" || node.Data == "noscript") {
		// Skip script and style tag
		return ""
	} else if node.FirstChild != nil {
		var buf bytes.Buffer
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			buf.WriteString(getNodeText(c) + " ")
		}
		return buf.String()
	}

	return ""
}

func (d Document) ExtractText() string {
	var buf bytes.Buffer

	// Slightly optimized vs calling Each: no single selection object created
	for _, n := range d.Nodes {
		buf.WriteString(getNodeText(n) + " ")
	}
	return buf.String()
}

func NewIndexer(incoming *utils.PopChannel) *Indexer {
	indexer := new(Indexer)
	indexer.incoming = incoming

	return indexer
}

func (i Indexer) Run() {
	for {
		select {
		case jobs := <-*i.incoming:
			i.processJobs(jobs)
		}
	}
}

func (i Indexer) processJobs(jobs []interface{}) {
	for _, job := range jobs {
		i.process(job.(*crawler.CrawlResponse))
	}
}

func validateWord(word string) bool {
	hasLatin := false

	for _, char := range word {
		if unicode.Is(unicode.Latin, char) {
			hasLatin = true
			continue
		} else if unicode.IsNumber(char) {
			continue
		} else if unicode.IsPunct(char) {
			continue
		}
		return false
	}

	// stop words
	switch word {
	case "'s", "the":
		return false
	}

	return hasLatin
}

func (i Indexer) process(job *crawler.CrawlResponse) {
	doc := Document(*job.Document)
	words := i.segment(doc.ExtractText())
	processed := make([]string, 0)

	for _, word := range words {
		if !validateWord(word) {
			continue
		}

		// ISSUE: triming point would make "U.S." became "U."
		word = strings.TrimFunc(word, func(c rune) bool {
			switch c {
			case ',', '.', '"', ':', '(', ')', '?':
				return true
			}
			return false
		})

		processed = append(processed, english.Stem(word, true))
	}

	//glog.Infof("words: %v", doc.ExtractText())
	//glog.Infof("%v", processed)

	i.index(job.Link, processed)
}

func (i Indexer) segment(text string) []string {
	return strings.FieldsFunc(text, func(c rune) bool {
		// TODO: need more consideration on using '-' as a separator
		return unicode.IsSpace(c) || c == '/' || c == '-'
	})
}

func (i Indexer) index(link *crawler.URLContext, words []string) {
	var count map[string]uint

	count = make(map[string]uint)

	var total uint

	for _, word := range words {
		count[word] += 1
		total += 1
	}

	c := redis.GetConn()
	defer c.Close()

	c.Send("MULTI")
	for word, num := range count {
		//glog.Infof("key: %s, type: %s", redis.BuildKey(TermPrefix, "%s", word), reflect.TypeOf(word))
		c.Send("ZADD", redis.BuildKey(TermPrefix, "%s", word), float32(num)/float32(total), link.URL().String())
		c.Send("SADD", redis.BuildKey(UrlPrefix, "%s", link.URL().String()), word)
	}
	_, err := c.Do("EXEC")

	if err != nil {
		glog.Errorf("%s", err)
	}
}
