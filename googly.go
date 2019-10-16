package main

import (
	"./engines"

	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"time"
	"strconv"
	"strings"

	"github.com/akamensky/argparse"
)

const UA = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.78 Safari/537.36 Vivaldi/2.8.1664.35"


func main() {
	parser := argparse.NewParser("", "Search using various search engines from the comfort of your terminal")
	query := parser.String("q", "query", &argparse.Options{Required: true, Help: "String to query search engine for",})
	lang := parser.String("l", "lang", &argparse.Options{Help: "Search result language", Default: "en",})
	pages := parser.Int("p", "pages", &argparse.Options{Help: "The amount of pages to scrape", Default: 5})
	format := parser.Selector("f", "format", []string{"cli", "json", "xml"}, &argparse.Options{Help: "Output format", Default: "cli"})
	engine := parser.Selector("e", "engine", []string{"google", "ecosia", "startpage", "yahoo", "ddg", "naver"}, &argparse.Options{Help: "Search engine to use", Default: "google"})
	verbose := parser.Flag("v", "verbose", &argparse.Options{Help: "Print more request infos"})
	from := parser.String("", "from", &argparse.Options{Help: "Start date for the search"})
	to := parser.String("", "to", &argparse.Options{Help: "End date for the search"})
	time := parser.Selector("t", "time-range", []string{"any", "hour", "day", "week", "month", "year"}, &argparse.Options{Help: "Time range in which to search", Default: "any"})
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	crawl(*engine, *query, *lang, *pages, *format, *verbose, *from, *to, *time)
}

func crawl(engine string, query string, lang string, pages int, format string, verbose bool, from string, to string, timerange string) {
	var searchEngine engines.SearchEngine

	switch engine {
	case "google":
		searchEngine = engines.Google()
	case "ecosia":
		searchEngine = engines.Ecosia()
	case "startpage":
		searchEngine = engines.Startpage()
	case "yahoo":
		searchEngine = engines.Yahoo()
	case "ddg":
		searchEngine = engines.DuckDuckGo()
	case "naver":
		searchEngine = engines.Naver()
	}

	var fromTime *time.Time
	var toTime *time.Time

	if from != "" {
		tmp := parseDate(from)
		fromTime = &tmp
	}

	if to != "" {
		tmp := parseDate(to)
		toTime = &tmp
	}

	results := searchEngine.Crawl(query, &engines.SearchOptions{
		Lang: lang,
		Pages: pages,
		Verbose: verbose,
		From: fromTime,
		To: toTime,
		Timerange: timerange,
	})

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

func parseDate(str string) time.Time {
	parts := strings.Split(str, "-")
	year, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	day, _ := strconv.Atoi(parts[2])
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
