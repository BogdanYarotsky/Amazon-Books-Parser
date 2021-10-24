package main

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/chromedp/chromedp"
)

const (
	resultsClass = "s-main-slot s-result-list s-search-results sg-row"
	amazonURL    = "https://www.amazon.com"
	userAgent    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.54 Safari/537.36"
)

func FindAmazonBooks(query string) ([]*Book, error) {
	pages, err := createAmazonURLs(query)
	if err != nil {
		return nil, err
	}

	var booklist []*Book
	for _, page := range pages {
		root, err := getAmazonRoot(page)
		if err != nil {
			return nil, err
		}

		results, err := getAmazonResults(root)
		if err != nil {
			return nil, err
		}

		books, err := getAmazonBooks(results)
		if err != nil {
			return nil, err
		}

		booklist = append(booklist, books...)
	}

	return booklist, nil
}

func createAmazonURLs(query string) ([]string, error) {
	searchString := "/Books-Search/s?k=" + strings.ReplaceAll(query, " ", "+")
	selectTopRated := "&i=stripbooks&rh=n%3A283155%2Cp_72%3A1250221011&dc"
	var urls []string
	for i := 1; i <= 1; i++ {
		pageQuery := fmt.Sprintf("&page=%d&qid=1634582114&rnid=1250219011&ref=sr_pg_%d", i, i)
		urls = append(urls, amazonURL+searchString+selectTopRated+pageQuery)
	}
	return urls, nil
}

// All starts here
func getAmazonRoot(url string) (*html.Node, error) {
	//resp, err := http.Get(url)
	//if err != nil {
	//	return nil, err
	//}
	//
	//log.Printf("Got the http response from %q\n", url)
	//defer resp.Body.Close()
	//
	//if resp.StatusCode != http.StatusOK {
	//
	//	return nil, fmt.Errorf("getting %s: %s", url, resp.Status)
	//}

	o := append(chromedp.DefaultExecAllocatorOptions[:],
		//... any options here
		chromedp.UserAgent(userAgent),
		//chromedp.ProxyServer("http://username:password@proxyserver.com:31280"),
	)

	cx, cancel := chromedp.NewExecAllocator(context.Background(), o...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(cx)
	defer cancel()

	var HTML string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.OuterHTML("body", &HTML, chromedp.ByQuery),
	); err != nil {
		log.Fatal(err)
	}

	resp := strings.NewReader(HTML)

	bodyBytes, err := ioutil.ReadAll(resp)
	if err != nil {
		log.Fatal(err)
	}

	bodyString := string(bodyBytes)
	fmt.Print(bodyString)

	return html.Parse(resp)
}

func getAmazonResults(node *html.Node) (*html.Node, error) {
	if node.Type == html.ElementNode && node.Data == "div" {
		for _, a := range node.Attr {
			if a.Key == "class" && a.Val == resultsClass {
				return node, nil
			}
		}
	}

	//dfs
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if n, err := getAmazonResults(c); err == nil {
			return n, nil
		}
	}

	return nil, errors.New("product node not found")
}

func getAmazonBooks(node *html.Node) ([]*Book, error) {
	var books []*Book

	// loop through children
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		book := &Book{Source: "Amazon"}
		extractAmazonBook(book, c)
		if book.BookURL != "" && book.Rating != 0.0 {
			books = append(books, book)
		}
	}

	return books, nil
}

func PrintBooks(books ...*Book) {
	for i, book := range books {
		fmt.Println()
		fmt.Println("#", i)
		fmt.Println("Image:", book.ImgURL)
		fmt.Println("Title:", book.Title)
		fmt.Println("Link:", book.BookURL)
		fmt.Println("Author:", book.Author)
		fmt.Printf("Average rating: %.1f\n", book.Rating)
		fmt.Println("Total reviews:", book.Reviews)
		fmt.Println("Browse reviews:", book.ReviewsURL)
		fmt.Println("Book source:", book.Source)
		fmt.Println("====")
	}
}

// the parsing itself
func extractAmazonBook(book *Book, product *html.Node) {
	// check current node
	if img, err := getImage(product); err == nil {
		book.ImgURL = img
	}
	if name, link, err := getTitleAndLink(product); err == nil {
		book.Title = name
		book.BookURL = link
	}
	if auth, err := getAuthor(product); err == nil {
		book.Author = auth
	}
	if rtn, err := getRating(product); err == nil {
		book.Rating = rtn
	}
	if num, link, err := getReviews(product); err == nil {
		book.Reviews = num
		book.ReviewsURL = link
	}

	// loop through every child node
	for child := product.FirstChild; child != nil; child = child.NextSibling {
		extractAmazonBook(book, child)
	}
}

func getImage(n *html.Node) (string, error) {
	if n.Type == html.ElementNode && n.Data == "img" {
		found := false
		for _, attr := range n.Attr {
			if (attr.Key == "class") && (attr.Val == "s-image") {
				found = true
			}
		}
		if found {
			for _, attr2 := range n.Attr {
				if attr2.Key == "src" {
					return attr2.Val, nil
				}
			}
		}
	}
	return "", errors.New("didn't find the image link this time")
}

// only testing and returning string if all is ok
func getTitleAndLink(n *html.Node) (string, string, error) {
	if n.Type == html.ElementNode && n.Data == "h2" {
		return getText(n), getLink(n.FirstChild), nil
	}

	return "", "", errors.New("no title or link in this node")
}

func getAuthor(n *html.Node) (string, error) {
	if n.Type == html.ElementNode && n.Data == "span" && n.FirstChild != nil && n.FirstChild.Data == "by " {
		if link := n.NextSibling.FirstChild; link != nil {
			return link.Data, nil
		}
	}

	return "", errors.New("no author in this node")
}

func getRating(n *html.Node) (float64, error) {
	if n.Type == html.TextNode {
		match, _ := regexp.MatchString("... out of . stars", n.Data)
		if match {
			return strconv.ParseFloat(n.Data[:3], 32)
		}
	}

	return 0.0, errors.New("no rating in this node")
}

func getReviews(n *html.Node) (int, string, error) {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" && strings.Contains(attr.Val, "#customerReviews") {
				link := amazonURL + attr.Val
				num := getText(n)
				num = strings.ReplaceAll(num, ",", "")
				num = strings.TrimSpace(num)
				reviews, err := strconv.Atoi(num)

				return reviews, link, err
			}
		}
	}
	return 0, "", errors.New("no reviews for this node")
}

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

func getLink(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				return amazonURL + attr.Val
			}
		}
	}
	return ""
}
