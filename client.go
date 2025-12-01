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
}

func (f *FetchTicketListingsInput) applyDefaults() {
	if f.MaxNumber == 0 {
		f.MaxNumber = 10
	}
	if f.CreatedBefore.IsZero() {
		f.CreatedBefore = time.Now()
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
		return nil, fmt.Errorf("failed ready response body: %w", err)
	}

	return UnmarshalTwicketsFeedJson(bodyBytes)
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

	// Iterate through feeds until have the number of listings desired
	// or listings creation time is before the created after input
	listings := make(TicketListings, 0, input.MaxNumber)
	earliestTicketTime := input.CreatedBefore
	numListingsRemaining := input.MaxNumber
	for {

		// Get feed url
		feedUrl, err := FeedUrl(FeedUrlInput{
			APIKey:     c.apiKey,
			Country:    input.Country,
			Regions:    input.Regions,
			BeforeTime: earliestTicketTime,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get feed url: %w", err)
		}

		// Fetch new listings
		newListings, err := c.FetchTicketListingsByFeedUrl(ctx, feedUrl)
		if err != nil {
			return nil, err
		}
		if len(newListings) == 0 {
			return nil, errors.New("no listings returned")
		}

		// Process listings, ignoring those created too early.
		// Will return shouldBreak if a break condition is met.
		processedListings, shouldBreak := processFeedListings(
			newListings, numListingsRemaining, input.CreatedAfter,
		)

		// Update listings
		listings = append(listings, processedListings...)
		if shouldBreak {
			break
		}

		// Update loop variables
		earliestTicketTime = listings[len(listings)-1].CreatedAt.Time
		numListingsRemaining = input.MaxNumber - len(listings)
	}

	return listings, nil
}

// processFeedListings, ignoring those created too early.
// Returns the processed ticket listings, an whether iteration should continue.
func processFeedListings(
	listings TicketListings,
	maxNumber int,
	createdAfter time.Time,
) ([]TicketListing, bool) {
	processedListings := make([]TicketListing, 0, len(listings))
	for _, listing := range listings {

		// If listing NOT created after the earliest allowed time, break
		if !listing.CreatedAt.After(createdAfter) {
			return processedListings, true
		}

		// Update processes listings
		processedListings = append(processedListings, listing)

		// If number of listings matches the max number, break
		if len(processedListings) == maxNumber {
			return processedListings, true
		}
	}

	return processedListings, false
}

type ClientOpt func(*req.Client) error

// NewClient creates a new Twickets client
func NewClient(apiKey string, opts ...ClientOpt) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("api key must be set")
	}

	client := req.C()
	client = client.ImpersonateChrome()
	for _, opt := range opts {
		err := opt(client)
		if err != nil {
			return nil, err
		}
	}

	return &Client{
		client: client,
		apiKey: apiKey,
	}, nil
}
