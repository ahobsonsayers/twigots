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

const TwicketsURL = "https://www.twickets.live"

var twicketsUrl *url.URL

func init() {
	var err error
	twicketsUrl, err = url.Parse(TwicketsURL)
	if err != nil {
		log.Fatal("failed to parse twickets url")
	}
}

// ListingURL gets the url of a listing given its id and the number of tickets in the listing.
//
// Format is:
// https://www.twickets.live/app/block/<ticketId>,<numTickets>
func ListingURL(listingId string, numTickets int) string {
	ticketUrl := cloneURL(twicketsUrl)
	ticketUrl = ticketUrl.JoinPath("app", "block", fmt.Sprintf("%s,%d", listingId, numTickets))
	return ticketUrl.String()
}

type FeedUrlInput struct {
	// Required fields
	Country Country

	// Optional fields
	Regions    []Region  // Defaults to all country regions
	BeforeTime time.Time // Defaults to current time
}

// Validate the input struct used to get the feed url.
// This is used internally to check the input, but can also be used externally.
func (f FeedUrlInput) Validate() error {
	if f.Country.Value == "" {
		return errors.New("country must be set")
	}
	if !Countries.Contains(f.Country) {
		return fmt.Errorf("country '%s' is not valid", f.Country)
	}
	return nil
}

// FeedUrl gets the url of a ticket listings feed.
// Note: The number of ticket listings (that are non-delisted) in the feed at this url will ALWAYS be 10.
// There may be any number of additional delisted ticket listings.
//
// Format is:
// https://www.twickets.live/services/g2/catalogue?countryCode=GB&limit=10
func FeedUrl(input FeedUrlInput) (string, error) {
	err := input.Validate()
	if err != nil {
		return "", fmt.Errorf("invalid input parameters: %w", err)
	}

	feedUrl := cloneURL(twicketsUrl)
	feedUrl = feedUrl.JoinPath("services", "g2", "catalogue")

	// Set query params
	queryParams := feedUrl.Query()

	queryParams.Set("countryCode", string(input.Country.Value))

	regionCodes := make([]string, 0, len(input.Regions))
	for _, region := range input.Regions {
		if Regions.Contains(region) {
			regionCodes = append(regionCodes, string(region.Value))
		}
	}
	if len(regionCodes) > 0 {
		queryParams.Set("regionCodes", strings.Join(regionCodes, ","))
	}

	if !input.BeforeTime.IsZero() {
		maxTime := input.BeforeTime.UnixMilli()
		queryParams.Set("maxTime", strconv.Itoa(int(maxTime)))
	}

	queryParams.Set("limit", "10") // limit must always be 10 to not get an error

	feedUrl.RawQuery = queryParams.Encode()

	return feedUrl.String(), nil
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
