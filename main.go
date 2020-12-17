package main

import (
	"bytes"
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/chromedp/chromedp"
)

func main() {
	// Create options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.WindowSize(1000, 1000),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// create context
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Take screenshot
	var buf []byte
	if err := chromedp.Run(ctx, screenshot(`https://finviz.com/map.ashx?t=sec`, `.chart`, &buf)); err != nil {
		log.Fatal(err)
	}

	// Upload image
	if err := uploadImage(buf); err != nil {
		log.Fatalf("Failed to upload heatmap image: %v", err)
	}
}

// screenshot takes a screenshot of a specific element.
func screenshot(urlstr, sel string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.WaitVisible(sel, chromedp.ByQuery),
		chromedp.Screenshot(sel, res, chromedp.NodeVisible, chromedp.ByQuery),
	}
}

// uploadImage uploads image to the heatmap S3 bucket
func uploadImage(buf []byte) error {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	uploader := s3manager.NewUploader(sess)

	bucket := "stockbot-heatmap"
	keyName := "marketHeatmap.png"
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
