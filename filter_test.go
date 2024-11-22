package twigots_test

import (
	"testing"
	"time"

	"github.com/ahobsonsayers/twigots"
	"github.com/stretchr/testify/require"
)

func TestFilterTicketListingsName(t *testing.T) {
	strangerThingsAsked := "Stranger Things"
	strangerThingsGot := "Stranger Things: The First Shadow"

	backToTheFutureAsked := "Back To The Future"
	backToTheFutureGot := "Back To The Future: The Musical"

	harryPotterAsked := "Harry Potter and the Cursed Child"
	harryPotterGot := "Harry Potter & The Cursed Child Parts 1 & 2"

	wizardOfOzAsked := "The Who"
	wizardOfOzGot := "The The" // This shouldn't match

	gotTickets := twigots.TicketListings{
		{Event: twigots.Event{Name: strangerThingsGot}},
		{Event: twigots.Event{Name: backToTheFutureGot}},
		{Event: twigots.Event{Name: harryPotterGot}},
		{Event: twigots.Event{Name: wizardOfOzGot}},
	}

	// Stranger Things
	filteredListings := gotTickets.Filter(twigots.Filter{
		Event: strangerThingsAsked,
	})
	require.Len(t, filteredListings, 1)
	require.Equal(t, strangerThingsGot, filteredListings[0].Event.Name)

	// Back to the Future
	filteredListings = gotTickets.Filter(twigots.Filter{
		Event: backToTheFutureAsked,
	})
	require.Len(t, filteredListings, 1)
	require.Equal(t, backToTheFutureGot, filteredListings[0].Event.Name)

	// Harry Potter
	filteredListings = gotTickets.Filter(twigots.Filter{
		Event: harryPotterAsked,
	})
	require.Len(t, filteredListings, 1)
	require.Equal(t, harryPotterGot, filteredListings[0].Event.Name)

	// Wizard of Oz
	filteredListings = gotTickets.Filter(twigots.Filter{
		Event: wizardOfOzAsked,
	})
	require.Empty(t, filteredListings)
}

func TestFilterTicketListingsCreatedAfter(t *testing.T) {
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
