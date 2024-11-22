package twigots_test

import (
	"testing"
	"time"

	"github.com/ahobsonsayers/twigots"
	"github.com/stretchr/testify/require"
)

func TestFilterEventName(t *testing.T) {
	// These should match
	strangerThings := "Stranger Things: The First Shadow"
	strangerThingsFilter := "Stranger Things"

	backToTheFuture := "Back To The Future: The Musical"
	backToTheFutureFilter := "Back To The Future"

	harryPotter := "Harry Potter & The Cursed Child Parts 1 & 2"
	harryPotterFilter := "Harry Potter and the Cursed Child"

	// These shouldn't match
	theThe := "The The"
	theWhoFilter := "The Who"

	listings := twigots.TicketListings{
		{Event: twigots.Event{Name: strangerThings}},
		{Event: twigots.Event{Name: backToTheFuture}},
		{Event: twigots.Event{Name: harryPotter}},
		{Event: twigots.Event{Name: theThe}},
	}

	// Stranger Things
	filteredListings := listings.Filter(twigots.Filter{
		Event: strangerThingsFilter,
	})
	require.Len(t, filteredListings, 1)
	require.Equal(t, strangerThings, filteredListings[0].Event.Name)

	// Back to the Future
	filteredListings = listings.Filter(twigots.Filter{
		Event: backToTheFutureFilter,
	})
	require.Len(t, filteredListings, 1)
	require.Equal(t, backToTheFuture, filteredListings[0].Event.Name)

	// Harry Potter
	filteredListings = listings.Filter(twigots.Filter{
		Event: harryPotterFilter,
	})
	require.Len(t, filteredListings, 1)
	require.Equal(t, harryPotter, filteredListings[0].Event.Name)

	// The Who
	filteredListings = listings.Filter(twigots.Filter{
		Event: theWhoFilter,
	})
	require.Empty(t, filteredListings)
}

func TestFilterCreatedAfter(t *testing.T) {
	currentTime := time.Now()
	listings := twigots.TicketListings{
		{CreatedAt: twigots.UnixTime{currentTime.Add(-1 * time.Minute)}},
		{CreatedAt: twigots.UnixTime{currentTime.Add(-2 * time.Minute)}},
		{CreatedAt: twigots.UnixTime{currentTime.Add(-4 * time.Minute)}},
		{CreatedAt: twigots.UnixTime{currentTime.Add(-5 * time.Minute)}},
	}

	filteredListings := listings.Filter(twigots.Filter{
		CreatedAfter: currentTime.Add(-3 * time.Minute),
	})
	require.Equal(t, listings[0:2], filteredListings)
}
