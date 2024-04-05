package database

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var client *dynamodb.Client
var tableName string = "ScrapedData"

func init() {
	ctx := context.Background()

	config, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("unable to load SDK config " + err.Error())
	}

	config.Region = "us-east-1"

	client = dynamodb.NewFromConfig(config)
}

func Client() *dynamodb.Client {
	return client
}

func PutItem(ctx context.Context, item map[string]types.AttributeValue) error {
	_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &tableName,
		Item:      item,
	})

	return err
}
