package filter

import (
	"testing"
	"time"

	"github.com/ahobsonsayers/twigots"
	"github.com/stretchr/testify/require"
)

func TestCreatedAfterPredicate(t *testing.T) {
	currentTime := time.Now()

	// Should match (created after 3 minutes ago)
	createdAfterTime := currentTime.Add(-3 * time.Minute)
	actualCreatedTime := currentTime.Add(-1 * time.Minute)
	predicate := CreatedAfter(createdAfterTime)
	listing := twigots.TicketListing{
		Event:     twigots.Event{Name: "test"},
		CreatedAt: twigots.UnixTime{Time: actualCreatedTime},
	}
	match := predicate(listing)
	require.True(t, match)

	createdAfterTime = currentTime.Add(-3 * time.Minute)
	actualCreatedTime = currentTime.Add(-2 * time.Minute)
	predicate = CreatedAfter(createdAfterTime)
	listing = twigots.TicketListing{
		Event:     twigots.Event{Name: "test"},
		CreatedAt: twigots.UnixTime{Time: actualCreatedTime},
	}
	match = predicate(listing)
	require.True(t, match)

	// Should not match (created before 3 minutes ago)
	createdAfterTime = currentTime.Add(-3 * time.Minute)
	actualCreatedTime = currentTime.Add(-4 * time.Minute)
	predicate = CreatedAfter(createdAfterTime)
	listing = twigots.TicketListing{
		Event:     twigots.Event{Name: "test"},
		CreatedAt: twigots.UnixTime{Time: actualCreatedTime},
	}
	match = predicate(listing)
	require.False(t, match)
}

func TestMaxTicketPriceInclFeePredicate(t *testing.T) {
	// Should match (total ticket incl fee price below £15)
	predicate := MaxTicketPriceInclFee(15)
	listing := twigots.TicketListing{
		Event:      twigots.Event{Name: "test"},
		NumTickets: 2,
		TotalPriceExclFee: twigots.Price{
			Currency: twigots.CurrencyGBP,
			Amount:   24 * 100, // £24 - £12 per ticket excl fee
		},
		TwicketsFee: twigots.Price{
			Currency: twigots.CurrencyGBP,
			Amount:   3 * 100, // £4 - £14 per ticket incl fee
		},
	}
	match := predicate(listing)
	require.True(t, match)

	// Should not match (total ticket price incl fee above £15)
	predicate = MaxTicketPriceInclFee(15)
	listing = twigots.TicketListing{
		Event:      twigots.Event{Name: "test"},
		NumTickets: 2,
		TotalPriceExclFee: twigots.Price{
			Currency: twigots.CurrencyGBP,
			Amount:   28 * 100, // £28 - £14 per ticket excl fee
		},
		TwicketsFee: twigots.Price{
			Currency: twigots.CurrencyGBP,
			Amount:   3 * 100, // £4 - £16 per ticket incl fee
		},
	}
	match = predicate(listing)
	require.False(t, match)
}
