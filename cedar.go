package main

import (
	"io"
	"net/http"
)

func newsFeed() (string, error) {
	resp, err := http.Get("https://fsf.org.in/news/feed.atom")
	if err != nil {
		return "", err
	}

	// Init vars.
	chunk := make([]byte, 100)
	feed := make([]byte, 0)

	// Read feed.
	for {
		c, err := resp.Body.Read(chunk)
		if c < 1 {
			break
		}
		if err != nil && err != io.EOF {
			return "", err
		}
		feed = append(feed, chunk[0:c]...)
	}
	return string(feed), nil
}

func main() {}
