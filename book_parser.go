package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/chromedp/chromedp"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.54 Safari/537.36"

// curently supported websites
type BookSource string

const (
	Amazon    BookSource = "Amazon"
	Goodreads BookSource = "Goodreads"
)

type Book struct {
	Title   string  // done
	Author  string  // done
	Rating  float64 // done
	Reviews int     // done

	ImgURL     string     // done
	BookURL    string     // done
	ReviewsURL string     // done
	Source     BookSource // done
}

type BookHTML struct {
	Root   *html.Node
	Source BookSource
}

type Chrome struct {
	Tabs   []context.Context
	Cancel []context.CancelFunc
	Cookie http.Cookie
}

var chrome *Chrome

func SpawnChrome() error {
	chrome = new(Chrome)

	o := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(userAgent),
		func(a *chromedp.ExecAllocator) { chromedp.Headless(a) },
		func(a *chromedp.ExecAllocator) { chromedp.NoSandbox(a) },
	)
	// browser setup
	browser, cancel := chromedp.NewExecAllocator(context.Background(), o...)
	chrome.Cancel = append(chrome.Cancel, cancel)

	// start a tab
	tab1, cancel := chromedp.NewContext(browser)
	chrome.Cancel = append(chrome.Cancel, cancel)

	if err := chromedp.Run(tab1); err != nil {
		return err
	}
	chrome.Tabs = append(chrome.Tabs, tab1)

	for i := 0; i < 5; i++ {
		tab, cancel := chromedp.NewContext(tab1)
		chrome.Tabs = append(chrome.Tabs, tab)
		chrome.Cancel = append(chrome.Cancel, cancel)
	}

	log.Println("Chrome instance successfully created")
	return nil
}

func CleanupChrome() {
	if chrome != nil {
		for i := len(chrome.Tabs) - 1; i >= 0; i-- {
			cancel := chrome.Cancel[i]
			cancel()
		}
	}
}

func GetBooks(query string) ([]*Book, error) {

	start := time.Now()
	amazonURLs, err := createAmazonURLs(query)
	if err != nil {
		return nil, errors.New("failed to create Amazon URLs")
	}
	goodreadsURLs, err := createGoodreadsURLs(query)
	if err != nil {
		return nil, errors.New("failed to create Goodreads URLs")
	}
	URLs := append(amazonURLs, goodreadsURLs...)
	fmt.Println("Creating URLs took:", time.Since(start))

	start = time.Now()
	HTMLs, err := getChromeParsedHTMLs(URLs)
	if err != nil {
		return nil, errors.New("something bad happened during parsing")
	}
	fmt.Println("Total chromedp time:", time.Since(start))

	var books []*Book
	start = time.Now()
	for _, HTML := range HTMLs {
		var items []*Book

		if HTML.Source == Amazon {
			items, err = FindAmazonBooks(HTML.Root)
			if err != nil {
				continue
			}
		}

		if HTML.Source == Goodreads {
			items, err = FindGoodreadsBooks(HTML.Root)
			if err != nil {
				continue
			}
		}

		books = append(books, items...)
	}
	fmt.Println("Parsing HTMLs by go took:", time.Since(start))

	return books, nil
}

// Creates new Chrome instance - must be invoked once per search query
func getChromeParsedHTMLs(urls []string) ([]*BookHTML, error) {
	if chrome == nil {
		SpawnChrome()
	}

	var parsedPages []*BookHTML
	chParsedHTML := make(chan *BookHTML)
	chIsFinished := make(chan bool)

	for i, url := range urls {
		go chromeFetchAndParse(url, chrome.Tabs[i], chParsedHTML, chIsFinished)
	}

	for i := 0; i < len(urls); {
		select {
		case root := <-chParsedHTML:
			parsedPages = append(parsedPages, root)
		case <-chIsFinished:
			i++
		}
	}

	if len(parsedPages) < 1 {
		return nil, errors.New("no roots where gathered")
	} else {
		return parsedPages, nil
	}
}

func chromeFetchAndParse(page string, tab context.Context, chParsedHTML chan *BookHTML, chIsFinished chan bool) {
	fmt.Println(page)
	var HTML string
	if err := chromedp.Run(tab,
		chromedp.Navigate(page),
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

	u, err := url.Parse(page)
	if err != nil {
		log.Println("Something bad happened during url parsing")
		chParsedHTML <- nil
		return
	}

	var source BookSource
	switch u.Host {
	case "www.amazon.com":
		source = Amazon
	case "www.goodreads.com":
		source = Goodreads
	default:
		log.Println("Unknown website for parsing")
		chParsedHTML <- nil
		return
	}

	chParsedHTML <- &BookHTML{root, source}
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
