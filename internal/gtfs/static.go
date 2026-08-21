package gtfs

import (
	"StartRomagnaAPI/internal/auth"
	"fmt"
	"net/http"

	"github.com/Leocraft1/gtfsparser-with-reader"
)

func GetStaticFeed(url string, skip_valid bool) (*gtfsparserwr.Feed, error) {
	req, err := auth.BasicAuth("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	feed := gtfsparserwr.NewFeed()
	if skip_valid {
		feed.Opts.SkipStopTimeValidation
	}
	if err := feed.ParseReader(resp.Body); err != nil {
        return nil, fmt.Errorf("%s: parse gtfs: %w", err)
    }

	return feed, nil
}