package heatmap

import (
	"context"
	"io/ioutil"
	"log"

	"github.com/chromedp/chromedp"
)

// GetHeatMap saves a screenshot of the market heatmap from finviz.com
func GetHeatMap() {
	// Create options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.WindowSize(1000, 1000),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// create context
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var buf []byte
	if err := chromedp.Run(ctx, screenshot(`https://finviz.com/map.ashx?t=sec`, `.chart`, &buf)); err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile("./heatmap/screenshot.png", buf, 0644); err != nil {
		log.Fatal(err)
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
