package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/ahobsonsayers/twigots"
	"github.com/ahobsonsayers/twigots/filter"
)

func main() {
	apiKey := "my_api_key"

	// Fetch ticket listings
	client := twigots.NewClient() // Or use a custom http client
	listings, err := client.FetchTicketListings(
		context.Background(),
		twigots.FetchTicketListingsInput{
			// Required
			APIKey:  apiKey,
			Country: twigots.CountryUnitedKingdom, // Only UK is supported at the moment
			// Optional. See all options in godoc
			CreatedBefore: time.Now(),
			CreatedAfter:  time.Now().Add(time.Duration(-5 * time.Minute)), // 5 mins ago
			MaxNumber:     100,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Fetched %d ticket listings", len(listings))

	// Filter ticket listings just by name
	hamiltonListings := filter.FilterTicketListings(
		listings,
		filter.EventName("Hamilton", 0.9), // Similarity of 0.9 - allow minor mismatches
	)
	for _, listing := range hamiltonListings {
		slog.Info(
			"Found Hamilton ticket listing",
			"Event", listing.Event.Name,
			"NumTickets", listing.NumTickets,
			"Price", listing.TotalPriceInclFee().String(),
			"OriginalPrice", listing.OriginalTicketPrice().String(),
			"URL", listing.URL(),
		)
	}

	// Filter ticket listings just by several filters
	coldplayListings := filter.FilterTicketListings(
		listings,
		filter.EventName("Coldplay", 1), // Similarity of 1 - exact match only
		filter.EventRegion( // Only in specific regions
			twigots.RegionLondon,
			twigots.RegionSouth,
		),
		filter.NumTickets(2),    // Exactly 2 tickets in the listing
		filter.MinDiscount(0.1), // Discount of > 10%
	)
	for _, listing := range coldplayListings {
		slog.Info(
			"Found Coldplay ticket listing",
			"Event", listing.Event.Name,
			"NumTickets", listing.NumTickets,
			"Price", listing.TotalPriceInclFee().String(),
			"OriginalPrice", listing.OriginalTicketPrice().String(),
			"URL", listing.URL(),
		)
	}
}
