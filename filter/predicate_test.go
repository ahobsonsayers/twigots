package filter

import (
	"testing"
	"time"

	"github.com/ahobsonsayers/twigots"
	"github.com/stretchr/testify/require"
)

func TestCreatedAfterPredicate(t *testing.T) {
	currentTime := time.Now()

	// Should match (created after 3 minutes ago)
	createdAfterTime := currentTime.Add(-3 * time.Minute)
	actualCreatedTime := currentTime.Add(-1 * time.Minute)
	predicate := CreatedAfter(createdAfterTime)
	listing := twigots.TicketListing{
		Event:     twigots.Event{Name: "test"},
		CreatedAt: twigots.UnixTime{Time: actualCreatedTime},
	}
	match := predicate(listing)
	require.True(t, match)

	createdAfterTime = currentTime.Add(-3 * time.Minute)
	actualCreatedTime = currentTime.Add(-2 * time.Minute)
	predicate = CreatedAfter(createdAfterTime)
	listing = twigots.TicketListing{
		Event:     twigots.Event{Name: "test"},
		CreatedAt: twigots.UnixTime{Time: actualCreatedTime},
	}
	match = predicate(listing)
	require.True(t, match)

	// Should not match (created before 3 minutes ago)
	createdAfterTime = currentTime.Add(-3 * time.Minute)
	actualCreatedTime = currentTime.Add(-4 * time.Minute)
	predicate = CreatedAfter(createdAfterTime)
	listing = twigots.TicketListing{
		Event:     twigots.Event{Name: "test"},
		CreatedAt: twigots.UnixTime{Time: actualCreatedTime},
	}
	match = predicate(listing)
	require.False(t, match)
}
