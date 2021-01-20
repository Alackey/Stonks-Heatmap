package main

import (
	"context"
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

	// Stock Market Heatmap
	var stockURL string
	if err := chromedp.Run(ctx, getStockMarketHeatMap(&stockURL)); err != nil {
		log.Fatalln("Error getting stock market heatmap:", err)
		os.Exit(1)
	}

	err := updateURL("Stock Market", stockURL)
	if err != nil {
		log.Fatalln("Error updating the Stock Market URL in the database:", err)
		os.Exit(1)
	}

	// Crypto Heatmap
	var ok bool
	var cryptoURL string
	if err := chromedp.Run(ctx, getCryptoHeatMap(&cryptoURL, &ok)); err != nil {
		log.Fatalln("Error getting stock market heatmap:", err)
		os.Exit(1)
	} else if !ok {
		log.Fatalln("Error getting the Coin360.com heatmap url from the share popup")
		os.Exit(1)
	}

	err = updateURL("Crypto", cryptoURL)
	if err != nil {
		log.Fatalln("Error updating the Crypto URL in the database:", err)
		os.Exit(1)
	}
}

// getStockMarketHeatMap gets the stock market heatmap image url
func getStockMarketHeatMap(url *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate("https://finviz.com/map.ashx?t=sec"),
		chromedp.WaitVisible(".chart", chromedp.ByQuery),
		chromedp.Click("//*[@id=\"share-map\"]", chromedp.NodeVisible),
		chromedp.Value("//*[@id=\"static\"]", url, chromedp.NodeVisible),
	}
}

// getCryptoHeatMap gets the stock market heatmap image url
func getCryptoHeatMap(url *string, ok *bool) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate("https://coin360.com/"),
		chromedp.WaitVisible("//*[@id=\"MAP_ID\"]"),
		chromedp.Click("//*[@id=\"app\"]/section/div[1]/section/div[2]/div[3]/div[1]/div", chromedp.NodeVisible),
		chromedp.AttributeValue("/html/body/section/section/div[4]/div/div[4]/div[3]/a", "href", url, ok, chromedp.NodeVisible),
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
