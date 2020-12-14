package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

var config Config

func doSlowCall(c context.Context, a chromedp.QueryAction) error {
	ctx, cancel := context.WithTimeout(c, time.Second*2)
	defer cancel()
	err := a.Do(ctx)
	return err
}

// func handleFreeGames(c context.Context, nodes []*cdp.Node) {
func handleFreeGames(c context.Context, urls []string) {
	fmt.Printf("Handling %d games!\n", len(urls))
	for _, url := range urls {
		fmt.Printf("Checking URL: %s\n", url)
		chromedp.Navigate(url).Do(c)
		doSlowCall(c, chromedp.WaitVisible(`//button[text()[contains(.,"Continue")]]`))
		doSlowCall(c, chromedp.Click(`//button[text()[contains(.,"Continue")]]`))
		fmt.Println("Checking if already owned.")
		// doSlowCall(c, chromedp.WaitVisible(`//button[./span[text()[contains(.,"Owned")]]]`))
		err := doSlowCall(c, chromedp.QueryAfter(`//button[./span[text()[contains(.,"Owned")]]]`, func(a context.Context, b runtime.ExecutionContextID, nodes ...*cdp.Node) error {
			return errors.New("already owned")
		}))
		if err != nil && err.Error() == "already owned" {
			chromedp.NavigateBack().Do(c)
			fmt.Println("Already owned, continuing.")
			continue
		}
		fmt.Println("Waiting for GET button")
		if err := doSlowCall(c, chromedp.WaitEnabled(`//button[@data-testid="purchase-cta-button"]`)); err == nil {
			fmt.Println("Clicking GET button")
			chromedp.Click(`//button[@data-testid="purchase-cta-button"]`).Do(c)
		}
		fmt.Println("Waiting for checkbox")
		if err := doSlowCall(c, chromedp.WaitEnabled(`//i[@class="icon-checkbox-unchecked radio-unchecked"]`)); err == nil {
			fmt.Println("Clicking checkbox")
			chromedp.Click(`//i[@class="icon-checkbox-unchecked radio-unchecked"]`).Do(c)
		}
		fmt.Println("Waiting for Place Order")
		if err := doSlowCall(c, chromedp.WaitEnabled(`//button[./span[text()[contains(.,"Place Order")]]]`)); err == nil {
			fmt.Println("Clicking Place Order")
			chromedp.Click(`//button[./span[text()[contains(.,"Place Order")]]]`).Do(c)
		}
		fmt.Println("Waiting for Agreement")
		if err := doSlowCall(c, chromedp.WaitEnabled(`//button[span[text()="I Agree"]]`)); err == nil {
			fmt.Println("Clicking I Agree")
			chromedp.Click(`//button[span[text()="I Agree"]]`).Do(c)
		}
		doSlowCall(c, chromedp.WaitEnabled(`//span[text()="Thank you for buying"]`))
	}
}

func getFreeGameURLs(ctx context.Context) (urls []string) {
	chromedp.Run(ctx,
		chromedp.Navigate("https://www.epicgames.com/store/en-US/free-games"),
		chromedp.WaitVisible(`//a[@aria-label[contains(., "Free Games")]]`),
		chromedp.Sleep(time.Second*5),
		chromedp.QueryAfter(`//a[@aria-label[contains(., "Free Games")]]`, func(ctx context.Context, bla runtime.ExecutionContextID, nodes ...*cdp.Node) error {
			if len(nodes) < 1 {
				return errors.New("expected at least one node")
			}
			for _, node := range nodes {
				url, _ := node.Attribute("href")
				urls = append(urls, "https://www.epicgames.com"+url)
			}
			return nil
		}))
	return urls
}

func getAccessibilityCookie(ctx context.Context) {
	tries := 2
	for _, link := range config.HCaptchaURLs {
		chromedp.Navigate(link).Do(ctx)
		chromedp.WaitEnabled(`//button[@data-cy='setAccessibilityCookie']`).Do(ctx)
		for i := 0; i < tries; i++ {
			chromedp.Sleep(time.Millisecond * time.Duration(rand.Intn(2500)+2500)).Do(ctx)
			chromedp.Click(`//button[@data-cy='setAccessibilityCookie']`).Do(ctx)
			chromedp.Sleep(time.Millisecond * time.Duration(rand.Intn(2500)+2500)).Do(ctx)
			if accessibilityCookie, _ := checkCookies(ctx); accessibilityCookie {
				fmt.Println("Acquired hCaptcha cookie!")
				return
			}
		}
	}
	log.Fatal("Couldn't get accessibility cookie from hCaptcha, so cannot bypass captcha. Try again another time.")
}

func getEpicStoreCookie(ctx context.Context) {
	chromedp.Run(ctx,
		chromedp.Navigate(`https://www.epicgames.com/login`),
		chromedp.WaitEnabled(`//div[@id='login-with-epic']`),
		chromedp.Click(`//div[@id='login-with-epic']`),
		chromedp.WaitEnabled(`//input[@id='email']`),
		chromedp.SendKeys(`//input[@id='email']`, config.Username),
		chromedp.WaitEnabled(`//input[@id='password']`),
		chromedp.SendKeys(`//input[@id='password']`, config.Password),
		chromedp.WaitEnabled(`//button[@id='sign-in']`),
		chromedp.Click(`//button[@id='sign-in']`),
	)
	chromedp.Sleep(time.Millisecond * time.Duration(rand.Intn(2500)+2500)).Do(ctx)
	if _, epicStoreCookie := checkCookies(ctx); !epicStoreCookie {
		log.Fatal("Apparently logging in is not successful. Too bad :(.")
	} else {
		fmt.Println("Logged into Epic Games Store.")
	}
}

func checkCookies(ctx context.Context) (accessibilityCookie bool, epicCookie bool) {
	cookies, err := network.GetAllCookies().Do(ctx)
	if err != nil {
		panic(err)
	}
	for _, cookie := range cookies {
		if cookie.Name == "hc_accessibility" {
			accessibilityCookie = true
		}
		if cookie.Name == "EPIC_SSO" {
			epicCookie = true
		}
	}
	return accessibilityCookie, epicCookie
}

func getCookies(ctx context.Context) {
	accessibilityCookie, epicGamesCookie := checkCookies(ctx)
	if epicGamesCookie {
		fmt.Println("Existing cookie found for Epic Games Store. Doing nothing =].")
		return
	}
	if !accessibilityCookie {
		fmt.Println("Need to bypass hCaptcha to login to Epic Games Store.")
		getAccessibilityCookie(ctx)
	}
	getEpicStoreCookie(ctx)
}

func main() {
	config = readConfig()
	dir, err := ioutil.TempDir("", "free-game-fetcher-2000")
	if err != nil {
		log.Fatalf("Could not create user data dir for Chrome in %s\n", dir)
	}
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.UserDataDir(dir),
		chromedp.DisableGPU,
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("start-maximized", true),
	}
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()
	if err := chromedp.Run(taskCtx); err != nil {
		log.Fatalf("Could not start Chrome?\n%s\n", err)
	}
	if err := chromedp.Run(taskCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			getCookies(ctx)
			handleFreeGames(ctx, getFreeGameURLs(ctx))
			fmt.Println("Done!")
			return nil
		}),
	); err != nil {
		log.Fatal(err)
	}
}
