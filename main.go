package main

import (
	"os"
	"sort"
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

	// get rid of book less than 4.4
	var betterBooks []*Book

	for _, book := range books {
		if book.Rating > 4.5 {
			betterBooks = append(betterBooks, book)
		}
	}

	// get those with most reviews
	sort.Slice(betterBooks, func(i, j int) bool {
		return betterBooks[i].Reviews > betterBooks[j].Reviews
	})

	top := betterBooks[:10]

	// sort them by rating again, it's reasonable
	sort.Slice(top, func(i, j int) bool {
		if top[i].Rating != top[j].Rating {
			return top[i].Rating > top[j].Rating
		} else {
			return top[i].Reviews > top[j].Reviews
		}

	})

	PrintBooks(top...)
}
