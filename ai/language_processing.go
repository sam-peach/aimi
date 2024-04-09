package ai

import (
	"context"

	language "cloud.google.com/go/language/apiv1"
	"cloud.google.com/go/language/apiv1/languagepb"
)

var languageClient *language.Client
var ctx context.Context

func InitLanguageClient(ctx context.Context) {
	var err error
	languageClient, err = language.NewClient(ctx)
	if err != nil {
		panic(err)
	}
}

func CategoriesArticle(text string) []string {
	resp, err := languageClient.ClassifyText(ctx, &languagepb.ClassifyTextRequest{
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

	result := make([]string, len(resp.Categories))
	for _, category := range resp.Categories {
		result = append(result, category.Name)
	}

	return result
}
