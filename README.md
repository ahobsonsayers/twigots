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

## Getting an API Key

To use this tool, you will need a Twickets API key. Although Twickets doesn't provide a free API, you can easily obtain a key by following these steps:

1.  Visit the [Twickets Live Feed](https://www.twickets.live/app/catalog/browse)
2.  Open your browser's Developer Tools (F12) and navigate to the Network tab
3.  Look for the GET request to `https://www.twickets.live/services/catalogue` and copy the `api_key` query parameter. You might need to refresh the page first if nothing appears in this tab.

This API key is not provided here due to liability concerns, but it appears to be a fixed, unchanging value.

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

[![Hits](https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Fahobsonsayers%2Ftwigots&count_bg=%2379C83D&title_bg=%23555555&icon=&icon_color=%23E7E7E7&title=visitors+day+%2F+total&edge_flat=false)](https://hits.seeyoufarm.com)
