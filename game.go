package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

func handleFreeGames(c context.Context, urls []string, user User) {
	fmt.Printf("Handling %d games!\n", len(urls))
	for _, url := range urls {
		claimSuccessful := false
		// Sometimes we receive forking 403... try this a couple of times too
		for i := 0; i < 5; i++ {
			fmt.Printf("Checking URL: %s - try %d out of %d...\n", url, i+1, 5)
			if err := callWithTimeout(c, chromedp.Navigate(url), longTimeout); err != nil {
				log.Printf("Received error on navigating to %s: %s", url, err.Error())
				continue
			}
			if err := callWithTimeout(c, chromedp.WaitEnabled(`//button[text()[contains(.,"Continue")]]`), shortTimeOut); err == nil {
				callWithTimeout(c, chromedp.Click(`//button[text()[contains(.,"Continue")]]`), timeOut)
			}
			fmt.Println("Checking if already owned.")
			if err := callWithTimeout(c, chromedp.WaitVisible(`//button[./span[text()[contains(.,"Owned")]]]`), shortTimeOut); err == nil {
				fmt.Println("Already owned. Skipping.")
				claimSuccessful = true
				break
			}
			fmt.Println("Waiting for GET button")
			if err := callWithTimeout(c, chromedp.WaitEnabled(`//button[.//text()[contains(.,"Get")]]`), timeOut); err == nil {
				fmt.Println("Clicking GET button")
				chromedp.Click(`//button[.//text()[contains(.,"Get")]]`).Do(c)
			} else {
				fmt.Println("Could not find the GET button.")
				continue
			}
			fmt.Println("Waiting for Place Order")
			if err := callWithTimeout(c, chromedp.WaitEnabled(`//button[./span[text()[contains(.,"Place Order")]]]`), timeOut); err == nil {
				fmt.Println("Clicking Place Order")
				chromedp.Click(`//button[./span[text()[contains(.,"Place Order")]]]`).Do(c)
			} else {
				fmt.Println("Could not find the Place Order button.")
				continue
			}
			fmt.Println("Waiting for Agreement")
			if err := callWithTimeout(c, chromedp.WaitEnabled(`//button[span[text()="I Agree"]]`), timeOut); err == nil {
				fmt.Println("Clicking I Agree")
				chromedp.Click(`//button[span[text()="I Agree"]]`).Do(c)
			} else {
				fmt.Println("Could not find the 'I Agree' button.")
				continue
			}
			if err := callWithTimeout(c, chromedp.WaitEnabled(`//span[text()="Thank you for buying"]`), longTimeout); err == nil {
				log.Println("Claiming appears to be succesful.")
				claimSuccessful = true
				sendTelegramMessage(url, yoinkSuccess, user)
				break
			}
		}
		if !claimSuccessful {
			log.Printf("Claiming %s unsuccessful.\n", url)
			log.Printf("Link to screenshot: %s", screenshot(c))
		}
	}
}

func getFreeGameURLs(ctx context.Context, user User) (urls []string) {
	tries := 5
	for i := 0; i < tries; i++ {
		if err := callWithTimeout(ctx, chromedp.Navigate("https://www.epicgames.com/store/en-US/free-games"), timeOut); err != nil {
			log.Println("Could not navigate to free games page.")
			log.Printf("Link to screenshot: %s", screenshot(ctx))
		}
		if err := callWithTimeout(ctx, chromedp.WaitVisible(`//a[.//text()[starts-with(.,"Free Now")]]`), timeOut); err != nil {
			log.Println("Could not find urls with text that start with 'Free Now'.")
			log.Printf("Link to screenshot: %s", screenshot(ctx))
		}
		if err := callWithTimeout(ctx,
			chromedp.QueryAfter(`//a[.//text()[starts-with(.,"Free Now")]]`, func(ctx context.Context, bla runtime.ExecutionContextID, nodes ...*cdp.Node) error {
				if len(nodes) < 1 {
					return errors.New("no free games found")
				}
				for _, node := range nodes {
					url, _ := node.Attribute("href")
					urls = append(urls, "https://www.epicgames.com"+url)
				}
				return nil
			}), timeOut); err != nil {
			log.Println(err.Error())
			log.Printf("Link to screenshot: %s", screenshot(ctx))
			sendTelegramMessage("Couldn't find a free game...", yoinkFailure, user)
		}
		if len(urls) > 0 {
			return
		}
	}

	log.Fatal("Couldn't find more than 0 urls.")
	return
}
