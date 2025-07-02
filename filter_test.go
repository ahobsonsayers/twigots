package twigots

import (
	"testing"
	"time"

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

	listings := TicketListings{
		{Event: Event{Name: strangerThings}},
		{Event: Event{Name: backToTheFuture}},
		{Event: Event{Name: harryPotter}},
		{Event: Event{Name: theThe}},
	}

	// Stranger Things
	filteredListings, err := listings.Filter(Filter{
		Event: strangerThingsFilter,
	})
	require.NoError(t, err)
	require.Len(t, filteredListings, 1)
	require.Equal(t, strangerThings, filteredListings[0].Event.Name)

	// Back to the Future
	filteredListings, err = listings.Filter(Filter{
		Event: backToTheFutureFilter,
	})
	require.NoError(t, err)
	require.Len(t, filteredListings, 1)
	require.Equal(t, backToTheFuture, filteredListings[0].Event.Name)

	// Harry Potter
	filteredListings, err = listings.Filter(Filter{
		Event: harryPotterFilter,
	})
	require.NoError(t, err)
	require.Len(t, filteredListings, 1)
	require.Equal(t, harryPotter, filteredListings[0].Event.Name)

	// The Who
	filteredListings, err = listings.Filter(Filter{
		Event: theWhoFilter,
	})
	require.NoError(t, err)
	require.Empty(t, filteredListings)
}

func TestFilterCreatedAfter(t *testing.T) {
	event := Event{Name: "test"}
	currentTime := time.Now()
	listings := TicketListings{
		{
			Event:     event,
			CreatedAt: UnixTime{currentTime.Add(-1 * time.Minute)},
		},
		{
			Event:     event,
			CreatedAt: UnixTime{currentTime.Add(-2 * time.Minute)},
		},
		{
			Event:     event,
			CreatedAt: UnixTime{currentTime.Add(-4 * time.Minute)},
		},
		{
			Event:     event,
			CreatedAt: UnixTime{currentTime.Add(-5 * time.Minute)},
		},
	}

	filteredListings, err := listings.Filter(Filter{
		Event:        event.Name,
		CreatedAfter: currentTime.Add(-3 * time.Minute),
	})
	require.NoError(t, err)
	require.Equal(t, listings[0:2], filteredListings)
}

func TestSubstringSimilarity(t *testing.T) {
	subString := "taylor swift"
	targetString := "taylor swift the eras tour"
	similarity := substringSimilarity(subString, targetString)
	require.InDelta(t, 1, similarity, 0.001)

	subString = "taylor swift the eras tour"
	targetString = "taylor swift"
	similarity = substringSimilarity(subString, targetString)
	require.Less(t, similarity, 0.1)

	subString = "taylor swift"
	targetString = "miss americana a tribute to taylor swift"
	similarity = substringSimilarity(subString, targetString)
	require.InDelta(t, 1, similarity, 0.001)
}

func TestNormaliseString(t *testing.T) {
	// Test leading and trailing spaces get removed
	normalisedEventName := normaliseString(" Spaced  Out   ")
	expectedNormalisedEventName := "spaced out"
	require.Equal(t, expectedNormalisedEventName, normalisedEventName)

	// Test accents get replaced with with a-z alternative
	normalisedEventName = normaliseString("Über Gâteaux")
	expectedNormalisedEventName = "uber gateaux"
	require.Equal(t, expectedNormalisedEventName, normalisedEventName)

	// Test the prefix gets removed, and ',' get removed
	normalisedEventName = normaliseString("The Lion, the Witch and the Wardrobe")
	expectedNormalisedEventName = "lion the witch and the wardrobe"
	require.Equal(t, expectedNormalisedEventName, normalisedEventName)

	// Test a prefix of the without a space doesn't get removed
	normalisedEventName = normaliseString("here WE gO")
	expectedNormalisedEventName = "here we go"
	require.Equal(t, expectedNormalisedEventName, normalisedEventName)

	// Test removal of ':' and '.'
	normalisedEventName = normaliseString("This Is: A Colon.")
	expectedNormalisedEventName = "this is a colon"
	require.Equal(t, expectedNormalisedEventName, normalisedEventName)

	// Test replacements of '&' with 'and' regardless of spaces
	normalisedEventName = normaliseString("This & That& This &That&This")
	expectedNormalisedEventName = "this and that and this and that and this"
	require.Equal(t, expectedNormalisedEventName, normalisedEventName)

	// Test removal of 2+ spaces
	normalisedEventName = normaliseString("------gimme_____lines------")
	expectedNormalisedEventName = "gimme lines"
	require.Equal(t, expectedNormalisedEventName, normalisedEventName)
}
