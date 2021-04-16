package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"

	img "cdp/imgur"
)

func callWithTimeout(c context.Context, a chromedp.QueryAction, seconds time.Duration) error {
	ctx, cancel := context.WithTimeout(c, seconds)
	defer cancel()
	err := a.Do(ctx)
	return err
}

func screenshot(ctx context.Context) string {
	screenshotName := "screenshot.png"
	var buf []byte
	chromedp.CaptureScreenshot(&buf).Do(ctx)
	if err := ioutil.WriteFile(screenshotName, buf, os.FileMode(0755)); err != nil {
		log.Println("Screenshot failed.")
		return ""
	}
	// cmd := exec.Command("/usr/bin/scrot", screenshotName)
	// cmd.Env = append(os.Environ(),
	// 	"DISPLAY=:99",
	// )
	// err := cmd.Run()
	// if err != nil {
	// 	log.Printf("Screenshotting failed: %s", err)
	// }
	url, err := img.Upload(screenshotName)
	if err != nil {
		log.Println(err)
		return ""
	}
	return url
}

const shortTimeOut = time.Second * 3
const timeOut = time.Second * 10
const longTimeout = time.Second * 30

func main() {
	config = readConfig()
	dir, err := ioutil.TempDir("", "epic-games-yoinker")
	if err != nil {
		log.Fatalf("Could not create user data dir for Chrome in %s\n", dir)
	}
	// dir := "~/.config/google-chrome"
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.UserDataDir(dir),
		chromedp.DisableGPU,
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
	}
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()
	if err := chromedp.Run(taskCtx); err != nil {
		log.Fatalf("Could not start Chrome?\n%s\n", err)
	}
	// Authenticate for imgur to be able to create an image
	if len(config.ImgurClientID) > 0 {
		img.Authenticate(config.ImgurClientID, config.ImgurSecret, config.ImgurRefreshToken)
	}
	if err := chromedp.Run(taskCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			setupLogger(ctx)
			getCookies(ctx)
			setCookies(ctx)
			handleFreeGames(ctx, getFreeGameURLs(ctx))
			fmt.Println("Done!")
			return nil
		}),
	); err != nil {
		log.Fatal(err)
	}
}
