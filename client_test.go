package twigots_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ahobsonsayers/twigots"
	"github.com/ahobsonsayers/utilopia/testutils"
	"github.com/davecgh/go-spew/spew"
	"github.com/jarcoal/httpmock"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"muzzammil.xyz/jsonc"
)

var (
	testApiKey      = "test"
	testCountry     = twigots.CountryUnitedKingdom
	testNumListings = 10 // Num listings per request - this is the default of 10
	testBeforeTime  = time.Date(2025, 1, 1, 12, 1, 0, 0, time.UTC)

	testBandNames = []string{
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
)

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
	// Create client
	twicketsClient, err := twigots.NewClient(testApiKey)
	require.NoError(t, err)

	// Setup mock
	mockUrl := getMockUrl()
	mockResponder := getMockResponder(t)
	httpmock.ActivateNonDefault(twicketsClient.Client())
	httpmock.RegisterResponder("GET", mockUrl, mockResponder)

	// Fetch ticket listings
	// This should return all 10 in the test feed response
	listings, err := twicketsClient.FetchTicketListings(
		context.Background(),
		twigots.FetchTicketListingsInput{
			Country:       twigots.CountryUnitedKingdom,
			MaxNumber:     10, // 10 is the default
			CreatedBefore: testBeforeTime,
		},
	)
	require.NoError(t, err)
	for i := 0; i < 5; i++ {
		require.Equal(t, testBandNames[i], listings[i].Event.Name)
	}
}

func TestGetLatestTicketListingsMaxNumber(t *testing.T) {
	// Create client
	twicketsClient, err := twigots.NewClient(testApiKey)
	require.NoError(t, err)

	// Setup mock
	mockUrl := getMockUrl()
	mockResponder := getMockResponder(t)
	httpmock.ActivateNonDefault(twicketsClient.Client())
	httpmock.RegisterResponder("GET", mockUrl, mockResponder)

	// Fetch ticket listings
	// This should return the first 5 in the test feed response
	listings, err := twicketsClient.FetchTicketListings(
		context.Background(),
		twigots.FetchTicketListingsInput{
			Country:       twigots.CountryUnitedKingdom,
			MaxNumber:     5,
			CreatedBefore: testBeforeTime,
		},
	)
	require.NoError(t, err)
	require.Len(t, listings, 5)
	for i := 0; i < 5; i++ {
		require.Equal(t, testBandNames[i], listings[i].Event.Name)
	}
}

func TestGetLatestTicketListingsCreateAfter(t *testing.T) {
	// Create client
	twicketsClient, err := twigots.NewClient(testApiKey)
	require.NoError(t, err)

	// Setup mock
	mockUrl := getMockUrl()
	mockResponder := getMockResponder(t)
	httpmock.ActivateNonDefault(twicketsClient.Client())
	httpmock.RegisterResponder("GET", mockUrl, mockResponder)

	// Fetch ticket listings
	// This should return the first 5 in the test feed response
	listings, err := twicketsClient.FetchTicketListings(
		context.Background(),
		twigots.FetchTicketListingsInput{
			Country:       twigots.CountryUnitedKingdom,
			CreatedBefore: testBeforeTime,
			CreatedAfter:  testBeforeTime.Add(-5 * 5 * time.Minute),
		},
	)
	require.NoError(t, err)
	require.Len(t, listings, 5)
	for i := 0; i < 5; i++ {
		require.Equal(t, testBandNames[i], listings[i].Event.Name)
	}
}

func getMockUrl() string {
	return fmt.Sprintf(
		"https://www.twickets.live/services/catalogue?api_key=%s&count=%d&maxTime=%d&q=countryCode=%s",
		testApiKey, testNumListings, testBeforeTime.UnixMilli(), testCountry.Value,
	)
}

func getMockResponder(t *testing.T) httpmock.Responder {
	return func(_ *http.Request) (*http.Response, error) {
		// Read test feed response jsonc
		testFeedResponsePath := testutils.ProjectDirectoryJoin(t, "test/data/testFeedResponse.jsonc")
		testFeedResponseJsonc, err := os.ReadFile(testFeedResponsePath)
		require.NoError(t, err)

		// Convert jsonc to json
		testFeedResponseJson := jsonc.ToJSON(testFeedResponseJsonc)

		// Create a new HTTP response with the JSON data
		response := httpmock.NewBytesResponse(http.StatusOK, testFeedResponseJson)
		response.Header.Set("Content-Type", "application/json; charset=utf-8")

		return response, nil
	}
}
