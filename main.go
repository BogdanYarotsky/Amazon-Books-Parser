package main

import (
	"fmt"
	"os"
)

const sample_query = "python"

func main() {
	var query string

	if len(os.Args) == 2 {
		query = os.Args[1]
	} else {
		query = sample_query
	}

	books, err := GetBooks(query)
	if err != nil {
		panic(err)
	}

	fmt.Println("Amazon top:")
	PrintBooks(SortAmazonBooks(books)...)
	fmt.Println("Goodreads top:")
	PrintBooks(SortGoodreadsBooks(books)...)

}
