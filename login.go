package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/pquerna/otp/totp"
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

func handle2FA(ctx context.Context) (success bool) {
	// If OTP/2FA is enabled, fill in the code
	err := callWithTimeout(ctx, chromedp.WaitEnabled(`//input[@id='code']`), timeOut)
	if err == nil && len(config.OTPSecret) < 32 {
		log.Fatal("It appears 2FA is enabled for this account but the OTP Secret hasn't been configured in the configuration.")
	}
	if err != nil {
		return true
	}
	code, err := totp.GenerateCode(config.OTPSecret, time.Now())
	fmt.Println("OTP Password is " + code)
	if err != nil {
		log.Fatal("OTPSecret configured but cannot derive code from it. Double check the config.")
		return false
	}
	chromedp.SendKeys(`//input[@id='code']`, code).Do(ctx)
	time.Sleep(time.Second)
	err = callWithTimeout(ctx, chromedp.WaitEnabled(`//button[@id='continue']`), timeOut)
	if err == nil {
		chromedp.Click(`//button[@id='continue']`).Do(ctx)
		log.Println("Clicked Continue button in 2FA process.")
		return true
	}
	log.Println("Something went wrong inputting the 2FA code.")
	return false
}
