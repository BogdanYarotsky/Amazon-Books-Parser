package main

import (
	"html/template"
	"log"
	"net/http"
	"regexp"
)

type TopplerResult struct {
	Title string
	Books []Book
}

var indexPath = regexp.MustCompile("/.")
var resultPath = regexp.MustCompile("^/(results)/([a-zA-Z0-9]+)$")
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

	p := &TopplerResult{Title: "Hello Toppler", Books: nil}
	err := templates.ExecuteTemplate(w, "index.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	serveStaticFiles()
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
