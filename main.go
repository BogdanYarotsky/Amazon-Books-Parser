package main

import (
	"fmt"
)

const url = "https://www.amazon.com/s?k=python+books&ref=nb_sb_noss_2"

func main() {
	books, err := FindAmazonBooks(url)
	if err != nil {
		panic(err)
	}

	fmt.Println("Books: ")

	for i, book := range books {
		//fmt.Println("Href: ", i.Href)
		fmt.Println("#", i)
		fmt.Println("Title: ", book.Title)
		fmt.Println("Author: ", book.Author)
		fmt.Println("Average rating: ", book.Rating)
		fmt.Println("Total reviews: ", book.Reviews)
		fmt.Println()
	}
}
