package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func getEpicStoreCookie(ctx context.Context) {
	fmt.Println("Logging into Epic Games Store.")
	tries := 10
	for i := 0; i < tries; i++ {
		fmt.Printf("Trying to log in to Epic Games Store... %d of %d\n", i+1, tries)
		if err := callWithTimeout(ctx, chromedp.Navigate(`https://www.epicgames.com/id/login/epic`), longTimeout); err != nil {
			log.Println("Couldnt navigate to login page.")
			continue
		}
		if err := callWithTimeout(ctx, chromedp.WaitEnabled(`//input[@id='email']`), timeOut); err == nil {
			chromedp.SendKeys(`//input[@id='email']`, config.Username).Do(ctx)
		} else {
			log.Println("Could not find email field.")
			continue
		}
		if err := callWithTimeout(ctx, chromedp.WaitEnabled(`//input[@id='password']`), timeOut); err == nil {
			chromedp.SendKeys(`//input[@id='password']`, config.Password).Do(ctx)
		} else {
			log.Println("Could not find password field.")
			continue
		}
		if err := callWithTimeout(ctx, chromedp.WaitEnabled(`//button[@id='sign-in']`), timeOut); err == nil {
			callWithTimeout(ctx, chromedp.Click(`//button[@id='sign-in']`), timeOut)
		} else {
			log.Print("Could not find sign in button.")
			continue
		}
		if success := handle2FA(ctx); !success {
			continue
		}
		// Wait for 10 seconds to check if we're logged in
		chromedp.Sleep(time.Second * 10).Do(ctx)
		if _, epicStoreCookie := checkCookies(ctx); epicStoreCookie {
			fmt.Println("Logged into Epic Games Store.")
			return
		}
	}
	log.Println("Apparently, logging in is not successful. Too bad.")
	time.Sleep(time.Second * 5)
	var outer string
	fmt.Println("Dumping DOM...")
	chromedp.OuterHTML("//html", &outer).Do(ctx)
	fmt.Println(outer)
	if len(config.ImgurClientID) > 0 {
		log.Printf("Link to screenshot: %s", screenshot(ctx))
	}
	log.Fatal("Exiting.")
}

func getAccessibilityCookie(ctx context.Context) {
	tries := 5
	for _, link := range config.HCaptchaURLs {
		for i := 0; i < tries; i++ {
			fmt.Printf("Trying to find hCaptcha cookie button... %d of %d\n", i+1, tries)
			chromedp.Navigate(link).Do(ctx)
			if err := callWithTimeout(ctx, chromedp.WaitEnabled(`//button[@data-cy='setAccessibilityCookie']`), timeOut); err == nil {
				break
			}
		}
		for i := 0; i < tries; i++ {
			fmt.Printf("Trying to get hCaptcha accessibility cookie... %d of %d\n", i+1, tries)
			chromedp.Sleep(time.Millisecond * time.Duration(rand.Intn(2500)+2500)).Do(ctx)
			chromedp.Click(`//button[@data-cy='setAccessibilityCookie']`).Do(ctx)
			chromedp.Sleep(time.Millisecond * time.Duration(rand.Intn(2500)+2500)).Do(ctx)
			if accessibilityCookie, _ := checkCookies(ctx); accessibilityCookie {
				fmt.Println("Acquired hCaptcha cookie!")
				return
			}
		}
	}
	sendTelegramMessage("Couldn't get hCaptcha cookie...", yoinkFailure)
	log.Fatal("Couldn't get accessibility cookie from hCaptcha, so cannot bypass captcha. Try again another time.")
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
		fmt.Println("Existing cookie found for Epic Games Store. Doing nothing.")
		return
	}
	// This was necessary before, for some reason it isn't anymore?
	// If hCaptcha 2FA is presented, uncomment this.
	if !accessibilityCookie {
		// fmt.Println("Need to bypass hCaptcha to login to Epic Games Store.")
		// getAccessibilityCookie(ctx)
	}
	getEpicStoreCookie(ctx)
}

func setCookies(ctx context.Context) {
	fmt.Println("Setting cookies.")
	expiryTime := cdp.TimeSinceEpoch(time.Now().Add(time.Hour))
	cookies := []*network.CookieParam{
		{
			Name:    "OptanonAlertBoxClosed",
			Value:   "en-US",
			URL:     "epicgames.com",
			Expires: &expiryTime,
		},
	}
	network.SetCookies(cookies).Do(ctx)
	// Not sure if necessary but seems to increase changes of success.
	if err := callWithTimeout(ctx, chromedp.WaitEnabled(`//button[contains(@class, "onetrust-lg")]`), timeOut); err == nil {
		fmt.Println("Removed cookiebanner.")
		chromedp.Click(`//button[contains(@class, "onetrust-lg")]`).Do(ctx)
	}
}
