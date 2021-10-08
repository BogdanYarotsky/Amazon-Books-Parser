package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

type Book struct {
	Href  string
	Title string
}

func FindH2(url string) ([]Book, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	log.Printf("Got the http response from %q\n", url)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		return nil, fmt.Errorf("getting %s: %s", url, resp.Status)
	}

	root, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var links []Book
	var rec func(*html.Node)
	rec = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h2" {

			var text string
			if n.FirstChild != nil {
				text = grabText(n.FirstChild)
			}
			links = append(links, Book{"h2", text})

		}
		if n.FirstChild != nil {
			rec(n.FirstChild)
		}
		if n.NextSibling != nil {
			rec(n.NextSibling)
		}
	}
	rec(root)

	return links, nil

}

func grabText(n *html.Node) string {
	var sb strings.Builder
	var rec func(*html.Node)
	rec = func(n *html.Node) {
		if n.Type == html.TextNode {
			s := n.Data
			sb.WriteString(s)
		}
		if n.FirstChild != nil {
			rec(n.FirstChild)
		}
		if n.NextSibling != nil {
			rec(n.NextSibling)
		}
	}
	rec(n)

	return strings.Join(strings.Fields(sb.String()), " ")
}

// LinksString returns reasonable string listing links
func LinksString(links []Book) string {
	var maxW int
	for _, l := range links {
		if len(l.Href) > maxW {
			maxW = len(l.Href)
		}
	}
	maxW++

	var sb strings.Builder
	for _, l := range links {
		sb.WriteString(l.Href)
		for i := 0; i < maxW-len(l.Href); i++ {
			sb.WriteRune(' ')
		}
		sb.WriteString(l.Title)
		sb.WriteRune('\n')
	}

	return sb.String()
}
