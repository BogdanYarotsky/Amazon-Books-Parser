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

	// browser setup
	browser, cancel := chromedp.NewExecAllocator(context.Background(), o...)
	defer cancel()

	// start a tab
	tab1, cancel := chromedp.NewContext(browser)
	defer cancel()
	if err := chromedp.Run(tab1); err != nil {
		log.Fatal(err)
	}
	var tabs []context.Context
	tabs = append(tabs, tab1)

	// start next tabs
	for i := 0; i < len(urls)-1; i++ {
		tab, cancel := chromedp.NewContext(tab1)
		tabs = append(tabs, tab)
		defer cancel()
	}

	var roots []*html.Node
	chParsedHTML := make(chan *html.Node)
	chIsFinished := make(chan bool)

	for i, url := range urls {
		go FetchAndParse(url, tabs[i], chParsedHTML, chIsFinished)
	}

	for i := 0; i < len(urls); {
		select {
		case root := <-chParsedHTML:
			roots = append(roots, root)
		case <-chIsFinished:
			i++
		}
	}

	if len(roots) < 1 {
		return nil, errors.New("no roots where gathered")
	} else {
		return roots, nil
	}
}

func FetchAndParse(url string, tab context.Context, chParsedHTML chan *html.Node, chIsFinished chan bool) {
	fmt.Println(url)

	var HTML string
	if err := chromedp.Run(tab,
		chromedp.Navigate(url),
		chromedp.OuterHTML("html", &HTML, chromedp.ByQuery),
	); err != nil {
		log.Fatal(err)
	}

	defer func() {
		chIsFinished <- true
	}()

	// uncomment to check html response
	//fmt.Print(HTML)

	resp := strings.NewReader(HTML)
	root, err := html.Parse(resp)
	if err != nil {
		log.Println("Something bad happened during page parsing")
		chParsedHTML <- nil
		return
	}

	chParsedHTML <- root
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
