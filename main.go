package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/chromedp/cdproto/cdp"
	l "github.com/chromedp/cdproto/log"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/pquerna/otp/totp"

	img "cdp/imgur"
)

var config Config

func callWithTimeout(c context.Context, a chromedp.QueryAction, seconds time.Duration) error {
	ctx, cancel := context.WithTimeout(c, seconds)
	defer cancel()
	err := a.Do(ctx)
	return err
}

func handleFreeGames(c context.Context, urls []string) {
	fmt.Printf("Handling %d games!\n", len(urls))
	for _, url := range urls {
		// Sometimes we receive forking 403... try this a couple of times too
		for i := 0; i < 5; i++ {
			fmt.Printf("Checking URL: %s\n - try %d out of %d...", url, i+1, 5)
			if err := callWithTimeout(c, chromedp.Navigate(url), longTimeout); err != nil {
				log.Printf("Received error on navigating to %s: %s", url, err.Error())
				continue
			}
			if err := callWithTimeout(c, chromedp.WaitEnabled(`//button[text()[contains(.,"Continue")]]`), timeOut); err == nil {
				callWithTimeout(c, chromedp.Click(`//button[text()[contains(.,"Continue")]]`), timeOut)
			} else {
				continue
			}
			fmt.Println("Checking if already owned.")
			if err := callWithTimeout(c, chromedp.WaitVisible(`//button[./span[text()[contains(.,"Owned")]]]`), timeOut); err == nil {
				fmt.Println("Already owned. Skipping.")
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
			if err := callWithTimeout(c, chromedp.WaitEnabled(`//span[text()="Thank you for buying"]`), timeOut); err == nil {
				log.Println("Claiming appears to be succesful.")
				sendTelegramMessage(url, yoinkSuccess)
			}
		}
	}
}

const (
	yoinkSuccess = iota
	yoinkFailure = iota
)

func sendTelegramMessage(url string, status int) {
	if !(len(config.TelegramID) > 0) {
		return
	}
	tgParamsJSON, _ := json.Marshal(TelegramPost{ID: config.TelegramID, URL: url, Status: status})
	res, err := http.Post("https://epic-games-yoinker-api.azurewebsites.net/message/send", "application/json", bytes.NewBuffer(tgParamsJSON))
	if err != nil {
		log.Println(err)
		return
	}
	body, _ := ioutil.ReadAll(res.Body)
	log.Println(string(body))
}

type TelegramPost struct {
	ID     string `json:"Id"`
	URL    string `json:"Url"`
	Status int    `json:"Status"`
}

func getFreeGameURLs(ctx context.Context) (urls []string) {
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
			sendTelegramMessage("Couldn't find a free game...", yoinkFailure)
		}
		if len(urls) > 0 {
			return
		}
	}

	log.Fatal("Couldn't find more than 0 urls.")
	return
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
		// getAccessibilityCookie(ctx)
	}
	getEpicStoreCookie(ctx)
}

func setupLogger(ctx context.Context) {
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		go func() {
			if e, ok := ev.(*runtime.EventConsoleAPICalled); ok {
				for _, arg := range e.Args {
					if arg.Type != runtime.TypeUndefined {
						// if e.Type == runtime.APITypeError && arg.Type != runtime.TypeUndefined {
						fmt.Printf("Console Entry: %s\n", arg.Value)
					}
				}
			}
			if e, ok := ev.(*l.EventEntryAdded); ok {
				fmt.Printf("Console Log Entry: %s\n", e.Entry.Text)
			}
		}()
	})
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
	if err := callWithTimeout(ctx, chromedp.WaitEnabled(`//button[contains(@class, "onetrust-lg")]`), timeOut); err == nil {
		fmt.Println("Removed stoopid cookiebanner.")
		chromedp.Click(`//button[contains(@class, "onetrust-lg")]`).Do(ctx)
	}
	network.SetCookies(cookies).Do(ctx)
}

func screenshot(ctx context.Context) string {
	// var buf []byte
	// chromedp.CaptureScreenshot(&buf).Do(ctx)
	// if err := ioutil.WriteFile(screenshotName, buf, os.FileMode(0755)); err != nil {
	// log.Println("Screenshot failed.")
	// return ""
	// }
	screenshotName := "screenshot.png"
	cmd := exec.Command("/usr/bin/scrot", screenshotName)
	cmd.Env = append(os.Environ(),
		"DISPLAY=:99",
	)
	err := cmd.Run()
	if err != nil {
		log.Printf("Screenshotting failed: %s", err)
	}
	url, err := img.Upload(screenshotName)
	if err != nil {
		log.Println(err)
		return ""
	}
	return url
}

const timeOut = time.Second * 10
const longTimeout = time.Second * 30

func main() {
	config = readConfig()
	dir, err := ioutil.TempDir("", "free-game-fetcher-2000")
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
		chromedp.Flag("start-maximized", true),
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
