package twigots

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	TwicketsURL = "https://www.twigots.live"

	countryQueryKey = "countryCode"
	regionQueryKey  = "regionCode"
)

var twicketsUrl *url.URL = func() *url.URL {
	twicketsUrl, err := url.Parse(TwicketsURL)
	if err != nil {
		log.Fatal("failed to parse twickets url")
	}
	return twicketsUrl
}()

// ListingURL gets the url of a listing given its id and the
// number of tickets in the listing.
// Format is:
// https://www.twigots.live/app/block/<ticketId>,<numTickets>
func ListingURL(listingId string, numTickets int) string {
	ticketUrl := cloneURL(twicketsUrl)
	ticketUrl = ticketUrl.JoinPath("app", "block", fmt.Sprintf("%s,%d", listingId, numTickets))
	return ticketUrl.String()
}

type FeedUrlInput struct {
	// Required fields
	APIKey  string
	Country Country

	// Optional fields
	Regions     []Region  // Defaults to all country regions
	NumListings int       // Defaults to 10 ticket listings
	BeforeTime  time.Time // Defaults to current time
}

func (f FeedUrlInput) validate() error {
	if f.APIKey == "" {
		return errors.New("api key must be set")
	}
	if f.Country.Value == "" {
		return errors.New("country must be set")
	}
	if !Countries.Contains(f.Country) {
		return fmt.Errorf("country '%s' is not valid", f.Country)
	}
	return nil
}

// FeedUrl gets the url of a feed of ticket listings
// E.g. https://www.twigots.live/services/catalogue?q=countryCode=GB&count=100&api_key=<api_key>
func FeedUrl(input FeedUrlInput) (string, error) {
	err := input.validate()
	if err != nil {
		return "", fmt.Errorf("invalid input parameters: %w", err)
	}

	feedUrl := cloneURL(twicketsUrl)
	feedUrl = feedUrl.JoinPath("services", "catalogue")

	// Set query params
	queryParams := feedUrl.Query()

	locationQuery := apiLocationQuery(input.Country, input.Regions...)
	if locationQuery != "" {
		queryParams.Set("q", locationQuery)
	}

	if !input.BeforeTime.IsZero() {
		maxTime := input.BeforeTime.UnixMilli()
		queryParams.Set("maxTime", strconv.Itoa(int(maxTime)))
	}

	if input.NumListings > 0 {
		count := strconv.Itoa(input.NumListings)
		queryParams.Set("count", count)
	}

	queryParams.Set("api_key", input.APIKey)

	// Set query
	encodedQuery := queryParams.Encode()
	encodedQuery = strings.ReplaceAll(encodedQuery, "%3D", "=")
	encodedQuery = strings.ReplaceAll(encodedQuery, "%2C", ",")
	feedUrl.RawQuery = encodedQuery

	return feedUrl.String(), nil
}

// apiLocationQuery converts a country and selection of regions to an api query string
func apiLocationQuery(country Country, regions ...Region) string {
	if !Countries.Contains(country) {
		return ""
	}

	queryParts := make([]string, 0, len(regions)+1)

	countryQuery := fmt.Sprintf("%s=%s", countryQueryKey, country.Value)
	queryParts = append(queryParts, countryQuery)

	for _, region := range regions {
		if Regions.Contains(region) {
			regionQuery := fmt.Sprintf("%s=%s", regionQueryKey, region.Value)
			queryParts = append(queryParts, regionQuery)
		}
	}

	return strings.Join(queryParts, ",")
}

// cloneUrl clones a url. Copied directly from net/http internals
// See: https://github.com/golang/go/blob/go1.19/src/net/http/clone.go#L22
func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	u2 := new(url.URL)
	*u2 = *u
	if u.User != nil {
		u2.User = new(url.Userinfo)
		*u2.User = *u.User
	}
	return u2
}
