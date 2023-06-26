package steam

import (
	"encoding/json"
	"fmt"
	"github.com/sansmoraxz/any-scrape-go/queue"
	"io"
	"net/http"
)

type AppDetails struct {
	Client   queue.Queue
	SaveFunc func([]byte, int, string) error
	Params   AppDetailsParams
}

type AppDetailsParams struct {
	AppID int
}

func (p *AppDetailsParams) GetURL() string {
	return fmt.Sprintf("https://store.steampowered.com/api/appdetails?appids=%d", p.AppID)
}

func (p *AppDetailsParams) Fetch() ([]byte, error) {
	resp, err := http.Get(p.GetURL())
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	return body, err
}

func NewSteamAppDetailsParamsWithAppID(appid int) *AppDetailsParams {
	return &AppDetailsParams{
		AppID: appid,
	}
}

type AppDetailsData struct {
	Success bool `json:"success"`
}

// appDetailsResponse make sure to dereference to field App not this struct
type appDetailsResponse struct {
	App map[string]AppDetailsData
}

func (data appDetailsResponse) validate() bool {
	for _, v := range data.App {
		return v.Success
	}
	return false
}

// HasNext always returns false because this is a single page scraper
func (p *AppDetails) HasNext() bool {
	return false
}

func (p *AppDetails) AddToQueue() error {
	buffer, err := json.Marshal(p.Params)
	if err != nil {
		return err
	}

	err = p.Client.WriteMessage(buffer, nil)
	if err != nil {
		return err
	}

	return nil
}

func (p *AppDetails) ScrapeFromQueue() error {
	fn := func(msg []byte) error {
		var _params AppDetailsParams
		err := json.Unmarshal(msg, &_params)
		if err != nil {
			return err
		}
		appdata, err := _params.Fetch()
		if err != nil {
			return err
		}
		var _data appDetailsResponse
		err = json.Unmarshal(appdata, &_data)
		if err != nil {
			return err
		}
		if !_data.validate() {
			return fmt.Errorf("invalid app data for appid %d", _params.AppID)
		}

		return p.SaveFunc(msg, p.Params.AppID, "")
	}
	_, err := p.Client.ReadMessage(fn)
	return err
}
