package twigots_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/ahobsonsayers/twigots"
	"github.com/ahobsonsayers/utilopia/testutils"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalFeedJson(t *testing.T) {
	listings := testTicketListings(t)

	require.Len(t, listings, 4)

	require.Equal(t, "Foo Fighters", listings[0].Event.Name)
	require.Len(t, listings[0].Event.Lineup, 3)
	require.Equal(t, "Foo Fighters", listings[0].Event.Lineup[0].Artist.Name)
	require.Equal(t, "Wet Leg", listings[0].Event.Lineup[1].Artist.Name)
	require.Equal(t, "Shame", listings[0].Event.Lineup[2].Artist.Name)
	require.Equal(t, "London Stadium", listings[0].Event.Venue.Name)
	require.Equal(t, "Foo Fighters - Everything Or Nothing At All Tour", listings[0].Tour.Name)
	require.Equal(t, 3, listings[0].NumTickets)
	require.Equal(t, "£180.00", listings[0].TotalPriceExclFee.String())
	require.Equal(t, "£38.25", listings[0].TwicketsFee.String())
	require.Equal(t, "£255.00", listings[0].OriginalTotalPrice.String())

	require.Equal(t, "Mean Girls", listings[1].Event.Name)
	require.Empty(t, listings[1].Event.Lineup)
	require.Equal(t, "Savoy Theatre", listings[1].Event.Venue.Name)
	require.Equal(t, "Mean Girls", listings[1].Tour.Name)
	require.Equal(t, 2, listings[1].NumTickets)
	require.Equal(t, "£130.00", listings[1].TotalPriceExclFee.String())
	require.Equal(t, "£18.20", listings[1].TwicketsFee.String())
	require.Equal(t, "£130.00", listings[1].OriginalTotalPrice.String())

	require.Equal(t, "South Africa v Wales", listings[2].Event.Name)
	require.Empty(t, listings[2].Event.Lineup)
	require.Equal(t, "Twickenham Stadium", listings[2].Event.Venue.Name)
	require.Equal(t, "South Africa v Wales", listings[2].Tour.Name)
	require.Equal(t, 4, listings[2].NumTickets)
	require.Equal(t, "£380.00", listings[2].TotalPriceExclFee.String())
	require.Equal(t, "£53.20", listings[2].TwicketsFee.String())
	require.Equal(t, "£380.00", listings[2].OriginalTotalPrice.String())

	require.Equal(t, "Download Festival 2024", listings[3].Event.Name)
	require.Empty(t, listings[3].Event.Lineup)
	require.Equal(t, "Donington Park", listings[3].Event.Venue.Name)
	require.Equal(t, "Download Festival 2024", listings[3].Tour.Name)
	require.Equal(t, 1, listings[3].NumTickets)
	require.Equal(t, "£280.00", listings[3].TotalPriceExclFee.String())
	require.Equal(t, "£30.80", listings[3].TwicketsFee.String())
	require.Equal(t, "£322.00", listings[3].OriginalTotalPrice.String())
}

func TestTicketListingsGetById(t *testing.T) {
	listings := testTicketListings(t)
	ticket := listings.GetById("156783487261837")
	require.NotNil(t, ticket)
}

func TestTicketListingDiscountString(t *testing.T) {
	tickets := testTicketListings(t)
	discountString := tickets[0].DiscountString()
	require.Equal(t, "14.41%", discountString)
}

func testTicketListings(t *testing.T) twigots.TicketListings {
	projectDirectory := testutils.ProjectDirectory(t)
	feedJsonFilePath := filepath.Join(projectDirectory, "test", "data", "fullFeedResponse.json")

	feedJsonFile, err := os.Open(feedJsonFilePath)
	require.NoError(t, err)
	feedJson, err := io.ReadAll(feedJsonFile)
	require.NoError(t, err)

	tickets, err := twigots.UnmarshalTwicketsFeedJson(feedJson)
	require.NoError(t, err)

	return tickets
}
