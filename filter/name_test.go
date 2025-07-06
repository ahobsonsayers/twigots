package filter

import (
	"testing"

	"github.com/ahobsonsayers/twigots"
	"github.com/stretchr/testify/require"
)

func TestEventNamePredicate(t *testing.T) {
	// Stranger Things should be an exact match even without : and a subtitle
	desiredEventName := "Stranger Things"
	actualEventName := "Stranger Things: The First Shadow"

	predicate := EventName(desiredEventName, 1)
	listing := twigots.TicketListing{Event: twigots.Event{Name: actualEventName}}

	match := predicate(listing)
	require.True(t, match)

	// Stranger Things should be an exact match even using and, without : and a subtitle
	desiredEventName = "Harry Potter and the Cursed Child"
	actualEventName = "Harry Potter & The Cursed Child Parts 1 & 2"

	predicate = EventName(desiredEventName, 1)
	listing = twigots.TicketListing{Event: twigots.Event{Name: actualEventName}}

	match = predicate(listing)
	require.True(t, match)

	// Oasish shouldn't match Oasis
	desiredEventName = "Oasis"
	actualEventName = "Oasish"

	predicate = EventName(desiredEventName, 0.9)
	listing = twigots.TicketListing{Event: twigots.Event{Name: actualEventName}}

	match = predicate(listing)
	require.False(t, match)

	// The The shouldn't match the Who
	desiredEventName = "The Who"
	actualEventName = "The The"

	predicate = EventName(desiredEventName, 0.9)
	listing = twigots.TicketListing{Event: twigots.Event{Name: actualEventName}}

	match = predicate(listing)
	require.False(t, match)
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
