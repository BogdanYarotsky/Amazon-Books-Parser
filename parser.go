package main

import (
	"fmt"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type Book struct {
	Index   int
	Title   string
	Author  string
	Rating  float32
	Reviews int
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
		book := &Book{}
		extractBookInfo(book, product)
		book.Index = i
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
		fmt.Println("Average rating: ", book.Rating)
		fmt.Println("Total reviews: ", book.Reviews)
		fmt.Println("====")
	}
}

// the parsing itself
func extractBookInfo(book *Book, product *html.Node) {
	fmt.Println("Start: ", book)

	if t := getTitle(product); t != "" {
		fmt.Println("Got title!: ", t)
		book.Title = t
	}

	//book.Author = getAuthor() // check the current node for author name
	//book.Rating = getRating()
	//book.Reviews = getReviews() // check the current node for # reviews

	for child := product.FirstChild; child != nil; child = child.NextSibling {
		extractBookInfo(book, child)
	}

	fmt.Println("Finish: ", book)
}

// only testing and returning string if all is ok
func getTitle(n *html.Node) string {

	if n.Type == html.ElementNode && n.Data == "h2" {
		fmt.Println("I've found the header!")
		return getText(n)
	}

	return ""
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
