package steam

import (
	"encoding/json"
	"fmt"
	"github.com/sansmoraxz/any-scrape-go/queue"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type AppNews struct {
	Client   queue.Queue
	SaveFunc func([]byte, int, string) error
	Params   AppNewsParams
	hasNext  bool `default:"true"`
}

type AppNewsParams struct {
	AppID int
	// Maximum length for the content to return, if this is 0 the full content is returned, if it's less; then a blurb is generated to fit.
	MaxLength int
	// Maximum number of news items to retrieve (default 20)
	Count int
	// Retrieve posts earlier than this date (unix epoch timestamp)
	// If omitted, defaults to the current time
	EndDate int
	// list of feed names to return news for
	Feeds []string
}

func (p *AppNewsParams) GetURL() string {
	var urlParams []string
	if p.MaxLength != 0 {
		urlParams = append(urlParams, "maxlength="+strconv.Itoa(p.MaxLength))
	}
	if p.EndDate != 0 {
		urlParams = append(urlParams, "enddate="+strconv.Itoa(p.EndDate))
	}
	if p.Count != 0 {
		urlParams = append(urlParams, "count="+strconv.Itoa(p.Count))
	}
	if len(p.Feeds) != 0 {
		urlParams = append(urlParams, "feeds="+strings.Join(p.Feeds, ","))
	}
	url := "https://api.steampowered.com/ISteamNews/GetNewsForApp/v2/?appid=" + strconv.Itoa(p.AppID) + strings.Join(urlParams, "&")
	return url
}

func (p *AppNewsParams) Fetch() ([]byte, error) {
	resp, err := http.Get(p.GetURL())
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	return body, err
}

type AppNewsData struct {
	AppNews NewsData `json:"appnews"`
}

type NewsData struct {
	AppID     int        `json:"appid"`
	NewsItems []NewsItem `json:"newsitems"`
	Count     int        `json:"count"`
}

type NewsItem struct {
	Gid string `json:"gid"`

	FeedLabel string `json:"feedlabel"`
	FeedName  string `json:"feedname"`
	FeedType  int    `json:"feed_type"`

	Date int `json:"date"`
}

func (r *AppNews) HasNext() bool {
	return r.hasNext
}

func (r *AppNews) AddToQueue() error {
	buffer, err := json.Marshal(r.Params)
	if err != nil {
		return err
	}

	err = r.Client.WriteMessage(buffer, nil)
	if err != nil {
		return err
	}
	return nil
}

func (r *AppNews) ScrapeFromQueue() error {
	fn := func(msg []byte) error {
		var _params AppNewsParams
		err := json.Unmarshal(msg, &_params)
		if err != nil {
			return err
		}

		reviewData, err := r.Params.Fetch()
		if err != nil {
			return err
		}

		var _data AppNewsData
		err = json.Unmarshal(reviewData, &_data)
		if err != nil {
			return err
		}

		r.hasNext = _data.AppNews.Count > 0
		if r.hasNext {
			_params.EndDate = _data.AppNews.NewsItems[len(_data.AppNews.NewsItems)-1].Date - 1
			timeSpan := fmt.Sprintf("%d-%d", _data.AppNews.NewsItems[0].Date, _data.AppNews.NewsItems[len(_data.AppNews.NewsItems)-1].Date)
			r.Params = _params
			return r.SaveFunc(reviewData, _params.AppID, timeSpan)
		} else {
			return nil
		}
	}
	_, err := r.Client.ReadMessage(fn)
	return err
}
