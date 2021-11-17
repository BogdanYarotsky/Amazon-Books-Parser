package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type TopplerResult struct {
	Title          string
	AmazonBooks    []*Book
	GoodreadsBooks []*Book
}

// http://localhost:8080/search?q=ffff
var indexPath = regexp.MustCompile("/.")

//var resultPath = regexp.MustCompile("^/(results)/([a-zA-Z0-9]+)$")
var resultPath = regexp.MustCompile(`^/(search\?q)=([\w\+]+)$`)

var templates = template.Must(template.ParseGlob("templates/*.html"))

// for static files
func serveStaticFiles() {
	fs := http.FileServer(http.Dir("./static/"))
	http.Handle("/static/", http.StripPrefix("/static", fs))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	m := indexPath.FindStringSubmatch(r.URL.Path)

	if m != nil {
		http.NotFound(w, r)
		return
	}

	p := &TopplerResult{Title: "Hello Toppler"}
	err := templates.ExecuteTemplate(w, "index.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.String())
	m := resultPath.FindStringSubmatch(r.URL.String())
	log.Println(m)

	if m == nil {
		http.NotFound(w, r)
		return
	}

	if m[2] == "" {
		http.NotFound(w, r)
		return
	}

	books, err := GetBooks(strings.ReplaceAll(m[2], "+", " "))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	amazon, goodreads := SortAmazonBooks(books), SortGoodreadsBooks(books)

	p := &TopplerResult{Title: "Hello Toppler", AmazonBooks: amazon, GoodreadsBooks: goodreads}
	err = templates.ExecuteTemplate(w, "results.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	fmt.Println("Serving on http://localhost:8080/")
	serveStaticFiles()
	http.HandleFunc("/search", queryHandler)
	http.HandleFunc("/", indexHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
	// will be moved to the method for page results

	//var query string
	//if len(os.Args) == 2 {
	//	query = os.Args[1]
	//}
	//books, err := GetBooks(query)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("Amazon top:")
	//PrintBooks(SortAmazonBooks(books)...)
	//fmt.Println("Goodreads top:")
	//PrintBooks(SortGoodreadsBooks(books)...)

}
