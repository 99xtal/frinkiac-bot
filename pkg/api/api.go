package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

var client = &http.Client{}

//Frames Holds data about a Frinkiac (or Morbotron) search
type Frame struct {
	ID        int    `json:"Id"`
	Episode   string `json:"Episode"`
	Timestamp int    `json:"Timestamp"`
}

func (f *Frame) GetPhotoUrl() string {
	return fmt.Sprintf("http://frinkiac.com/img/%s/%d.jpg", f.Episode, f.Timestamp)
}

type FrinkiacClient struct {}

func (f *FrinkiacClient) Search(query string) ([]*Frame, error) {
	var info []*Frame
	req, err := http.NewRequest("GET", "https://frinkiac.com/api/search?q="+url.QueryEscape(query), nil)
	if err != nil {
		return info, err
	}
	req.Header.Set("User-Agent", "Frinkiac_Api_Go/0.1")
	resp, err := client.Do(req)
	if err != nil {
		return info, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return info, err
	}
	json.Unmarshal(body, &info)
	return info, nil
}

func NewFrinkiacClient() *FrinkiacClient {
	return &FrinkiacClient{}
}