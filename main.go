package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	language "cloud.google.com/go/language/apiv1"
	"cloud.google.com/go/language/apiv1/languagepb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gocolly/colly"
	db "github.com/sam-peach/rus_scaper/database"
)

func main() {
	ctx := context.Background()

	c := colly.NewCollector()
	pattern := "^/russia/.+"
	categoryPatterns := []string{
		"/Science/Engineering & Technology.*",
		"/Science/Computer Science.*",
		"/Law & Government/Military.*",
	}

	re := regexp.MustCompile(pattern)
	links := make([]string, 0)

	c.OnHTML("a", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		if re.MatchString(href) {
			links = append(links, href)
		}
	})

	c.Visit("https://www.rt.com/russia/")

	c.OnHTMLDetach("a")

	shouldSave := false
	textCollection := ""

	c.OnHTML("p", func(e *colly.HTMLElement) {

		client, err := language.NewClient(ctx)
		if err != nil {
			panic(err)
		}
		defer client.Close()

		text := e.Text
		textCollection += " " + text

		resp, err := client.ClassifyText(ctx, &languagepb.ClassifyTextRequest{
			ClassificationModelOptions: &languagepb.ClassificationModelOptions{
				ModelType: &languagepb.ClassificationModelOptions_V2Model_{},
			},
			Document: &languagepb.Document{
				Type:     languagepb.Document_PLAIN_TEXT,
				Source:   &languagepb.Document_Content{Content: text},
				Language: "en",
			},
		})
		if err != nil {
			panic(err)
		}

		for _, category := range resp.Categories {
			for _, pattern := range categoryPatterns {
				re := regexp.MustCompile(pattern)
				if re.MatchString(category.Name) {
					shouldSave = true
				}
			}
		}
	})

	fmt.Printf(">> Found %d links\n", len(links))

	for _, link := range links {
		textCollection = ""
		shouldSave = false
		completeLink := "https://www.rt.com" + link
		c.Visit(completeLink)

		if shouldSave {
			fmt.Println(">> Saving ", link)
			textCollection = strings.TrimSpace(textCollection)
			hasher := sha256.New()
			hasher.Write([]byte(textCollection))
			hashSum := hasher.Sum(nil)
			hashString := hex.EncodeToString(hashSum)

			now := time.Now().Unix()

			err := db.PutItem(ctx, map[string]types.AttributeValue{
				"Id":        &types.AttributeValueMemberS{Value: hashString},
				"Text":      &types.AttributeValueMemberS{Value: textCollection},
				"Country":   &types.AttributeValueMemberS{Value: "rus"},
				"Url":       &types.AttributeValueMemberS{Value: completeLink},
				"TimeStamp": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", now)},
			})

			if err != nil {
				panic(err)
			}
		}
	}

	os.Exit(0)
}
