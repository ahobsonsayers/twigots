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

var testEventNames = []string{
	"Coldplay",
	"The 1975",
	"Arctic Monkeys",
	"The Killers",
	"Imagine Dragons",
	"Panic! At The Disco",
	"Fall Out Boy",
	"Green Day",
	"Sum 41",
	"Blink-182",
}

func TestGetLatestTicketListingsReal(t *testing.T) {
	testutils.SkipIfCI(t)

	projectDirectory := testutils.ProjectDirectory(t)
	_ = godotenv.Load(filepath.Join(projectDirectory, ".env"))

	twicketsAPIKey := os.Getenv("TWICKETS_API_KEY")
	require.NotEmpty(t, twicketsAPIKey, "TWICKETS_API_KEY is not set")

	twicketsClient, err := twigots.NewClient(twicketsAPIKey)
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

func TestGetLatestTicketListings(t *testing.T) {
	testTime := time.Now()

	// Create client
	twicketsClient, err := twigots.NewClient(testAPIKey)
	require.NoError(t, err)

	// Setup mock
	url, responder := getMockUrlAndResponder(t, testEventNames, testTime, time.Minute)
	httpmock.ActivateNonDefault(twicketsClient.Client())
	httpmock.RegisterResponder("GET", url, responder)

	// Fetch ticket listings
	// This should return all 10 in the test feed response
	listings, err := twicketsClient.FetchTicketListings(
		context.Background(),
		twigots.FetchTicketListingsInput{
			Country:       twigots.CountryUnitedKingdom,
			MaxNumber:     10, // 10 is the default
			CreatedBefore: testTime,
		},
	)
	require.NoError(t, err)
	for i := 0; i < 5; i++ {
		require.Equal(t, testEventNames[i], listings[i].Event.Name)
	}
}

func TestGetLatestTicketListingsMaxNumber(t *testing.T) {
	testTime := time.Now()

	// Create client
	twicketsClient, err := twigots.NewClient(testAPIKey)
	require.NoError(t, err)

	// Setup mock
	url, responder := getMockUrlAndResponder(t, testEventNames[0:5], testTime, time.Minute)
	httpmock.ActivateNonDefault(twicketsClient.Client())
	httpmock.RegisterResponder("GET", url, responder)

	// Fetch ticket listings
	// This should return the first 5 in the test feed response
	listings, err := twicketsClient.FetchTicketListings(
		context.Background(),
		twigots.FetchTicketListingsInput{
			Country:       twigots.CountryUnitedKingdom,
			MaxNumber:     5,
			CreatedBefore: testTime,
		},
	)
	require.NoError(t, err)
	require.Len(t, listings, 5)
	for i := 0; i < 5; i++ {
		require.Equal(t, testEventNames[i], listings[i].Event.Name)
	}
}

func TestGetLatestTicketListingsCreatedAfter(t *testing.T) {
	testTime := time.Now()

	// Create client
	twicketsClient, err := twigots.NewClient(testAPIKey)
	require.NoError(t, err)

	// Setup mock
	url, responder := getMockUrlAndResponder(t, testEventNames, testTime, time.Minute)
	httpmock.ActivateNonDefault(twicketsClient.Client())
	httpmock.RegisterResponder("GET", url, responder)

	// Fetch ticket listings
	// This should return the first 5 in the test feed response
	listings, err := twicketsClient.FetchTicketListings(
		context.Background(),
		twigots.FetchTicketListingsInput{
			Country:       twigots.CountryUnitedKingdom,
			CreatedBefore: testTime,
			CreatedAfter:  testTime.Add(-5 * time.Minute),
		},
	)
	require.NoError(t, err)
	require.Len(t, listings, 5)
	for i := 0; i < 5; i++ {
		require.Equal(t, testEventNames[i], listings[i].Event.Name)
	}
}

// getMockUrlAndResponder returns a mock url and responder for testing purposes.
// The responder returns events spaced at the specified interval backwards from startTime.
func getMockUrlAndResponder(
	t *testing.T,
	events []string,
	startTime time.Time,
	interval time.Duration,
) (string, httpmock.Responder) {
	url := getMockUrl(events, startTime)

	responseJson := getMockResponseJson(t, events, startTime, interval)

	responder := func(_ *http.Request) (*http.Response, error) {
		response := httpmock.NewBytesResponse(http.StatusOK, responseJson)
		response.Header.Set("Content-Type", "application/json; charset=utf-8")
		return response, nil
	}

	return url, responder
}

func getMockUrl(events []string, startTime time.Time) string {
	return fmt.Sprintf(
		"https://www.twickets.live/services/catalogue?api_key=%s&count=%d&maxTime=%d&q=countryCode=%s",
		testAPIKey,
		len(events),
		startTime.UnixMilli(),
		twigots.CountryUnitedKingdom.Value,
	)
}

func getMockResponseJson(t *testing.T, events []string, startTime time.Time, interval time.Duration) []byte {
	// Create response.
	// All uneeded/unused fields have been stripped.
	// To see the real full feed response, see feelFeedResponse.json
	var responseListings []any
	for i, event := range events {
		id := rand.Int()
		createdAt := startTime.Add(-interval * time.Duration(i))

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

	// Create final response
	response := map[string]any{
		"responseData": responseListings,
	}

	// Marshal response
	responseJson, err := json.Marshal(response)
	require.NoError(t, err)

	return responseJson
}
