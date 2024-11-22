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
			Regions:       []twigots.Region{twigots.RegionLondon},
			CreatedBefore: time.Now(),
			CreatedAfter:  time.Now().Add(time.Duration(-5 * time.Minute)), // 5 mins ago
			MaxNumber:     100,
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
		NumTickets:  2,
		MinDiscount: 10, // Minimum 10% discount (INCLUDING booking fee)

	}

	filteredTickets := tickets.Filter(filter)
	for _, ticket := range filteredTickets {
		slog.Info(
			"Found tickets for monitored event",
			"eventName", ticket.Event.Name,
			"numTickets", ticket.NumTickets,
			"ticketPrice", ticket.TotalPriceInclFee().String(),
			"originalTicketPrice", ticket.OriginalTicketPrice().String(),
			"link", ticket.URL(),
		)
	}
}
