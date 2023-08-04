package api

import (
	"encoding/base64"
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

type Episode struct {
	ID	int	`json:"Id"`
	Key	string	`json:"Key"`
	Season	int `json:"Season"`
	EpisodeNumber int `json:"EpisodeNumber"`
	Title	string	`json:"Title"`
	Director string `json:"Director"`
	Writer string `json:"Writer"`
	OriginalAirDate	string	`json:"OriginalAirDate"`
	WikiLink	string	`json:"WikiLeak"`
}

type Subtitle struct {
	ID	int `json:"Id"`
	RepresentativeTimestamp	string	`json:"RepresentativeTimestamp"`
	Episode	string	`json:"Episode"`
	StartTimestamp	string `json:"StartTimestamp"`
	EndTimestamp	string	`json:"EndTimestamp"`
	Content	string	`json:"Content"`
	Language	string `json:"Language"`
}

type Caption struct {
	Episode Episode `json:"Episode"`
	Frame	Frame	`json:"Frame"`
	Subtitles	[]Subtitle	`json:"Subtitles"`
	Nearby	[]Frame	`json:"Nearby"`
}


func (f *Frame) GetPhotoUrl() string {
	return fmt.Sprintf("http://frinkiac.com/img/%s/%d.jpg", f.Episode, f.Timestamp)
}

func (f *Frame) GetCaptionPhotoUrl(caption string) string {
	b64Caption := base64.StdEncoding.EncodeToString([]byte(caption))
	return fmt.Sprintf("http://frinkiac.com/meme/%s/%d.jpg?b64lines=%s", f.Episode, f.Timestamp, b64Caption)
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

func (f *FrinkiacClient) GetCaption(episode string, timestamp string) (Caption, error) {
	var info Caption
	req, err := http.NewRequest("GET", fmt.Sprintf("https://frinkiac.com/api/caption?e=%s&t=%s", episode, timestamp), nil)
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