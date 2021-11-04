package main

import (
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	resultsClass = "s-main-slot s-result-list s-search-results sg-row"
	amazonURL    = "https://www.amazon.com"
)

func FindAmazonBooks(node *html.Node) ([]*Book, error) {
	var booklist []*Book

	results, err := getAmazonResults(node)
	if err != nil {
		return nil, err
	}

	books, err := getAmazonBooks(results)
	if err != nil {
		return nil, err
	}

	booklist = append(booklist, books...)

	return booklist, nil
}

func SortAmazonBooks(books []*Book) []*Book {
	// get rid of book less than 4.4
	var betterBooks []*Book
	for _, book := range books {
		if book.Source == Amazon && book.Rating > 4.5 {
			betterBooks = append(betterBooks, book)
		}
	}

	// get those with most reviews
	sort.Slice(betterBooks, func(i, j int) bool {
		return betterBooks[i].Reviews > betterBooks[j].Reviews
	})

	var top []*Book
	if len(betterBooks) >= 12 {
		top = betterBooks[:12]
	} else {
		top = betterBooks
	}

	// sort them by rating again, it's reasonable
	sort.Slice(top, func(i, j int) bool {
		if top[i].Rating != top[j].Rating {
			return top[i].Rating > top[j].Rating
		} else {
			return top[i].Reviews > top[j].Reviews
		}

	})

	return top
}

func createAmazonURLs(query string) ([]string, error) {
	query = strings.ReplaceAll(query, "+", "%2B")
	searchString := "/Books-Search/s?k=" + strings.ReplaceAll(query, " ", "+")
	selectTopRated := "&i=stripbooks&rh=n%3A283155%2Cp_72%3A1250221011&dc"
	var urls []string
	for i := 1; i <= 2; i++ {
		pageQuery := fmt.Sprintf("&page=%d&qid=1634582114&rnid=1250219011&ref=sr_pg_%d", i, i)
		urls = append(urls, amazonURL+searchString+selectTopRated+pageQuery)
	}
	return urls, nil
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
		book := &Book{Source: Amazon}
		extractAmazonBook(book, c)
		if book.BookURL != "" && book.Rating != 0.0 {
			books = append(books, book)
		}
	}

	return books, nil
}

// the parsing itself
func extractAmazonBook(book *Book, product *html.Node) {
	// check current node
	if img, err := getAmazonBookImage(product); err == nil {
		book.ImgURL = img
	}
	if name, link, err := getAmazonTitleAndLink(product); err == nil {
		book.Title = name
		book.BookURL = link
	}
	if auth, err := getAmazonAuthor(product); err == nil {
		book.Author = auth
	}
	if rtn, err := getAmazonRating(product); err == nil {
		book.Rating = rtn
	}
	if num, link, err := getAmazonReviews(product); err == nil {
		book.Reviews = num
		book.ReviewsURL = link
	}

	for child := product.FirstChild; child != nil; child = child.NextSibling {
		extractAmazonBook(book, child)
	}
}

func getAmazonBookImage(n *html.Node) (string, error) {
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
func getAmazonTitleAndLink(n *html.Node) (string, string, error) {
	if n.Type == html.ElementNode && n.Data == "h2" {
		return getText(n), amazonURL + getHref(n.FirstChild), nil
	}

	return "", "", errors.New("no title or link in this node")
}

func getAmazonAuthor(n *html.Node) (string, error) {
	if n.Type == html.ElementNode && n.Data == "span" && n.FirstChild != nil && n.FirstChild.Data == "by " {
		if link := n.NextSibling.FirstChild; link != nil {
			return link.Data, nil
		}
	}

	return "", errors.New("no author in this node")
}

func getAmazonRating(n *html.Node) (float64, error) {
	if n.Type == html.TextNode {
		match, _ := regexp.MatchString("... out of . stars", n.Data)
		if match {
			return strconv.ParseFloat(n.Data[:3], 32)
		}
	}

	return 0.0, errors.New("no rating in this node")
}

func getAmazonReviews(n *html.Node) (int, string, error) {
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
