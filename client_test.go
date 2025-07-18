package twigots_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/ahobsonsayers/twigots"
	"github.com/ahobsonsayers/utilopia/testutils"
	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func TestClientWithProxy(t *testing.T) {
	testutils.SkipIfCI(t)

	projectDirectory := projectDirectory(t)
	_ = godotenv.Load(filepath.Join(projectDirectory, ".env"))

	twicketsAPIKey := os.Getenv("TWICKETS_API_KEY")
	require.NotEmpty(t, twicketsAPIKey, "TWICKETS_API_KEY is not set")
	proxyUser := os.Getenv("PROXY_USER")
	require.NotEmpty(t, proxyUser, "PROXY_USER is not set")
	proxyPassword := os.Getenv("PROXY_PASSWORD")
	require.NotEmpty(t, proxyPassword, "PROXY_PASSWORD is not set")
	proxyHost := os.Getenv("PROXY_HOST")
	require.NotEmpty(t, proxyHost, "PROXY_HOST is not set")
	proxyPort := os.Getenv("PROXY_PORT")
	require.NotEmpty(t, proxyPort, "PROXY_PORT is not set")
	proxyPortInt, err := strconv.Atoi(proxyPort)
	require.NoError(t, err, "failed to convert PROXY_PORT to int")
	proxy, err := twigots.NewProxy(proxyHost, proxyPortInt, proxyUser, proxyPassword)
	require.NoError(t, err)
	require.NoError(t, err)

	twicketsClient, err := twigots.NewClient(twicketsAPIKey, []twigots.Proxy{*proxy})
	require.NoError(t, err)

	listings, err := twicketsClient.FetchTicketListings(
		context.Background(),
		twigots.FetchTicketListingsInput{
			Country:   twigots.CountryUnitedKingdom,
			MaxNumber: 10,
		},
	)
	require.NoError(t, err)
	spew.Dump(listings)
}

// TODO: Use httptest client

func TestGetLatestTicketListings(t *testing.T) {
	testutils.SkipIfCI(t)

	projectDirectory := projectDirectory(t)
	_ = godotenv.Load(filepath.Join(projectDirectory, ".env"))

	twicketsAPIKey := os.Getenv("TWICKETS_API_KEY")
	require.NotEmpty(t, twicketsAPIKey, "TWICKETS_API_KEY is not set")

	twicketsClient, err := twigots.NewClient(twicketsAPIKey, nil)
	require.NoError(t, err)

	listings, err := twicketsClient.FetchTicketListings(
		context.Background(),
		twigots.FetchTicketListingsInput{
			Country:   twigots.CountryUnitedKingdom,
			MaxNumber: 10,
		},
	)
	require.NoError(t, err)
	spew.Dump(listings)
}

func projectDirectory(t *testing.T) string {
	workingDirectory, err := os.Getwd()
	if err != nil {
		require.NoError(t, err, "failed to get path of current working directory")
	}

	directory := workingDirectory
	for directory != "/" {
		_, err := os.Stat(filepath.Join(directory, "go.mod"))
		if err == nil {
			break
		}
		directory = filepath.Dir(directory)
	}
	require.NotEqual(t, "failed find project directory", directory, "/")

	return directory
}

func testTicketListings(t *testing.T) twigots.TicketListings {
	projectDirectory := projectDirectory(t)
	feedJsonFilePath := filepath.Join(projectDirectory, "testdata", "feed.json")

	feedJsonFile, err := os.Open(feedJsonFilePath)
	require.NoError(t, err)
	feedJson, err := io.ReadAll(feedJsonFile)
	require.NoError(t, err)

	tickets, err := twigots.UnmarshalTwicketsFeedJson(feedJson)
	require.NoError(t, err)

	return tickets
}
