package engines

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type Result struct {
	Title       string
	Link        string
	Description string
}

type SearchEngine struct {
	SearchUrl          func(query string, options *SearchOptions) string
	Url                func(path string, lang string) string
	Result             func(e *colly.HTMLElement) Result
	Pagination         func(page int, options *SearchOptions, e *colly.HTMLElement) string
	resultSelector     string
	paginationSelector string
	browserConfig BrowserConfig
}

type SearchOptions struct {
	Lang  		string
	Pages 		int
	From  		*time.Time
	To    		*time.Time
	Timerange   string
	UserAgent 	string
	Verbose 	bool
}

func (en *SearchEngine) Crawl(query string, options *SearchOptions) []Result {
	var results []Result

	searchCollector := colly.NewCollector()
	if options.UserAgent != "" {
		searchCollector.UserAgent = options.UserAgent
	} else {
		searchCollector.UserAgent = RandomUA(&en.browserConfig)
	}
	if options.Verbose {
		fmt.Println(searchCollector.UserAgent)
	}

	searchCollector.WithTransport(&http.Transport{
		DisableCompression: true,
	})

	searchCollector.OnHTML(en.resultSelector, func(e *colly.HTMLElement) {
		if options.Verbose {
			c := e.DOM.Nodes[0]
			p := e.DOM.Parent().Nodes[0]
			fmt.Println("Selected:", c.Attr, "Parent:", p.Attr)
		}
		results = append(results, en.Result(e))
	})

	var page = 1
	searchCollector.OnHTML(en.paginationSelector, func(e *colly.HTMLElement) {
		if  page < options.Pages || options.Pages == -1 {
			page++
			_ = searchCollector.Visit(en.Pagination(page, options, e))
		}
	})

	searchCollector.OnRequest(func(r *colly.Request) {
		if options.Verbose {
			fmt.Println(r.URL)
		}
	})

	searchCollector.OnError(func(r *colly.Response, err error) {
		fmt.Fprintln(os.Stderr, err)
		if options.Verbose {
			fmt.Fprintln(os.Stderr, r)
		}
		os.Exit(r.StatusCode)
	})

	_ = searchCollector.Visit(en.SearchUrl(query, options))

	return results
}

func Combined(query string, options *SearchOptions, engines ... SearchEngine) []Result {
	ch := make(chan []Result, len(engines))
	for _, eng := range engines {
		go func(eng SearchEngine) {
			ch <- eng.Crawl(query, options)
		} (eng)
	} 
	var results [][]Result
	for i := 0; i < len(engines); i++ {
		results = append(results, <-ch)
	}
	return unique(merge(results))
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
		browserConfig: BrowserConfig {
			chrome: true,
			firefox: true,
		},
		Url: Url,
		SearchUrl: func(query string, options *SearchOptions) string {
			var timebox = ""
			if options.From != nil || options.To != nil {
				timebox += "&tbs=" + url.QueryEscape("cdr:1")
				if options.From != nil {
					timebox += url.QueryEscape(fmt.Sprintf(",cd_min:%02d/%02d/%d", options.From.Month(), options.From.Day(), options.From.Year()))
				}
				if options.To != nil {
					timebox += url.QueryEscape(fmt.Sprintf(",cd_max:%02d/%02d/%d", options.To.Month(), options.To.Day(), options.To.Year()))
				}
			} else if options.Timerange != "any" {
				timebox = "&tbs=qdr:"
				switch options.Timerange {
				case "hour":
					timebox += "h"
				case "day":
					timebox += "d"
				case "week":
					timebox += "w"
				case "month":
					timebox += "m"
				case "year":
					timebox += "y"
				}
			}
			return Url(fmt.Sprintf("search?q=%s%s", query, timebox), options.Lang)
		},
		Result: func(e *colly.HTMLElement) Result {
			return Result{
				Title:       e.ChildText("h3"),
				Link:        e.ChildAttr("a", "href"),
				Description: e.ChildText("span.st"),
			}
		},
		Pagination: func(page int, options *SearchOptions, e *colly.HTMLElement) string {
			return Url(e.Attr("href"), options.Lang)
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
		browserConfig: BrowserConfig {
			chrome: true,
			firefox: true,
			chromeM: true,
			firefoxM: true,
		},
		Url: Url,
		SearchUrl: func(query string, options *SearchOptions) string {
			var timebox = ""
			if options.Timerange != "any" {
				timebox = "&freshness="
				switch options.Timerange {
				case "hour":
					// TODO: is there really no way to do this?
					timebox += "day"
				case "day":
					timebox += "day"
				case "week":
					timebox += "week"
				case "month":
					timebox += "month"
				case "year":
					// TODO: is there really no way to do this?
					timebox += "month"
				}
			}
			return Url(fmt.Sprintf("search?q=%s%s", query, timebox), options.Lang)
		},
		Result: func(e *colly.HTMLElement) Result {
			return Result{
				Title:       e.ChildText(".result-title"),
				Link:        e.ChildAttr("a.result-title", "href"),
				Description: e.ChildText(".result-snippet"),
			}
		},
		Pagination: func(page int, options *SearchOptions, e *colly.HTMLElement) string {
			return Url(e.Attr("href"), options.Lang)
		},
		resultSelector:     ".js-result .result-body",
		paginationSelector: "a.pagination-next",
	}
}

func Startpage() SearchEngine {
	Url := func(path string, lang string) string {
		return getUrl("https://www.startpage.com", path, lang, "language")
	}
	return SearchEngine{
		browserConfig: BrowserConfig {
			chrome: true,
			firefox: true,
			chromeM: true,
			firefoxM: true,
		},
		Url: Url,
		SearchUrl: func(query string, options *SearchOptions) string {
			var timebox = ""
			if options.Timerange != "any" {
				timebox = "&with_date="
				switch options.Timerange {
				case "hour":
					timebox += "h"
				case "day":
					timebox += "d"
				case "week":
					timebox += "w"
				case "month":
					timebox += "m"
				case "year":
					timebox += "y"
				}
			}
			return Url(fmt.Sprintf("do/search?query=%s&prfe=36c84513558a2d34bf0d89ea505333ad761002405484af2476571afac1710d79d80647dbf3b0d6646044dd543d05df3a%s", query, timebox), options.Lang)
		},
		Result: func(e *colly.HTMLElement) Result {
			return Result{
				Title:       e.ChildText(".w-gl__result-title h3"),
				Link:        e.ChildAttr("a.w-gl__result-title", "href"),
				Description: e.ChildText(".w-gl__description"),
			}
		},
		Pagination: func(page int, options *SearchOptions, e *colly.HTMLElement) string {
			url := e.Request.URL
			qry := url.Query()
			qry.Set("page", strconv.Itoa(page))
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
		browserConfig: BrowserConfig {
			chrome: true,
			firefox: true,
			chromeM: true,
			firefoxM: true,
		},
		Url: Url,
		SearchUrl: func(query string, options *SearchOptions) string {
			var timebox = ""
			if options.Timerange != "any" {
				timebox = "&fr2=time&age=1"
				switch options.Timerange {
				case "hour":
					// TODO: is there really no way to do this?
					timebox += "d&btf=d"
				case "day":
					timebox += "d&btf=d"
				case "week":
					timebox += "w&btf=w"
				case "month":
					timebox += "m&btf=m"
				case "year":
					// TODO: is there really no way to do this?
					timebox += "m&btf=m"
				}
			}
			return Url(fmt.Sprintf("search?p=%s%s", query, timebox), options.Lang)
		},
		Result: func(e *colly.HTMLElement) Result {
			return Result{
				Title:       e.ChildText("h3.title"),
				Link:        e.ChildAttr("h3.title a", "href"), // TODO: extract actual URL from redirect URL
				Description: e.ChildText("div.compText"),
			}
		},
		Pagination: func(page int, options *SearchOptions, e *colly.HTMLElement) string {
			return e.Attr("href")
		},
		resultSelector:     ".algo-sr",
		paginationSelector: ".compPagination a.next",
	}
}

func DuckDuckGo() SearchEngine {
	Url := func(path string, lang string) string {
		return getUrl("https://duckduckgo.com", path, lang, "kl")
	}
	return SearchEngine{
		browserConfig: BrowserConfig {
			chrome: true,
			firefox: true,
			chromeM: true,
			firefoxM: true,
		},
		Url: Url,
		SearchUrl: func(query string, options *SearchOptions) string {
			var timebox = ""
			if options.Timerange != "any" {
				timebox = "&df="
				switch options.Timerange {
				case "hour":
					// TODO: is there really no way to do this?
					timebox += "d"
				case "day":
					timebox += "d"
				case "week":
					timebox += "w"
				case "month":
					timebox += "m"
				case "year":
					timebox += "y"
				}
			}
			return Url(fmt.Sprintf("html?q=%s&kd=-1&kc=-1&kac=-1&k1=-1&kk=-1&kak=-1&kax=-1&kaq=-1&kao=-1&kap=-1&kau=-1&kz=-1%s", query, timebox), options.Lang)
		},
		Result: func(e *colly.HTMLElement) Result {
			return Result{
				Title:       e.ChildText(".result__title"),
				Link:        e.ChildAttr("a.result__a", "href"),
				Description: e.ChildText(".result__snippet"),
			}
		},
		Pagination: func(page int, options *SearchOptions, e *colly.HTMLElement) string {
			p := e.DOM.Parent()
			url := e.Request.URL
			qry := url.Query()
			s, _ := p.Find("[name='s']").Attr("value")
			qry.Set("s", strings.TrimSpace(s))
			dc, _ := p.Find("[name='dc']").Attr("value")
			qry.Set("dc", strings.TrimSpace(dc))
			url.RawQuery = qry.Encode()
			return url.String()
		},
		resultSelector:     ".serp__results .result",
		paginationSelector: ".nav-link [value='Next']",
	}
}
// todo fix
func Naver() SearchEngine {
	Url := func(path string, lang string) string {
		return getUrl("https://search.naver.com", path, lang, "hl")
	}
	return SearchEngine{
		browserConfig: BrowserConfig {
			chrome: true,
			firefox: true,
			chromeM: true,
			firefoxM: true,
		},
		Url: Url,
		SearchUrl: func(query string, options *SearchOptions) string {
			return Url(fmt.Sprintf("search.naver?where=webkr&query=%s", query), options.Lang)
		},
		Result: func(e *colly.HTMLElement) Result {
			return Result{
				Title:       e.ChildText(".title_link"),
				Link:        e.ChildAttr("a.title_link", "href"),
				Description: e.ChildText(".sh_web_passage"),
			}
		},
		Pagination: func(page int, options *SearchOptions, e *colly.HTMLElement) string {
			return Url("search.naver" + e.Attr("href"), options.Lang)
		},
		resultSelector:     ".paging a.next",
		paginationSelector: "a.pagination-next",
	}
}