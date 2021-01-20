package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/chromedp/chromedp"
)

func main() {
	// Setup browser
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.WindowSize(1920, 1080),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Get heatmap images
	var stockURL string
	if err := chromedp.Run(ctx, getStockMarketMap(&stockURL)); err != nil {
		log.Fatalln("Error getting stock market heatmap:", err)
		os.Exit(1)
	}

	// Update the database
	err := updateURL("Stock Market", stockURL)
	if err != nil {
		log.Fatalln("Error updating the URL in the database:", err)
		os.Exit(1)
	}
}

// getStockMarketMap gets the stock market heatmap image url
func getStockMarketMap(stockURL *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate("https://finviz.com/map.ashx?t=sec"),
		chromedp.WaitVisible(".chart", chromedp.ByQuery),
		chromedp.Click("//*[@id=\"share-map\"]", chromedp.NodeVisible),
		chromedp.Value("//*[@id=\"static\"]", stockURL, chromedp.NodeVisible),
	}
}

// updateURL calls updates the image url in DynamoDB
func updateURL(key string, url string) error {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := dynamodb.New(sess)
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":u": {
				S: aws.String(url),
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			"Name": {
				S: aws.String(key),
			},
		},
		TableName:        aws.String("Stonks_Heatmaps"),
		UpdateExpression: aws.String("SET Image = :u"),
	}

	_, err := svc.UpdateItem(input)
	if err != nil {
		return err
	}

	return nil
}
