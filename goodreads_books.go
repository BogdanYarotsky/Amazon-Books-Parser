package main

import (
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"strings"
)

func FindGoodreadsBooks(query string) ([]*Book, error) {
	pages, err := createGoodreadsURLs(query)
	if err != nil {
		return nil, err
	}

	roots, err := getParsedHTMLs(pages)
	if err != nil {
		return nil, err
	}

	var booklist []*Book
	for _, root := range roots {
		results, err := getGoodreadsResults(root)
		if err != nil {
			return nil, err
		}

		books, err := getGoodreadsBooks(results)
		if err != nil {
			return nil, err
		}

		booklist = append(booklist, books...)
	}

	return booklist, nil
}

func getGoodreadsBooks(node *html.Node) ([]*Book, error) {
	var books []*Book
	// loop through children
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "tr" {
			book := &Book{Source: "Goodreads"}
			extractGoodreadsBook(book, c)
			books = append(books, book)
		}
	}
	return books, nil
}

func extractGoodreadsBook(book *Book, node *html.Node) {
	if node.Type == html.ElementNode && node.Data == "span" {
		for _, attr := range node.Attr {
			if attr.Val == "minirating" {
				book.Title = getText(node)
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		extractGoodreadsBook(book, c)
	}

}

func getGoodreadsResults(node *html.Node) (*html.Node, error) {
	if node.Type == html.ElementNode && node.Data == "tbody" {
		return node, nil
	}

	//dfs
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if n, err := getGoodreadsResults(c); err == nil {
			return n, nil
		}
	}

	return nil, errors.New("product node not found")
}

func createGoodreadsURLs(query string) ([]string, error) {
	searchString := "https://www.goodreads.com/search?q=" + strings.ReplaceAll(query, " ", "+")
	var urls []string
	for i := 1; i <= 1; i++ {
		pageQuery := fmt.Sprintf("&page=%d&search_type=books", i)
		urls = append(urls, searchString+pageQuery)
	}
	return urls, nil
}
