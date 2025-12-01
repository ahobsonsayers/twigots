package twigots_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/ahobsonsayers/twigots"
	"github.com/ahobsonsayers/utilopia/testutils"
	"github.com/davecgh/go-spew/spew"
	"github.com/jarcoal/httpmock"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

const testAPIKey = "test"

var testEvents = []string{
	"Adele",
	"Arctic Monkeys",
	"Ariana Grande",
	"Bad Bunny",
	"Billie Eilish",
	"Blink-182",
	"Bruno Mars",
	"Coldplay",
	"Doja Cat",
	"Drake",
	"Dua Lipa",
	"Ed Sheeran",
	"Fall Out Boy",
	"Green Day",
	"Harry Styles",
	"Imagine Dragons",
	"Justin Bieber",
	"Olivia Rodrigo",
	"Panic! At The Disco",
	"Post Malone",
	"Sum 41",
	"Taylor Swift",
	"The 1975",
	"The Killers",
	"The Weeknd",
}

func TestFetchListingsReal(t *testing.T) {
	testutils.SkipIfCI(t)

	projectDirectory := testutils.ProjectDirectory(t)
	_ = godotenv.Load(filepath.Join(projectDirectory, ".env"))

	twicketsAPIKey := os.Getenv("TWICKETS_API_KEY")
	require.NotEmpty(t, twicketsAPIKey, "TWICKETS_API_KEY is not set")

	twicketsClient, err := twigots.NewClient(
		twicketsAPIKey,
		twigots.WithFlareSolverr("http://0.0.0.0:8191"),
	)
	require.NoError(t, err)

	// Fetch ticket listings
	listings, err := twicketsClient.FetchTicketListings(
		context.Background(),
		twigots.FetchTicketListingsInput{
			Country:       twigots.CountryUnitedKingdom,
			MaxNumber:     25,
			CreatedBefore: time.Now(),
		},
	)
	require.NoError(t, err)
	spew.Dump(listings)
	require.Len(t, listings, 25)
}

func TestFetchListings(t *testing.T) {
	testTime := time.Now().Truncate(time.Millisecond)

	// Create client
	twicketsClient, err := twigots.NewClient(testAPIKey)
	require.NoError(t, err)

	// Setup mock
	url, responder := getMockUrlAndResponder(t, testEvents[:10], testTime, time.Minute)
	httpmock.ActivateNonDefault(twicketsClient.Client())
	httpmock.RegisterResponder("GET", url, responder)

	// Fetch ticket listings
	// This should return all 10 in the test feed response
	listings, err := twicketsClient.FetchTicketListings(
		context.Background(),
		twigots.FetchTicketListingsInput{
			// By default gets 10 tickets
			Country:       twigots.CountryUnitedKingdom,
			CreatedBefore: testTime,
		},
	)
	require.NoError(t, err)
	require.Len(t, listings, 10)
	for i, listing := range listings {
		require.Equal(t, testEvents[i], listing.Event.Name)
	}
}

func TestFetchListingsPaginatedMaxNumber(t *testing.T) {
	testTime := time.Now().Truncate(time.Millisecond)

	// Create client
	twicketsClient, err := twigots.NewClient(testAPIKey)
	require.NoError(t, err)

	// Setup mock
	testEvents1 := testEvents[:10]
	testEvents2 := testEvents[10:20]
	testEvents3 := testEvents[20:23]

	testTime1 := testTime
	testTime2 := testTime1.Add(-10 * time.Minute)
	testTime3 := testTime2.Add(-10 * time.Minute)

	url1, responder1 := getMockUrlAndResponder(t, testEvents1, testTime1, time.Minute)
	url2, responder2 := getMockUrlAndResponder(t, testEvents2, testTime2, time.Minute)
	url3, responder3 := getMockUrlAndResponder(t, testEvents3, testTime3, time.Minute)

	httpmock.ActivateNonDefault(twicketsClient.Client())
	httpmock.RegisterResponder("GET", url1, responder1)
	httpmock.RegisterResponder("GET", url2, responder2)
	httpmock.RegisterResponder("GET", url3, responder3)

	// Fetch ticket listings
	// This should return all 10 in the test feed response
	listings, err := twicketsClient.FetchTicketListings(
		context.Background(),
		twigots.FetchTicketListingsInput{
			Country:       twigots.CountryUnitedKingdom,
			MaxNumber:     23,
			CreatedBefore: testTime,
		},
	)
	require.NoError(t, err)
	require.Len(t, listings, 23)
	for i, listing := range listings {
		require.Equal(t, testEvents[i], listing.Event.Name)
	}
}

func TestFetchListingsPaginatedCreatedAfter(t *testing.T) {
	testTime := time.Now().Truncate(time.Millisecond)

	// Create client
	twicketsClient, err := twigots.NewClient(testAPIKey)
	require.NoError(t, err)

	// Setup mock
	testEvents1 := testEvents[:10]
	testEvents2 := testEvents[10:20]

	testTime1 := testTime
	testTime2 := testTime1.Add(-10 * time.Minute)

	url1, responder1 := getMockUrlAndResponder(t, testEvents1, testTime1, time.Minute)
	url2, responder2 := getMockUrlAndResponder(t, testEvents2, testTime2, time.Minute)

	httpmock.ActivateNonDefault(twicketsClient.Client())
	httpmock.RegisterResponder("GET", url1, responder1)
	httpmock.RegisterResponder("GET", url2, responder2)

	// Fetch ticket listings
	// This should return the first 5 in the test feed response
	listings, err := twicketsClient.FetchTicketListings(
		context.Background(),
		twigots.FetchTicketListingsInput{
			Country:       twigots.CountryUnitedKingdom,
			MaxNumber:     100, // Large so we don't get limited by max number (default 10)
			CreatedBefore: testTime,
			CreatedAfter:  testTime.Add(-16 * time.Minute),
		},
	)
	require.NoError(t, err)
	require.Len(t, listings, 15)
	for i, listing := range listings {
		require.Equal(t, testEvents[i], listing.Event.Name)
	}
}

func TestFetchListingsCreatedAfterLatestListingReturnsNothing(t *testing.T) {
	testTime := time.Now().Truncate(time.Millisecond)

	// Create client
	twicketsClient, err := twigots.NewClient(testAPIKey)
	require.NoError(t, err)

	// Setup mock
	testEvents := testEvents[:10]
	url, responder := getMockUrlAndResponder(t, testEvents, testTime, time.Minute)
	httpmock.ActivateNonDefault(twicketsClient.Client())
	httpmock.RegisterResponder("GET", url, responder)

	// Fetch ticket listings
	// This should return no listings.
	listings, err := twicketsClient.FetchTicketListings(
		context.Background(),
		twigots.FetchTicketListingsInput{
			Country:       twigots.CountryUnitedKingdom,
			MaxNumber:     100, // Large so we don't get limited by max number (default 10)
			CreatedBefore: testTime,
			CreatedAfter:  testTime.Add(-time.Minute), // This should match latest ticket time
		},
	)
	require.NoError(t, err)
	require.Empty(t, listings)
}

// getMockUrlAndResponder returns a mock url and responder for testing purposes.
// The responder returns events spaced at the specified interval backwards from startTime.
func getMockUrlAndResponder(
	t *testing.T,
	events []string,
	startTime time.Time,
	interval time.Duration, // nolint:unparam
) (string, httpmock.Responder) {
	url := fmt.Sprintf(
		"https://www.twickets.live/services/catalogue?api_key=%s&count=10&maxTime=%d&q=countryCode=%s",
		testAPIKey,
		startTime.UnixMilli(),
		twigots.CountryUnitedKingdom.Value,
	)
	response := getMockResponse(events, startTime, interval)

	// Marshal response
	responseJson, err := json.Marshal(response)
	require.NoError(t, err)

	responder := func(_ *http.Request) (*http.Response, error) {
		response := httpmock.NewBytesResponse(http.StatusOK, responseJson)
		response.Header.Set("Content-Type", "application/json; charset=utf-8")
		return response, nil
	}

	return url, responder
}

func getMockResponse(events []string, startTime time.Time, interval time.Duration) map[string]any {
	// Create response.
	// All uneeded/unused fields have been stripped.
	// To see the real full feed response, see feelFeedResponse.json
	var responseListings []any
	for i, event := range events {
		id := rand.Int()
		createdAt := startTime.Add(-time.Duration(i+1) * interval)

		idString := strconv.Itoa(id)
		createdAtString := strconv.Itoa(int(createdAt.UnixMilli()))

		// Create event listing
		responseListing := map[string]any{
			"catalogBlockSummary": map[string]any{
				"blockId": idString,
				"created": createdAtString,
				"event": map[string]any{
					"id":        idString,
					"eventName": event,
				},
			},
		}
		responseListings = append(responseListings, responseListing)

		// Add a delisted listing after every second listing for testing
		if (i+1)%2 == 0 {
			delistedListing := map[string]any{"catalogBlockSummary": nil}
			responseListings = append(responseListings, delistedListing)
		}
	}

	// Return final response
	return map[string]any{
		"responseData": responseListings,
	}
}
