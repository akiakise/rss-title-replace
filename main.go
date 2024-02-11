package main

import (
	"bytes"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type RSS struct {
	Channel Channel `xml:"channel"`
}

// Channel is an RSS Channel
type Channel struct {
	Title       string `xml:"title,omitempty"`
	Link        string `xml:"link,omitempty"`
	Description string `xml:"description,omitempty"`
	Language    string `xml:"language,omitempty"`
	Copyright   string `xml:"copyright,omitempty"`
	PubDate     string `xml:"pubDate,omitempty"`
	Items       []Item `xml:"item,omitempty"`
}

// Item is an RSS Item
type Item struct {
	Title       string    `xml:"title,omitempty"`
	Link        string    `xml:"link,omitempty"`
	Description string    `xml:"description,omitempty"`
	Author      string    `xml:"author,omitempty"`
	Category    Category  `xml:"category,omitempty"`
	Enclosure   Enclosure `xml:"enclosure,omitempty"`
	PubDate     string    `xml:"pubDate,omitempty"`
}

// Category is category metadata for Feeds and Entries
type Category struct {
	Domain string `xml:"domain,attr"`
	Value  string `xml:",chardata"`
}

// Enclosure is a media object that is attached to the item
type Enclosure struct {
	URL    string `xml:"url,attr"`
	Length string `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

//goland:noinspection HttpUrlsUsage,GoUnreachableCode
func replaceTitle(w http.ResponseWriter, r *http.Request) {
	// parse origin url and replace "/https:/example.com" to "https://example.com" which was short by browser by default
	requestURI := r.RequestURI
	if !strings.HasPrefix(requestURI, "/http") {
		_, _ = w.Write([]byte("unsupported url: " + requestURI))
		return
	}
	var originUrl string
	if strings.HasPrefix(requestURI, "/https:/") {
		originUrl = strings.ReplaceAll(requestURI, "/https:/", "https://")
	} else {
		originUrl = strings.ReplaceAll(requestURI, "/http:/", "http://")
	}
	log.Println("origin url: ", originUrl)
	originUrlSplit := strings.Split(originUrl, "=")
	originQuery := originUrlSplit[len(originUrlSplit)-1]
	unescapeQuery, err := url.QueryUnescape(originQuery)
	if err != nil {
		_, _ = w.Write([]byte("unescape origin query url failed" + err.Error()))
		return
	}

	// request origin url
	resp, err := http.Get(originUrl)
	if err != nil {
		_, _ = w.Write([]byte("request origin url failed" + err.Error()))
		return
	}
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		_, _ = w.Write([]byte("read origin url response failed" + err.Error()))
		return
	}
	respString := string(respBytes)

	// response content check
	if !strings.HasPrefix(respString, "<rss") {
		_, _ = w.Write([]byte("origin url is not rss!"))
		return
	}

	// unmarshal to object, in order to parse title
	rss := RSS{}
	decoder := xml.NewDecoder(bytes.NewReader(respBytes))
	decoder.Strict = false
	err = decoder.Decode(&rss)
	if err != nil {
		_, _ = w.Write([]byte("unmarshal response failed" + err.Error()))
		return
	}

	// replace title
	trimTitle := strings.Trim(rss.Channel.Title, " \n")
	log.Println("replace title [", trimTitle, "] to [", unescapeQuery, "]")
	newRss := strings.Replace(respString, trimTitle, unescapeQuery, 1)

	_, _ = w.Write([]byte(newRss))
}

func main() {
	http.HandleFunc("/", replaceTitle)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalln("listen and serve error", err)
	}
}
