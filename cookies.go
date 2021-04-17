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

func getAccessibilityCookie(ctx context.Context, user User) {
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
	sendTelegramMessage("Couldn't get hCaptcha cookie...", yoinkFailure, user)
	log.Fatal("Couldn't get accessibility cookie from hCaptcha, so cannot bypass captcha. Try again another time.")
}

func checkCookies(ctx context.Context) (accessibilityCookie bool, epicCookie bool) {
	siteCookies, _ := network.GetCookies().Do(ctx)
	for _, cookie := range siteCookies {
		// log.Printf("Name: %s and Value: %s", cookie.Name, cookie.Value)
		if cookie.Name == "EPIC_SSO" {
			epicCookie = true
		}
	}
	allCookies, _ := network.GetAllCookies().Do(ctx)
	for _, cookie := range allCookies {
		// log.Printf("Name: %s and Value: %s", cookie.Name, cookie.Value)
		if cookie.Name == "hc_accessibility" {
			accessibilityCookie = true
		}
	}
	return accessibilityCookie, epicCookie
}

func getCookies(ctx context.Context, user User) {
	accessibilityCookie, epicGamesCookie := checkCookies(ctx)
	if epicGamesCookie {
		fmt.Println("Existing cookie found for Epic Games Store. Doing nothing.")
		return
	}
	// This was necessary before, for some reason it isn't anymore?
	// If hCaptcha 2FA is presented, uncomment this.
	if !accessibilityCookie {
		fmt.Println("Need to bypass hCaptcha to login to Epic Games Store.")
		getAccessibilityCookie(ctx, user)
	}
	loginEpicGames(ctx, user)
}

func setCookiesMisc(ctx context.Context) {
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
