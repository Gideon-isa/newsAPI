package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/Gideon-isa/newsapp/news"
	"github.com/joho/godotenv"
)

var tpl *template.Template

// var newsapi *http.Client
type Search struct {
	Query      string
	NextPage   int
	TotalPages int
	Results    *news.Results
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	tpl.Execute(w, nil)
}

func searchHandler(newsapi *news.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := url.Parse(r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		params := u.Query()
		searchQuery := params.Get("q")
		page := params.Get("page")
		if page == "" {
			page = "1"
		}

		results, err := newsapi.FetchEverything(searchQuery, page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		nextPage, err := strconv.Atoi(page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		search := &Search{
			Query:      searchQuery,
			NextPage:   nextPage,
			TotalPages: int(math.Ceil(float64(results.TotalResults)) / float64(newsapi.PageSize)),
			Results:    results,
		}
		buf := &bytes.Buffer{}

		err = tpl.Execute(buf, search)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		buf.WriteTo(w)

		fmt.Printf("%+v\n", results)

		fmt.Println("Search Query is: ", searchQuery)
		fmt.Println("Page is: ", page)
	}

}

func main() {
	tpl = template.Must(template.ParseFiles("index.html"))
	err := godotenv.Load("app.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	fmt.Println("Starting Server...")

	port := os.Getenv("PORT")
	apiKey := os.Getenv("news_api_key")
	if apiKey == "" {
		log.Fatal("Env: apiKey must be set")
	}

	myClient := &http.Client{Timeout: 10 * time.Second}
	newsapi := news.NewClient(myClient, apiKey, 20)
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./assets/"))
	mux.Handle("/assets/", http.StripPrefix("/assets", fs))
	// fs := http.FileServer(http.Dir(""))
	// strp := http.StripPrefix("", fs)
	// mux.Handle("base", strp)

	// How it works for the path
	// the mux.Hanlefunc + the mux.Handle + http.FileServer (the file or file directory)
	// i.e => "/" + "/assets/" + ...
	// "/assets/" the first backslash MUST be Stripped off before it to get working
	// if not it will be `//assets..` which is bad
	// And that is why http.StripPrefix is used to get rid of it

	// NB the folder name(assets) which is used by the http.Dir type as arguement
	// MUST be the same name used as the pattern string for the mux.Handle func as arguement
	// Because we need to inform the mux router to use this file server object for all paths
	// beginning with the /assets/ prefix
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/search", searchHandler(newsapi))

	http.ListenAndServe(port, mux)

}
