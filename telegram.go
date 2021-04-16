package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	yoinkSuccess = iota
	yoinkFailure = iota
)

type TelegramPost struct {
	ID     string `json:"Id"`
	URL    string `json:"Url"`
	Status int    `json:"Status"`
}

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
