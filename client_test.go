package twigots_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/ahobsonsayers/twigots"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

// TODO: Use httptest client

func TestGetLatestTicketListings(t *testing.T) {
	// t.Skip(t, "Does not work on CI atm. Fix this.")

	projectDirectory := projectDirectory(t)
	_ = godotenv.Load(filepath.Join(projectDirectory, ".env"))

	twicketsAPIKey := os.Getenv("TWICKETS_API_KEY")
	require.NotEmpty(t, twicketsAPIKey, "TWICKETS_API_KEY is not set")

	twicketsClient := twigots.NewClient()
	listings, err := twicketsClient.FetchTicketListings(
		context.Background(),
		twigots.FetchTicketListingsInput{
			APIKey:  twicketsAPIKey,
			Country: twigots.CountryUnitedKingdom,
			Regions: []twigots.Region{
				twigots.RegionLondon,
				twigots.RegionNorthWest,
			},
			MaxNumber: 10,
		},
	)
	require.NoError(t, err)
	require.Len(t, listings, 10)
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
