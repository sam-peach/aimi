package main

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/gocolly/colly"
	"github.com/sam-peach/rus_scaper/scraping"
)

const baseUrl string = "https://www.rt.com"
const searchUrl string = "https://www.rt.com/search?q=ai"

func main() {
	ctx := context.Background()
	// ai.InitLanguageClient(ctx)

	linkTable := make(map[string]bool)
	c := colly.NewCollector()

	pattern := "^/(news|business|russia)/.+"

	re := regexp.MustCompile(pattern)

	c.OnHTML("a", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		_, found := linkTable[href]
		if re.MatchString(href) && !found {
			linkTable[href] = true
		}
	})

	c.Visit(baseUrl)

	c.OnHTMLDetach("a")

	links := make([]string, 0)
	for link := range linkTable {
		links = append(links, link)
	}

	fmt.Printf(">> Found %d links\n", len(links))

	for _, link := range links {
		scraping.ProcessLink(ctx, link, c)
	}

	os.Exit(0)
}
