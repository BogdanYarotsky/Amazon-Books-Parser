package main

import (
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Book struct {
	Index   int     // done
	Title   string  // done
	Author  string  // done
	Rating  float64 // done
	Reviews int     // done

	ImgURL     string // done
	BookURL    string // done
	ReviewsURL string // to do
}

const spanID = "MAIN-SEARCH_RESULTS-.*"

func FindAmazonBooks(url string) ([]Book, error) {
	root, err := getRootNode(url)
	if err != nil {
		return nil, err
	}

	// step 1 - get a slice of "root" nodes - spans with all necessary info - done
	products, err := getProductNodes(root)
	if err != nil {
		return nil, err
	}
	//printNodes(products)

	//printProductTree(products[0])

	// step 2 - traverse spans for very specific info
	var books []*Book
	for i, product := range products {
		book := &Book{Index: i}
		extractBookInfo(book, product)
		books = append(books, book)
	}

	// debug
	printBooks(books)
	// return books, nil
	return nil, nil
}

// All starts here
func getRootNode(url string) (*html.Node, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	log.Printf("Got the http response from %q\n", url)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		return nil, fmt.Errorf("getting %s: %s", url, resp.Status)
	}

	return html.Parse(resp.Body)
}

// It just works
func getProductNodes(node *html.Node) ([]*html.Node, error) {
	var products []*html.Node
	if node.Type == html.ElementNode && node.Data == "span" {
		for _, a := range node.Attr {
			match, _ := regexp.MatchString(spanID, a.Val)
			if match {
				products = append(products, node)
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		nodes, _ := getProductNodes(c)
		products = append(products, nodes...)
	}
	return products, nil
}

// for debug
func printBooks(books []*Book) {
	for _, book := range books {
		fmt.Println("#", book.Index)
		fmt.Println("Image: ", book.ImgURL)
		fmt.Println("Title: ", book.Title)
		fmt.Println("Link: ", book.BookURL)
		fmt.Println("Author: ", book.Author)
		fmt.Printf("Average rating: %.1f\n", book.Rating)
		fmt.Println("Total reviews: ", book.Reviews)
		fmt.Println("Browse reviews: ", book.ReviewsURL)
		fmt.Println("====")
	}
}

// the parsing itself
func extractBookInfo(book *Book, product *html.Node) {
	fmt.Println("Start: ", book)

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

	for child := product.FirstChild; child != nil; child = child.NextSibling {
		extractBookInfo(book, child)
	}

	fmt.Println("Finish: ", book)
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
		fmt.Println("I've found the header!")
		return getText(n), getLink(n.FirstChild), nil
	}

	return "", "", errors.New("no title or link in this node")
}

func getAuthor(n *html.Node) (string, error) {
	if n.Type == html.ElementNode && n.Data == "span" && n.FirstChild != nil && n.FirstChild.Data == "by " {
		fmt.Println("I've found the author!")
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
				link := "https://amazon.com" + attr.Val
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
		fmt.Println(n.Data)
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
				return "https://amazon.com" + attr.Val
			}
		}
	}
	return ""
}
