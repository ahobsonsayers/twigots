# twigots

A go package to fetch ticket listings from the [Twickets](https://www.twickets.live) Live Feed.

Includes utilities to help filter the ticket listings and find the ones you want!

Powers (the similarly creatively named)
[twitchets](https://github.com/ahobsonsayers/twitchets), an application to watch for new listings of tickets you want, and send notifications so you can snap them up!

## Getting an API key

To use this package you will need a Twickets API key. Twickets currently do not have a free API
HOWEVER it is possible to **easily** obtain an API key to use with this library.

To do this simply visit the [Twickets Live Feed](https://www.twickets.live/app/catalog/browse)
open you browser `Developer Tools` (by pressing F12), navigate to the `Network` tab, look for the
`GET` request to `https://www.twickets.live/services/catalogue` and copy the `api_key` query
parameter from the request.

This API is not provided here due to liability concerns, but seems to be fixed/unchanging, and
is very easy to find.

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
