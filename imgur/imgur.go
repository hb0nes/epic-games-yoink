package imgur

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var accessToken string

func Authenticate(clientID string, clientSecret string, refreshToken string) {
	if len(accessToken) > 0 {
		log.Println("Already authenticated.")
	}
	vals := url.Values{
		"client_id":     []string{"48ada7f7db57cfa"},
		"client_secret": []string{"4bfd5e32b7f08294285237d3210f48eb65bb644b"},
		"grant_type":    []string{"refresh_token"},
		"refresh_token": []string{refreshToken},
	}
	res, err := http.PostForm(`https://api.imgur.com/oauth2/token`, vals)
	if err != nil {
		log.Println(err)
	}
	r, err := ioutil.ReadAll(res.Body)
	var dec map[string]interface{}
	err = json.Unmarshal(r, &dec)
	if err != nil {
		log.Println("Could not decode results.")
	}
	var ok bool
	accessToken, ok = dec["access_token"].(string)
	if !ok {
		log.Println("Could not decode results.")
	}
}

func Upload(path string) (string, error) {
	if len(accessToken) < 1 {
		return "", fmt.Errorf("not authenticated yet")
	}
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("Could not find file at %s.", path)
	}
	vals := url.Values{
		"image": []string{base64.StdEncoding.EncodeToString(fileBytes)},
	}
	vals.Set("type", "base64")
	vals.Set("name", "test")
	vals.Set("title", "test")
	req, _ := http.NewRequest(http.MethodPost, "https://api.imgur.com/3/upload", strings.NewReader(vals.Encode()))
	req.Header.Add("Authorization", accessToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(vals.Encode())))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Uploading failed: %s.", err.Error())
	}
	resRead, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var resDec map[string]interface{}
	json.Unmarshal(resRead, &resDec)
	imgData, ok := resDec["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("could not retrieve image data: %s", resDec)
	}
	imgURL, ok := imgData["link"].(string)
	if !ok {
		return "", fmt.Errorf("could not retrieve image URL: %s", imgData)
	}
	return imgURL, nil
}
