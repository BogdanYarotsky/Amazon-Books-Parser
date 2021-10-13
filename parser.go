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
	Index   int
	Title   string
	Author  string
	Rating  float64
	Reviews int

	ImgURL     string
	BookURL    string
	AuthorURL  string
	ReviewsURL string
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

	//// debug print
	//bodyBytes, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//bodyString := string(bodyBytes)
	//fmt.Print(bodyString)

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

// For debug
func printNodes(nodes []*html.Node) {
	for _, node := range nodes {
		fmt.Println(node)
	}
}

//for debug
func printProductTree(n *html.Node) {
	fmt.Println(n)

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		printProductTree(child)
	}
}

// for debug
func printBooks(books []*Book) {
	for _, book := range books {
		fmt.Println("#", book.Index)
		fmt.Println("Title: ", book.Title)
		fmt.Println("Author: ", book.Author)
		fmt.Printf("Average rating: %.1f\n", book.Rating)
		fmt.Println("Total reviews: ", book.Reviews)
		fmt.Println("====")
	}
}

// the parsing itself
func extractBookInfo(book *Book, product *html.Node) {
	fmt.Println("Start: ", book)

	if t, err := getTitle(product); err == nil {
		fmt.Println("Got title!: ", t)
		book.Title = t
	}

	if a, err := getAuthor(product); err == nil {
		fmt.Println("Got author!: ", a)
		book.Author = a
	}

	if r, err := getRating(product); err == nil {
		book.Rating = r
	}

	if rvw, err := getReviews(product); err == nil {
		book.Reviews = rvw
	}

	//book.Rating = getRating(product)
	//book.Reviews = getReviews() // check the current node for # reviews

	for child := product.FirstChild; child != nil; child = child.NextSibling {
		extractBookInfo(book, child)
	}

	fmt.Println("Finish: ", book)
}

// only testing and returning string if all is ok
func getTitle(n *html.Node) (string, error) {

	if n.Type == html.ElementNode && n.Data == "h2" {
		fmt.Println("I've found the header!")
		return getText(n), nil
	}

	return "", errors.New("No title in this node")
}

func getAuthor(n *html.Node) (string, error) {
	if n.Type == html.ElementNode && n.Data == "span" && n.FirstChild != nil && n.FirstChild.Data == "by " {
		fmt.Println("I've found the author!")
		if link := n.NextSibling.FirstChild; link != nil {
			return link.Data, nil
		}
	}

	return "", errors.New("No author in this node")
}

func getRating(n *html.Node) (float64, error) {
	if n.Type == html.TextNode {
		match, _ := regexp.MatchString("... out of . stars", n.Data)
		if match {
			return strconv.ParseFloat(n.Data[:3], 32)
		}
	}

	return 0.0, errors.New("No rating in this node")
}

func getReviews(n *html.Node) (int, error) {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" && strings.Contains(attr.Val, "#customerReviews") {
				num := getText(n)
				num = strings.ReplaceAll(num, ",", "")
				num = strings.TrimSpace(num)
				fmt.Println("Number string:", num)
				return strconv.Atoi(num)
			}
		}

	}
	return 0, errors.New("No reviews for this node")
}

func getText(n *html.Node) string {
	fmt.Println("Trying to get header text")
	var text string
	if n.Type == html.TextNode {
		fmt.Println(n.Data)
		return n.Data
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += getText(c)
	}

	result := strings.Trim(text, "\n")
	fmt.Println("Final result: ", result)
	return result
}
