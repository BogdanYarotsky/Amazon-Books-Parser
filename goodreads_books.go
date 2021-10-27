package main

import (
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"sort"
	"strconv"
	"strings"
)

const goodreadsURL = "https://www.goodreads.com/"

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
	if node.Type == html.ElementNode {
		switch node.Data {
		case "span":
			for _, attr := range node.Attr {
				if attr.Key == "class" && attr.Val == "minirating" {
					ratingAndReviews := getText(node)
					elems := strings.Fields(ratingAndReviews)
					rating, err := strconv.ParseFloat(elems[0], 32)
					if err != nil {
						return
					}
					book.Rating = rating
					reviews := strings.ReplaceAll(elems[4], ",", "")
					num, err := strconv.Atoi(reviews)
					if err != nil {
						return
					}
					book.Reviews = num
					break
				}

				// can be either author or title
				if attr.Key == "itemprop" && attr.Val == "name" {
					if len(node.Attr) > 1 {
						book.Title = node.FirstChild.Data
					} else {
						book.Author = node.FirstChild.Data
					}
					break
				}
			}

		case "a":
			if node.Attr[0].Val == "bookTitle" {
				book.BookURL = goodreadsURL + node.Attr[2].Val
				book.ReviewsURL = book.BookURL
			}

		case "img":
			if node.Attr[1].Val == "bookCover" {
				book.ImgURL = node.Attr[3].Val
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
	for i := 1; i <= 3; i++ {
		pageQuery := fmt.Sprintf("&page=%d&search_type=books", i)
		urls = append(urls, searchString+pageQuery)
	}
	return urls, nil
}

func SortGoodreadsBooks(books []*Book) []*Book {
	// get rid of book less than 4.4
	var betterBooks []*Book
	for _, book := range books {
		if book.Rating > 3.7 {
			betterBooks = append(betterBooks, book)
		}
	}

	// get those with most reviews
	sort.Slice(betterBooks, func(i, j int) bool {
		return betterBooks[i].Reviews > betterBooks[j].Reviews
	})
	top := betterBooks[:12]

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
