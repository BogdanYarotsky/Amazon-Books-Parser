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
	amazonBooks, err := FindAmazonBooks(query)
	if err != nil && amazonBooks == nil {
		panic(err)
	}
	//topAmazon := SortAmazonBooks(amazonBooks)
	//PrintBooks(topAmazon...)

	// Goodreads top
	goodreadsBooks, err := FindGoodreadsBooks(query)
	if err != nil && goodreadsBooks == nil {
		panic(err)
	}

	//topGoodreads := SortGoodreadsBooks(goodreadsBooks)
	//PrintBooks(topGoodreads...)

	//var similar []*Book
	//for _, bookA := range topAmazon {
	//	for _, bookG := range topGoodreads {
	//		if bookA.Author == bookG.Author {
	//			similar = append(similar, bookA)
	//		}
	//	}
	//}

	//fmt.Println("Similar books:")
	//PrintBooks(similar...)

}
