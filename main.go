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

const sample_query = "python"

func main() {
	var query string

	if len(os.Args) == 2 {
		query = os.Args[1]
	} else {
		query = sample_query
	}

	// Amazon top
	//amazonBooks, err := FindAmazonBooks(query)
	//if err != nil {
	//	panic(err)
	//}
	//top := SortAmazonBooks(amazonBooks)
	//PrintBooks(top...)

	// Goodreads top
	goodreadsBooks, err := FindGoodreadsBooks(query)
	if err != nil && goodreadsBooks == nil {
		panic(err)
	}
	top := SortGoodreadsBooks(goodreadsBooks)
	PrintBooks(top...)
}
