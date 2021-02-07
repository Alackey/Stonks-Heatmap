package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/chromedp/chromedp"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", GetHeatmap)
	log.Println("Listening on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// GetHeatmap gets the heatmaps for stocks and crypto
func GetHeatmap(w http.ResponseWriter, r *http.Request) {
	// Setup browser
	log.Println("Creating browser")
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.WindowSize(1920, 1080),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Stock Market Heatmap
	if isMarketHours() {
		err := saveStockMarketHeatMap(ctx)
		if err != nil {
			log.Println("Error getting stock market heatmap:", err)
		}
	}

	// Crypto Heatmap
	err := saveCryptoHeatMap(ctx)
	if err != nil {
		log.Println("Error getting crypto heatmap:", err)
	}

	fmt.Println("Finished getting heatmaps")
	w.WriteHeader(http.StatusOK)
}

// isMarketHours checks if the market is open at the current time
func isMarketHours() bool {
	// Market hours: 9am - 11:59pm UTC and 12am UTC next day
	currentTime := time.Now().UTC()
	currentHour := currentTime.Hour()

	if (currentTime.Weekday() != time.Saturday && currentTime.Weekday() != time.Sunday && (currentHour == 0 || (9 <= currentHour && currentHour <= 23))) ||
		(currentTime.Weekday() == time.Saturday && currentHour == 0) {
		return true
	}
	return false
}

// saveStockMarketHeatMap gets the stock market heatmap
func saveStockMarketHeatMap(ctx context.Context) error {
	log.Println("Getting stock market heatmap")
	var buf []byte
	if err := chromedp.Run(ctx, screenshot(`https://finviz.com/map.ashx?t=sec`, `.chart`, &buf, chromedp.ByQuery)); err != nil {
		return err
	}

	log.Println("Uploading stock market heatmap")
	if err := uploadImage("marketHeatmap.png", buf); err != nil {
		return err
	}
	return nil
}

// saveCryptoHeatMap gets the stock market heatmap image url
func saveCryptoHeatMap(ctx context.Context) error {
	log.Println("Getting crypto heatmap")
	var buf []byte
	if err := chromedp.Run(ctx, screenshot("https://coin360.com/", "#MAP_ID", &buf, chromedp.ByID)); err != nil {
		return err
	}

	log.Println("Uploading crypto heatmap")
	if err := uploadImage("cryptoHeatmap.png", buf); err != nil {
		return err
	}
	return nil
}

// screenshot takes a screenshot of a specific element.
func screenshot(urlstr, sel string, res *[]byte, opt chromedp.QueryOption) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.WaitVisible(sel, chromedp.ByQuery),
		chromedp.Screenshot(sel, res, chromedp.NodeVisible, opt),
	}
}

// uploadImage uploads image to the heatmap S3 bucket
func uploadImage(keyName string, buf []byte) error {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	uploader := s3manager.NewUploader(sess)

	bucket := "stockbot-heatmap"
	body := bytes.NewReader(buf)

	upParams := &s3manager.UploadInput{
		Bucket: &bucket,
		Key:    &keyName,
		Body:   body,
	}

	_, err := uploader.Upload(upParams)
	if err != nil {
		return err
	}

	return nil
}
