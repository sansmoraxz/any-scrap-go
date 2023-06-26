package steam

import (
	"encoding/json"
	"fmt"
	"github.com/sansmoraxz/any-scrape-go/queue"
	"io"
	"net/http"
)

type Reviews struct {
	Client   queue.Queue
	SaveFunc func([]byte, int, string) error
	Params   ReviewParams
	hasNext  bool
}

type ReviewParams struct {
	AppID    int
	Cursor   string `default:"*"`
	Language string `default:"all"`

	Filter     string `default:"recent"`
	NumPerPage int    `default:"100"`
}

func NewSteamReviewParamsWithAppID(appid int) *ReviewParams {
	return &ReviewParams{
		AppID:      appid,
		Cursor:     "*",
		Language:   "all",
		Filter:     "recent",
		NumPerPage: 100,
	}
}

func NewSteamReviewParamsWithAppIDAndCursor(appid int, cursor string) *ReviewParams {
	return &ReviewParams{
		AppID:      appid,
		Cursor:     cursor,
		Language:   "all",
		Filter:     "recent",
		NumPerPage: 100,
	}
}

type ReviewData struct {
	Success      int                `json:"success"`
	QuerySummary ReviewQuerySummary `json:"query_summary"`
	Reviews      interface{}        `json:"reviews"`
	Cursor       string             `json:"cursor"`
}

type ReviewQuerySummary struct {
	NumReviews      int    `json:"num_reviews"`
	ReviewScore     int    `json:"review_score"`
	ReviewScoreDesc string `json:"review_score_desc"`
	TotalPositive   int    `json:"total_positive"`
	TotalNegative   int    `json:"total_negative"`
	TotalReviews    int    `json:"total_reviews"`
}

func (s *Reviews) HasNext() bool {
	return s.hasNext || s.Params.Cursor == "*" || s.Params.Cursor == ""
}

func (s *Reviews) AddToQueue() error {
	buffer, err := json.Marshal(s.Params)
	if err != nil {
		return err
	}

	err = s.Client.WriteMessage(buffer, nil)
	if err != nil {
		return err
	}
	return nil
}

func (p *ReviewParams) GetURL() string {
	return fmt.Sprintf("https://store.steampowered.com/appreviews/%d?json=1&filter=%s&language=%s&num_per_page=%d&cursor=%s",
		p.AppID, p.Filter, p.Language, p.NumPerPage, p.Cursor)
}

func (p *ReviewParams) Fetch() ([]byte, error) {
	// http get request
	resp, err := http.Get(p.GetURL())
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	return body, err
}

func (data ReviewData) validate() bool {
	return data.Success == 1
}

func (s *Reviews) ScrapeFromQueue() error {
	fn := func(msg []byte) error {
		var _params ReviewParams
		err := json.Unmarshal(msg, &_params)
		if err != nil {
			return err
		}

		reviewData, err := s.Params.Fetch()
		if err != nil {
			return err
		}

		var _data ReviewData
		err = json.Unmarshal(reviewData, &_data)
		if err != nil {
			return err
		}

		if !_data.validate() {
			return fmt.Errorf("invalid review data for appid %d and cursor %s", _params.AppID, _params.Cursor)
		}

		s.Params = _params
		s.hasNext = _data.Cursor != _params.Cursor // as long as the cursor is different, there is more data to scrape
		s.Params.Cursor = _data.Cursor

		if s.hasNext {
			return s.SaveFunc(reviewData, _params.AppID, _params.Cursor)
		}
		// duplicate data not saved
		return nil
	}
	_, err := s.Client.ReadMessage(fn)
	return err
}
