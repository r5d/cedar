package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"os"
	"path"
)

type Link struct {
	XMLName xml.Name `xml:"link"`
	Href    string   `xml:"href,attr"`
}

type Entry struct {
	XMLName xml.Name `xml:"entry"`
	Id      string   `xml:"id"`
	Title   string   `xml:"title"`
	Link    Link
}

type Feed struct {
	XMLName xml.Name `xml:"feed"`
	Entry   []Entry  `xml:"entry"`
}

type Ids []string

func newsFeed() ([]byte, error) {
	// Init feed.
	feed := make([]byte, 0)

	resp, err := http.Get("https://fsf.org.in/news/feed.atom")
	if err != nil {
		return feed, err
	}

	// Read feed.
	chunk := make([]byte, 100)
	for {
		c, err := resp.Body.Read(chunk)
		if c < 1 {
			break
		}
		if err != nil && err != io.EOF {
			return feed, err
		}
		feed = append(feed, chunk[0:c]...)
	}
	return feed, nil
}

func parseFeed(feed []byte) (Feed, error) {
	f := Feed{}

	err := xml.Unmarshal(feed, &f)
	if err != nil {
		return f, err
	}

	return f, nil
}

func readFile(f *os.File) ([]byte, error) {
	bs, chunk := make([]byte, 0), make([]byte, 10)
	for {
		n, err := f.Read(chunk)
		if err != nil && err != io.EOF {
			return bs, err
		}
		bs = append(bs, chunk[0:n]...)

		if err == io.EOF {
			break
		}
	}
	return bs, nil
}

func cacheFor(section string) (Ids, error) {
	cache := make(Ids, 0)

	h, _ := os.UserHomeDir()
	d := path.Join(h, ".cedar")

	err := os.MkdirAll(d, 0700)
	if err != nil {
		return cache, err
	}

	f, err := os.Open(path.Join(d, section+".json"))
	if os.IsNotExist(err) {
		return cache, nil
	}

	bs, err := readFile(f)
	if err != nil {
		return cache, err
	}

	err = json.Unmarshal(bs, &cache)
	if err != nil {
		return cache, err
	}
	return cache, nil
}

func (cache *Ids) add(entry Entry) {
	n := len(*cache)

	// Expand cache
	c := make(Ids, n+1)
	copy(c, *cache)

	// Cache entry
	c[n] = entry.Id

	*cache = c
}

func (cache Ids) save() error {
	// Dummy for now.
	return nil
}

func main() {}
