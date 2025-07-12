package twigots

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/imroc/req/v3"
	"github.com/k3a/html2text"
)

type Client struct {
	client *req.Client
	apiKey string
}

func (c *Client) Client() *http.Client {
	return c.client.GetClient()
}

// FetchTicketListingsInput defines parameters when getting ticket listings.
//
// Ticket listings can either be fetched by maximum number or by time period.
// The default is to get a maximum number of ticket listings.
//
// If both a maximum number and a time period are set, whichever condition
// is met first will stop the fetching of ticket listings.
type FetchTicketListingsInput struct {
	// Required fields
	Country Country

	// Regions for which to fetch ticket listings from.
	// Leave this unset or empty to fetch listings from any region.
	// Defaults to any region (unset).
	Regions []Region

	// MaxNumber is the maximum number of ticket listings to fetch.
	// If getting ticket listings within in a time period using `CreatedAfter`, set this to an arbitrarily
	// large number (e.g. 250) to ensure all listings in the period are fetched, while preventing
	// accidentally fetching too many listings and possibly being rate limited or blocked.
	// Defaults to 10.
	// Set to -1 if no limit is desired. This is dangerous and should only be used with well constrained time periods.
	MaxNumber int

	// CreatedAfter is the time which ticket listings must have been created after to be fetched.
	// Set this to fetch listings within a time period.
	// Set `MaxNumber` to an arbitrarily large number (e.g. 250) to ensure all listings in the period are fetched,
	// while preventing  accidentally fetching too many listings and possibly being rate limited or blocked.
	CreatedAfter time.Time

	// CreatedBefore is the time which ticket listings must have been created before to be fetched.
	// Set this to fetch listings within a time period.
	// Defaults to current time.
	CreatedBefore time.Time

	// NumPerRequest is the number of ticket listings to fetch in each request.
	// Not all requested listings are fetched at once - instead a series of requests are made,
	// each fetching the number of listings specified here. In theory this can be arbitrarily
	// large to prevent having to make too many requests, however it has been known that any
	// other number than 10 can sometimes not work.
	// Defaults to 10. Usually can be ignored.
	NumPerRequest int
}

func (f *FetchTicketListingsInput) applyDefaults() {
	if f.MaxNumber == 0 {
		f.MaxNumber = 10
	}
	if f.CreatedBefore.IsZero() {
		f.CreatedBefore = time.Now()
	}
	if f.NumPerRequest <= 0 {
		f.NumPerRequest = 10
	}
}

// Validate the input struct used to get ticket listings.
// This is used internally to check the input, but can also be used externally.
func (f FetchTicketListingsInput) Validate() error {
	if f.Country.Value == "" {
		return errors.New("country must be set")
	}
	if !Countries.Contains(f.Country) {
		return fmt.Errorf("country '%s' is not valid", f.Country)
	}
	if f.CreatedBefore.Before(f.CreatedAfter) {
		return errors.New("created after time must be after the created before time")
	}
	if f.MaxNumber < 0 && f.CreatedAfter.IsZero() {
		return errors.New("if not limiting number of ticket listings, created after must be set")
	}
	return nil
}

// FetchTicketListings gets ticket listings using the specified input.
func (c *Client) FetchTicketListings(
	ctx context.Context,
	input FetchTicketListingsInput,
) (TicketListings, error) {
	input.applyDefaults()
	err := input.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// Iterate through feeds until have equal to or more ticket listings than desired
	ticketListings := make(TicketListings, 0, input.MaxNumber)
	earliestTicketTime := input.CreatedBefore
	for (input.MaxNumber < 0 || len(ticketListings) < input.MaxNumber) &&
		earliestTicketTime.After(input.CreatedAfter) {

		feedUrl, err := FeedUrl(FeedUrlInput{
			APIKey:      c.apiKey,
			Country:     input.Country,
			Regions:     input.Regions,
			NumListings: input.NumPerRequest,
			BeforeTime:  earliestTicketTime,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get feed url: %w", err)
		}

		feedTicketListings, err := c.FetchTicketListingsByFeedUrl(ctx, feedUrl)
		if err != nil {
			return nil, err
		}
		if len(feedTicketListings) == 0 {
			return nil, errors.New("no tickets returned")
		}

		ticketListings = append(ticketListings, feedTicketListings...)
		earliestTicketTime = feedTicketListings[len(feedTicketListings)-1].CreatedAt.Time
	}

	// Only return ticket listings requested
	ticketListings = sliceToMaxNumTicketListings(ticketListings, input.MaxNumber)
	ticketListings = filterToCreatedAfter(ticketListings, input.CreatedAfter)

	return ticketListings, nil
}

// FetchTicketListings gets ticket listings using the specified feel url.
func (c *Client) FetchTicketListingsByFeedUrl(
	ctx context.Context,
	feedUrl string,
) (TicketListings, error) {
	response, err := c.client.R().SetContext(ctx).Get(feedUrl)
	if err != nil {
		return nil, nil
	}

	if !response.IsSuccessState() {
		errorBody := html2text.HTML2Text(response.String())
		return nil, fmt.Errorf(
			"failed to fetch tickets: %s\n\nResponse:\n%s",
			response.GetStatus(), errorBody,
		)
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return UnmarshalTwicketsFeedJson(bodyBytes)
}

// NewClient creates a new Twickets client
func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("api key must be set")
	}

	client := req.C().ImpersonateChrome()
	return &Client{
		client: client,
		apiKey: apiKey,
	}, nil
}

func sliceToMaxNumTicketListings(listings TicketListings, maxNumTicketListings int) TicketListings {
	if len(listings) > maxNumTicketListings {
		listings = listings[:maxNumTicketListings]
	}
	return listings
}

func filterToCreatedAfter(listings TicketListings, createdAfter time.Time) TicketListings {
	filteredListings := make(TicketListings, 0, len(listings))
	for _, listing := range listings {
		if listing.CreatedAt.After(createdAfter) {
			filteredListings = append(filteredListings, listing)
		}
	}
	return filteredListings
}
