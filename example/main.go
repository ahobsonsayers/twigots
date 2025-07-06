package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/ahobsonsayers/twigots"
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
	hamiltonListings := twigots.FilterTicketListings(
		listings,
		twigots.EventNamePredicate("Hamilton", 0.9), // Similarity of 0.9 - allow minor mismatches
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
	coldplayListings := twigots.FilterTicketListings(
		listings,
		twigots.EventNamePredicate("Coldplay", 1), // Similarity of 1 - exact match only
		twigots.RegionsPredicate( // Only in specific regions
			twigots.RegionLondon,
			twigots.RegionSouth,
		),
		twigots.NumTicketsPredicate(2),    // Exactly 2 tickets in the listing
		twigots.MinDiscountPredicate(0.1), // Discount of > 10%
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
