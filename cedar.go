package main

import (
	"encoding/xml"
	"io"
	"net/http"
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

func main() {}
