package main

import (
	"io"
	"net/http"
)

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

func main() {}
