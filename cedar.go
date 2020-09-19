// SPDX-License-Identifier: ISC
// Copyright © 2020 rsiddharth <s@ricketyspace.net>

package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
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

const sendmail string = "/usr/sbin/sendmail"

var emailTo string

func init() {
	flag.StringVar(&emailTo, "t", "", "Email address for sending emails to")
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

func writeFile(f os.File, cache Ids) error {
	bs, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	n, err := f.Write(bs)
	if n != len(bs) {
		return err
	}
	return nil
}

func cacheFor(section string) (Ids, error) {
	cache := make(Ids, 0)

	h := os.Getenv("HOME")
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
	if len(bs) == 0 || err != nil {
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

func (cache Ids) save(section string) error {
	h := os.Getenv("HOME")
	d := path.Join(h, ".cedar")

	f, err := os.OpenFile(path.Join(d, section+".json"),
		os.O_CREATE|os.O_WRONLY,
		0600)
	if err != nil {
		return err
	}

	err = writeFile(*f, cache)
	if err != nil {
		return err
	}

	return nil
}

func (entry Entry) in(cache Ids) bool {
	for i := 0; i < len(cache); i++ {
		if entry.Id == cache[i] {
			return true
		}
	}
	return false
}

func (entry Entry) makeEmail(section string) string {
	return fmt.Sprintf(`To: %s
From: no-reply@gnu.org.in
Subject: FSF India - %s - %s

FSF India published "%s":

   %s
`,
		emailTo,
		strings.Title(section),
		entry.Title,
		entry.Title,
		entry.Link.Href)
}

func (entry Entry) email(section string) error {
	cmd := exec.Command(sendmail, "-t")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	io.WriteString(stdin, entry.makeEmail(section))
	stdin.Close()

	_, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Printf("Successfully sent %s to %s\n",
		entry.Id, emailTo)

	return nil
}

func processNews() error {
	newsXML, err := newsFeed()
	if err != nil {
		return err
	}

	news, err := parseFeed(newsXML)
	if err != nil {
		return err
	}

	cache, err := cacheFor("news")
	if err != nil {
		return err
	}

	for i := 0; i < len(news.Entry); i++ {
		if news.Entry[i].in(cache) {
			continue
		}

		err := news.Entry[i].email("news")
		if err != nil {
			return err
		}
		cache.add(news.Entry[i])
	}
	err = cache.save("news")
	if err != nil {
		return err
	}

	return nil
}

func main() {
	flag.Parse()

	// Quit if emailTo is not set.
	if flag.NFlag() != 1 {
		flag.PrintDefaults()
		return
	}

	err := processNews()
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
	}
}
