package steam

import (
	"encoding/json"
	"fmt"
	"github.com/sansmoraxz/any-scrape-go/queue"
	"io"
	"net/http"
	"strconv"
)

type Reviews struct {
	Client   queue.Queue
	SaveFunc func([]byte, int) error
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

func (p *ReviewParams) fetchSteamReviews() ([]byte, error) {
	// http get request
	resp, err := http.Get(
		"https://store.steampowered.com/appreviews/" + strconv.Itoa(p.AppID) +
			"?json=1&filter=" + p.Filter +
			"&language=" + p.Language +
			"&num_per_page=" + strconv.Itoa(p.NumPerPage) +
			"&cursor=" + p.Cursor)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	return body, err
}

func validateReviewData(data ReviewData) bool {
	return data.Success == 1
}

func (s *Reviews) ScrapeFromQueue() error {
	fn := func(msg []byte) error {
		var _params ReviewParams
		err := json.Unmarshal(msg, &_params)
		if err != nil {
			return err
		}

		reviewData, err := s.Params.fetchSteamReviews()
		if err != nil {
			return err
		}

		var _data ReviewData
		err = json.Unmarshal(reviewData, &_data)
		if err != nil {
			return err
		}

		if !validateReviewData(_data) {
			return fmt.Errorf("invalid review data for appid %d and cursor %s", _params.AppID, _params.Cursor)
		}

		s.Params = _params
		s.hasNext = _data.Cursor != _params.Cursor
		s.Params.Cursor = _data.Cursor

		return s.SaveFunc(reviewData, s.Params.AppID)
	}
	_, err := s.Client.ReadMessage(fn)
	if err != nil {
		return err
	}
	return nil
}
