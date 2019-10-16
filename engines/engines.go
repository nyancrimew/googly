package engines

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

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

func (en *SearchEngine) Crawl(query string, lang string, pages int, verbose bool, ua string) []Result {
	var results []Result

	searchCollector := colly.NewCollector(
		colly.UserAgent(ua),
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

func getUrl(base string, path string, lang string, langName string) string {
	var between = ""
	if !strings.Contains(path, "?") {
		between = "?"
	}
	return fmt.Sprintf("%s/%s%s&%s=%s", base, path, between, langName, lang)
}

func Google() SearchEngine {
	Url := func(path string, lang string) string {
		return getUrl("https://google.com", path, lang, "hl")
	}
	return SearchEngine{
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
}

func Ecosia() SearchEngine {
	Url := func(path string, lang string) string {
		return getUrl("https://www.ecosia.org", path, lang, "hl")
	}
	return SearchEngine{
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
}

func Startpage() SearchEngine {
	Url := func(path string, lang string) string {
		return getUrl("https://www.startpage.com", path, lang, "language")
	}
	return SearchEngine{
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
}

func Yahoo() SearchEngine {
	Url := func(path string, lang string) string {
		return getUrl("https://search.yahoo.com", path, lang, "lang")
	}
	return SearchEngine{
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