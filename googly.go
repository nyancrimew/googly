package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/akamensky/argparse"
	"github.com/gocolly/colly"
)

const UA = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.78 Safari/537.36 Vivaldi/2.8.1664.35"

type Result struct {
	Title       string
	Link        string
	Description string
}

type SearchEngine struct {
	SearchUrl          func(query string, lang string) string
	Url                func(path string, lang string) string
	Result             func(e *colly.HTMLElement) Result
	Pagination         func(lang string, e *colly.HTMLElement) string
	resultSelector     string
	paginationSelector string
}

func (en *SearchEngine) Crawl(query string, lang string, pages int, verbose bool) []Result {
	var results []Result

	searchCollector := colly.NewCollector(
		colly.UserAgent(UA),
	)

	searchCollector.WithTransport(&http.Transport{
		DisableCompression: true,
	})

	searchCollector.OnHTML(en.resultSelector, func(e *colly.HTMLElement) {
		if verbose {
			fmt.Println(e)
		}
		results = append(results, en.Result(e))
	})

	var page = 1
	searchCollector.OnHTML(en.paginationSelector, func(e *colly.HTMLElement) {
		if  page < pages || pages == -1 {
			page++
			_ = searchCollector.Visit(en.Pagination(lang, e))
		}
	})

	searchCollector.OnRequest(func(r *colly.Request) {
		if verbose {
			fmt.Println(r.URL)
		}
	})

	searchCollector.OnError(func(r *colly.Response, err error) {
		fmt.Fprintln(os.Stderr, err)
		if verbose {
			fmt.Fprintln(os.Stderr, r)
		}
		os.Exit(r.StatusCode)
	})

	_ = searchCollector.Visit(en.SearchUrl(query, lang))

	return results
}

func main() {
	parser := argparse.NewParser("", "Search using various search engines from the comfort of your terminal")
	query := parser.String("q", "query", &argparse.Options{Required: true, Help: "String to query search engine for",})
	lang := parser.String("l", "lang", &argparse.Options{Help: "Search result language", Default: "en",})
	pages := parser.Int("p", "pages", &argparse.Options{Help: "The amount of pages to scrape", Default: 5})
	format := parser.Selector("f", "format", []string{"cli", "json", "xml"}, &argparse.Options{Help: "Output format", Default: "cli"})
	engine := parser.Selector("e", "engine", []string{"google", "ecosia", "startpage", "yahoo"}, &argparse.Options{Help: "Search engine to use", Default: "google"})
	verbose := parser.Flag("v", "verbose", &argparse.Options{Help: "Print more request infos"})
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	crawl(*engine, *query, *lang, *pages, *format, *verbose)
}

func getUrl(base string, path string, lang string, langName string) string {
	var between = ""
	if !strings.Contains(path, "?") {
		between = "?"
	}
	return fmt.Sprintf("%s/%s%s&%s=%s", base, path, between, langName, lang)
}

func crawl(engine string, query string, lang string, pages int, format string, verbose bool) {
	var searchEngine SearchEngine

	switch engine {
	case "google":
		Url := func(path string, lang string) string {
			return getUrl("https://google.com", path, lang, "hl")
		}
		searchEngine = SearchEngine{
			Url: Url,
			SearchUrl: func(query string, lang string) string {
				return Url(fmt.Sprintf("search?q=%s", query), lang)
			},
			Result: func(e *colly.HTMLElement) Result {
				return Result{
					Title:       e.ChildText("h3"),
					Link:        e.ChildAttr("a", "href"),
					Description: e.ChildText("span.st"),
				}
			},
			Pagination: func(lang string, e *colly.HTMLElement) string {
				return Url(e.Attr("href"), lang)
			},
			resultSelector:     ".g .rc",
			paginationSelector: "a.pn",
		}
	case "ecosia":
		Url := func(path string, lang string) string {
			return getUrl("https://www.ecosia.org", path, lang, "hl")
		}
		searchEngine = SearchEngine{
			Url: Url,
			SearchUrl: func(query string, lang string) string {
				return Url(fmt.Sprintf("search?q=%s", query), lang)
			},
			Result: func(e *colly.HTMLElement) Result {
				return Result{
					Title:       e.ChildText(".result-title"),
					Link:        e.ChildAttr("a.result-title", "href"),
					Description: e.ChildText(".result-snippet"),
				}
			},
			Pagination: func(lang string, e *colly.HTMLElement) string {
				return Url(e.Attr("href"), lang)
			},
			resultSelector:     ".result-body",
			paginationSelector: "a.pagination-next",
		}
	case "startpage":
		Url := func(path string, lang string) string {
			return getUrl("https://www.startpage.com", path, lang, "language")
		}
		searchEngine = SearchEngine{
			Url: Url,
			SearchUrl: func(query string, lang string) string {
				return Url(fmt.Sprintf("do/search?query=%s&prfe=36c84513558a2d34bf0d89ea505333ad761002405484af2476571afac1710d79d80647dbf3b0d6646044dd543d05df3a", query), lang)
			},
			Result: func(e *colly.HTMLElement) Result {
				return Result{
					Title:       e.ChildText(".w-gl__result-title h3"),
					Link:        e.ChildAttr("a.w-gl__result-title", "href"),
					Description: e.ChildText(".w-gl__description"),
				}
			},
			Pagination: func(lang string, e *colly.HTMLElement) string {
				url := e.Request.URL
				var pageStr = url.Query().Get("page")
				if pageStr == "" {
					pageStr = "1"
				}
				var nPage = "2"
				if page, err := strconv.Atoi(pageStr); err == nil {
					nPage = strconv.Itoa(page + 1)
				}
				qry := url.Query()
				qry.Set("page", nPage)
				url.RawQuery = qry.Encode()
				return url.String()
			},
			resultSelector:     ".w-gl__result",
			paginationSelector: "button.next",
		}
	case "yahoo":
		Url := func(path string, lang string) string {
			return getUrl("https://search.yahoo.com", path, lang, "lang")
		}
		searchEngine = SearchEngine{
			Url: Url,
			SearchUrl: func(query string, lang string) string {
				return Url(fmt.Sprintf("search?p=%s", query), lang)
			},
			Result: func(e *colly.HTMLElement) Result {
				return Result{
					Title:       e.ChildText("h3.title"),
					Link:        e.ChildAttr("h3.title a", "href"), // TODO: extract actual URL from redirect URL
					Description: e.ChildText("div.compText"),
				}
			},
			Pagination: func(lang string, e *colly.HTMLElement) string {
				return e.Attr("href")
			},
			resultSelector:     "li div.dd.algo",
			paginationSelector: ".compPagination a.next",
		}
	}

	results := searchEngine.Crawl(query, lang, pages, verbose)

	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")

		_ = enc.Encode(results)
	case "xml":
		enc := xml.NewEncoder(os.Stdout)
		enc.Indent("", "  ")

		_ = enc.Encode(results)
	default:
		for i, el := range results {
			fmt.Println("[", i+1, "] ", el.Title)
			fmt.Println(el.Description)
			fmt.Println(el.Link)
			fmt.Println()
		}
	}
}
