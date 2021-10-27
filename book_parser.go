package main

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"log"
	"strings"

	"github.com/chromedp/chromedp"
)

const (
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.54 Safari/537.36"
)

// Creates new Chrome instance - must be invoked once per search query
func getParsedHTMLs(urls []string) ([]*html.Node, error) {
	o := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(userAgent),
	)

	cx, cancel := chromedp.NewExecAllocator(context.Background(), o...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(cx)
	defer cancel()

	var roots []*html.Node
	for _, url := range urls {
		fmt.Println(url)
		var HTML string
		if err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.OuterHTML("html", &HTML, chromedp.ByQuery),
		); err != nil {
			log.Fatal(err)
		}
		// uncomment to check html response
		//fmt.Print(HTML)

		resp := strings.NewReader(HTML)
		root, err := html.Parse(resp)
		if err != nil {
			log.Println("Something bad happened during page parsing")
			return nil, err
		}
		roots = append(roots, root)
	}

	if len(roots) < 1 {
		return nil, errors.New("no roots where gathered")
	} else {
		return roots, nil
	}
}

func PrintBooks(books ...*Book) {
	for i, book := range books {
		fmt.Println()
		fmt.Println("#", i+1)
		//fmt.Println("Image:", book.ImgURL)
		fmt.Println("Title:", book.Title)
		//fmt.Println("Link:", book.BookURL)
		fmt.Println("Author:", book.Author)
		fmt.Printf("Average rating: %.1f\n", book.Rating)
		fmt.Println("Total reviews:", book.Reviews)
		//fmt.Println("Browse reviews:", book.ReviewsURL)
		//fmt.Println("Book source:", book.Source)
		fmt.Println("====")
	}
}

// helper function to get child text node (goes deep)
func getText(n *html.Node) string {
	var text string
	if n.Type == html.TextNode {
		return n.Data
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += getText(c)
	}

	return strings.Trim(text, "\n")
}

// helper function to get href of "a" element (checks 1 node)
func getHref(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				return attr.Val
			}
		}
	}
	return ""
}
