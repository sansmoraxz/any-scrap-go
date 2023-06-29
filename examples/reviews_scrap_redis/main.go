package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/sansmoraxz/any-scrape-go/queue"
	"github.com/sansmoraxz/any-scrape-go/scrapables/steam"
	"os"
	"strconv"
)

func main() {
	println("Starting...")
	ctx := context.Background()
	var q queue.Queue
	var err error
	q, err = queue.NewRedis(
		ctx,
		queue.RedisConfig{
			RedisURL:      os.Getenv("REDIS_URL"),
			RedisPassword: os.Getenv("REDIS_PASSWORD"),
			RedisDB:       0,
			RedisKey:      "steam_reviews",
		},
	)
	if err != nil {
		panic(err)
	}
	defer func(q queue.Queue) {
		err = q.Close()
		if err != nil {
			panic(err)
		}
	}(q)
	appid, err := strconv.ParseInt(os.Getenv("APPID"), 10, 32)
	if err != nil {
		panic(err)
	}

	params := steam.NewSteamReviewParamsWithAppID(int(appid))

	saveFn := func(data []byte, appid int, cursor string) error {
		appidStr := fmt.Sprintf("%d", appid)
		fileName := fmt.Sprintf("/app/scrap_data/reviews_%s_%s.json", appidStr, hex.EncodeToString([]byte(cursor)))
		err := os.WriteFile(fileName, data, 0666)
		fmt.Printf("Saved %s.%s\n", appidStr, cursor)
		if err != nil {
			panic(err)
		}
		return nil
	}
	revs := steam.Reviews{
		Client:   q,
		Params:   *params,
		SaveFunc: saveFn,
	}

	for {
		if err := revs.AddToQueue(); err != nil {
			panic(err)
		}
		if err := revs.ScrapeFromQueue(); err != nil {
			panic(err)
		}
		if !revs.HasNext() {
			break
		}
	}
	fmt.Println("Done...")
}
