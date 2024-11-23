![twigots](assets/twigots.png)

[![Go Reference](https://pkg.go.dev/badge/github.com/ahobsonsayers/twigots.svg)](https://pkg.go.dev/github.com/ahobsonsayers/twigots)
[![Go Report
Card](https://goreportcard.com/badge/github.com/ahobsonsayers/twigots)](https://goreportcard.com/report/github.com/ahobsonsayers/twigots)
[![License - MIT](https://img.shields.io/badge/License-MIT-9C27B0)](LICENSE)

A go package to fetch ticket listings from the [Twickets](https://www.twickets.live) Live Feed.

Includes utilities to help filter the ticket listings and find the ones you want!

Powers (the similarly creatively named)
[twitchets](https://github.com/ahobsonsayers/twitchets), an application to watch for new
listings of tickets you want and send notifications so you can snap them up!

## Installation

```bash
go get -u github.com/ahobsonsayers/twigots
```

## Getting an API key

To use this package you will need a Twickets API key. Twickets currently do not have a free API
HOWEVER it is possible to **easily** obtain an API key to use with this library.

To do this simply visit the [Twickets Live Feed](https://www.twickets.live/app/catalog/browse),
open you browser `Developer Tools` (by pressing `F12`), navigate to the `Network` tab, look for the
`GET` request to `https://www.twickets.live/services/catalogue` and copy the `api_key` query
parameter from the request.

This API key is not provided here due to liability concerns, but the key seems to be fixed/unchanging and
is very easy to get using the instructions above.

## Example Usage

```go
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
	client := twigots.NewClient(nil) // Or use a custom http client
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

	// Filter ticket listing for the ones we want
	filteredListings, err := listings.Filter(
		// Filter for listings of Hamilton tickets
		twigots.Filter{Event: "Hamilton"},
		// Also filter for listings of Coldplay tickets.
		// Lets impose extra restrictions on these.
		twigots.Filter{
			Event:           "Coldplay", // Required
			EventSimilarity: 1,          // Avoid false positives
			Regions: []twigots.Region{
				twigots.RegionLondon,
				twigots.RegionSouth,
			},
			NumTickets:  2,   // Exactly 2 tickets in the listing
			MinDiscount: 0.1, // Discount of > 10%
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	for _, listing := range filteredListings {
		slog.Info(
			"Ticket listing found matching a filter",
			"Event", listing.Event.Name,
			"NumTickets", listing.NumTickets,
			"Price", listing.TotalPriceInclFee().String(),
			"OriginalPrice", listing.OriginalTicketPrice().String(),
			"URL", listing.URL(),
		)
	}
}
```

## Why the name twigots?

Because its a stupid mash up of Tickets and Go... and also why not?
