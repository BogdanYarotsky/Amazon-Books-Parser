package main

import (
	"os"
)

type Book struct {
	Title   string  // done
	Author  string  // done
	Rating  float64 // done
	Reviews int     // done

	ImgURL     string // done
	BookURL    string // done
	ReviewsURL string // done
	Source     string // enum?
	Year       int    // to do
}

// debug
const sample_query = "python"

func main() {
	var query string

	if len(os.Args) == 2 {
		query = os.Args[1]
	} else {
		query = sample_query
	}

	books, err := FindAmazonBooks(query)
	if err != nil {
		panic(err)
	}

	top := SortAmazonBooks(books)

	PrintBooks(top...)
}
