# Twigots

A (unofficial) go package to fetch ticket listings from the Twickets live feed.

Includes utilities to help filtered ticket listings to get the ones you want!

## Note: Getting a API key

To use this package you will need a Twickets API key. Twickets currently do not have a free API HOWEVER

## Example Usage

See the example in the `examples` directory, or see below.

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

	client := twigots.NewClient(nil) // Or use a custom http client
	tickets, err := client.FetchTicketListings(
		context.Background(),
		twigots.FetchTicketListingsInput{
			// Required
			APIKey:  apiKey,
			Country: twigots.CountryUnitedKingdom, // Only UK is supported at the moment
			// Optional. See all options in godoc
			Regions:              []twigots.Region{twigots.RegionLondon},
			CreatedBefore:        time.Now(),
			CreatedAfter:         time.Now().Add(time.Duration(-5 * time.Minute)), // 5 mins ago
			MaxNumTicketListings: 100,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	filter := twigots.Filter{
		// Required
		Event: "Coldplay",
		// Optional
		EventSimilarity: 100, // Avoid false positives
		Regions: []twigots.Region{
			twigots.RegionLondon,
			twigots.RegionSouth,
		},
		NumTickets: 2,
		Discount:   10, // Minimum 10% discount (INCLUDING booking fee)

	}

	filteredTickets := tickets.Filter(filter)
	for _, ticket := range filteredTickets {
		slog.Info(
			"Found tickets for monitored event",
			"eventName", ticket.Event.Name,
			"numTickets", ticket.TicketQuantity,
			"ticketPrice", ticket.TotalPriceInclFee().String(),
			"originalTicketPrice", ticket.OriginalTicketPrice().String(),
			"link", ticket.Link(),
		)
	}
}
```

## Why the name twigots?

Because its a stupid mash up of Tickets and Go... and also why not?
