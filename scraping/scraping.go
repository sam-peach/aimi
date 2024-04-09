package scraping

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gocolly/colly"
	"github.com/sam-peach/rus_scaper/ai"
	"github.com/sam-peach/rus_scaper/database"
)

const urlBase string = "https://www.rt.com"

func ProcessLink(ctx context.Context, link string, scraper *colly.Collector) {
	shouldSave := false
	textCollection := ""
	completeLink := urlBase + link

	scraper.OnHTML("p", func(e *colly.HTMLElement) {
		text := e.Text
		textCollection += " " + text
	})

	scraper.Visit(completeLink)
	shouldSave = ai.IsArticleRelevant(textCollection)

	fmt.Println(">> Link:", link)
	fmt.Println(">> Saving?:", shouldSave)

	if shouldSave {
		textCollection = strings.TrimSpace(textCollection)
		classification, _ := ai.Categorize(textCollection)
		hasher := sha256.New()
		hasher.Write([]byte(textCollection))
		hashSum := hasher.Sum(nil)
		hashString := hex.EncodeToString(hashSum)

		now := time.Now().Unix()

		fmt.Println(">> Primary Category ", classification.PrimaryCategory)
		fmt.Println(">> Secondary Category ", classification.SecondaryCategory)

		database.PutItem(ctx, map[string]types.AttributeValue{
			"Id":                &types.AttributeValueMemberS{Value: hashString},
			"Text":              &types.AttributeValueMemberS{Value: textCollection},
			"Country":           &types.AttributeValueMemberS{Value: "rus"},
			"Url":               &types.AttributeValueMemberS{Value: completeLink},
			"TimeStamp":         &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", now)},
			"PrimaryCategory":   &types.AttributeValueMemberS{Value: classification.PrimaryCategory.Key},
			"PrimaryScore":      &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", classification.PrimaryCategory.Relevance)},
			"SecondaryCategory": &types.AttributeValueMemberS{Value: classification.SecondaryCategory.Key},
			"SecondaryScore":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", classification.SecondaryCategory.Relevance)},
		})
	}
}
