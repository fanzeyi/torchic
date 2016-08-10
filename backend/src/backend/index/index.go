package index

import (
	"backend/crawler"
	"backend/mysql"
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
	TermPrefix        = "term"
	UrlPrefix         = "url"
	CountPrefix       = "count"
	TermSetKey        = "terms"
	TotalDocumentsKey = "total_documents"
)

var (
	WeightedElements = map[string]uint{
		"title": 4,
		"h1":    4,
		"h2":    3,
		"h3":    2,
	}
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

	// process body first
	words := i.processText(doc.ExtractText())

	if id, ok := i.saveDocument(job.Link, job.Document, words); ok {
		i.indexText(id, words, 1)

		// index title
		//for tag, weight := range WeightedElements {
		//elements := job.Document.Find(tag)

		//if len(elements.Nodes) == 0 {
		//continue
		//}

		//words = i.processText(elements.Text())
		//i.indexText(id, words, weight)
		//}
	}

	i.finishProcess()
}

func (i Indexer) finishProcess() {
	c := redis.GetConn()
	defer redis.ReturnConn(c)

	c.Do("INCR", TotalDocumentsKey)
}

func (i Indexer) processText(text string) []string {
	words := i.segment(text)
	processed := make([]string, 0)

	for _, word := range words {
		if !validateWord(word) {
			continue
		}

		// ISSUE: triming point would make "U.S." became "U."
		word = strings.TrimFunc(word, func(c rune) bool {
			return unicode.IsPunct(c)
		})

		processed = append(processed, word)
	}

	return processed
}

func (i Indexer) stemWords(words []string) []string {
	stemmed := make([]string, 0)

	for _, word := range words {
		stemmed = append(stemmed, english.Stem(word, true))
	}

	return stemmed
}

func (i Indexer) indexText(id int64, words []string, weight uint) {
	if len(words) > 0 {
		i.saveToRedis(id, i.stemWords(words), weight)
	}
}

func (i Indexer) segment(text string) []string {
	return strings.FieldsFunc(text, func(c rune) bool {
		// TODO: need more consideration on using '-' as a separator
		return unicode.IsSpace(c) || c == '/' || c == '-' || c == ':'
	})
}

func (i Indexer) saveDocument(link *crawler.URLContext, doc *goquery.Document, words []string) (id int64, ok bool) {
	db := mysql.GetConn()

	html, err := goquery.OuterHtml(doc.AndSelf())

	if err != nil {
		glog.Errorf("Error while getting html from document: %s", err)
	}

	res, err := db.Exec("INSERT INTO urls (hash, url, title, html, text) VALUES(?, ?, ?, ?, ?)", link.Hash(), link.URL().String(), doc.Find("title").Text(), html, strings.Join(words, " "))

	if err != nil {
		glog.Errorf("Error while inserting into MySQL: %s", err)
		return
	}

	id, err = res.LastInsertId()

	if err != nil {
		glog.Errorf("Error while getting last insert ID: %s", err)
		return
	}

	ok = true

	return
}

func (i Indexer) saveToRedis(id int64, words []string, weight uint) {
	var count map[string]uint

	count = make(map[string]uint)

	for _, word := range words {
		count[word] += 1
	}

	c := redis.GetConn()
	defer redis.ReturnConn(c)

	c.Send("MULTI")

	sadd := make([]interface{}, 0)
	sadd = append(sadd, redis.BuildKey(UrlPrefix, "%d", id))

	for word, num := range count {
		c.Send("ZADD", redis.BuildKey(TermPrefix, "%s", word), num*weight, id)
		sadd = append(sadd, word)
	}

	if len(sadd) > 1 {
		c.Send("SADD", sadd...)

		sadd[0] = TermSetKey
		c.Send("SADD", sadd...)
	}

	c.Send("INCRBY", redis.BuildKey(CountPrefix, "%d", id), len(count))
	c.Send("INCRBY", "total_words", len(count))

	_, err := c.Do("EXEC")

	if err != nil {
		glog.Errorf("%s %d", err, len(count))
	}
}
