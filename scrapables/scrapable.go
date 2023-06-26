package scrapables

type Scrapable interface {
	// HasNext returns true if there are more pages to scrape
	HasNext() bool
	// AddToQueue adds the current page to the queue
	AddToQueue() error
	// ScrapeFromQueue gets a page from the queue and saves it
	// and advances to the next page
	ScrapeFromQueue() error
}

type ScrapableParams interface {
	// Fetch gets the data from the API
	Fetch() ([]byte, error)
	// GetURL returns the URL to fetch data from
	GetURL() string
}
