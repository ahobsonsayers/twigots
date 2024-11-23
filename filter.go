package twigots

import (
	"errors"
	"fmt"
	"time"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/samber/lo"
)

const defaultSimilarity = 85.0

// A filter to use on ticket listing(s). A ticket listing can either match the filter or not.
// A ticket listing must satisfy all specified filter fields to match, making this an AND filter.
type Filter struct {
	// Name of event on ticket listing to match.
	// Required.
	Event string

	// Similarity of event name on ticket listing to the event name specified in the filter.
	// Specified as a float, between 0 and 1 (with 1 representing an exact match).
	// Defaults to 0.85 to account for variances in event names.
	// e.g. Taylor Swift and Taylor Swift: The Eras Tour should match.
	EventSimilarity float64

	// Regions on ticket listings to match.
	// Leave this unset or empty to match any region.
	// Defaults to unset.
	Regions []Region

	// Number of tickets in a listing to match.
	// Leave this unset or 0 to match any number of tickets.
	// Defaults to unset.
	NumTickets int

	// Minimum discount (including fee) of tickets in a listing to match.
	// Specified as a float, between 0 and 1 (with 1 representing 100% off).
	// Leave this unset or 0 to match any discount (including no discount).
	// Defaults to unset.
	MinDiscount float64

	// Time a listing must be created before to be match
	// Leave this unset to match any creation time.
	// Defaults to unset.
	CreatedBefore time.Time

	// Time a listing must be created after to be match
	// Leave this unset to match any creation time.
	// Defaults to unset.
	CreatedAfter time.Time
}

func (f Filter) validate() error {
	if f.Event == "" {
		return errors.New("event name must be set")
	}

	if f.EventSimilarity < 0 {
		return errors.New("similarity cannot be negative")
	} else if f.EventSimilarity > 1 {
		return errors.New("similarity cannot be > 1")
	}

	for _, region := range f.Regions {
		if !Regions.Contains(region) {
			return fmt.Errorf("region '%s' is not valid", region)
		}
	}

	if f.NumTickets < 0 {
		return errors.New("number of tickets cannot be negative")
	}

	if f.MinDiscount < 0 {
		return errors.New("discount cannot be negative")
	}
	if f.MinDiscount > 1 {
		return errors.New("discount cannot be > 1")
	}

	return nil
}

// matchesAnyFilter returns whether a ticket listing matches any of the filters provided.
// filters are assumed to have been validated first.
func matchesAnyFilter(listing TicketListing, filters ...Filter) bool {
	if len(filters) == 0 {
		return true
	}

	for _, filter := range filters {
		if matchesFilter(listing, filter) {
			return true
		}
	}

	return false
}

// matchesAnyFilter returns whether a ticket listing matches the filters provided.
// filter is assumed to have been validated first.
func matchesFilter(listing TicketListing, filter Filter) bool {
	return matchesEventName(listing, filter.Event, filter.EventSimilarity) &&
		matchesRegions(listing, filter.Regions) &&
		matchesNumTickets(listing, filter.NumTickets) &&
		matchesDiscount(listing, filter.MinDiscount) &&
		matchesCreatedBefore(listing, filter.CreatedBefore) &&
		matchesCreatedAfter(listing, filter.CreatedAfter)
}

// matchesEventName returns whether a ticket listing matches a desired event name
func matchesEventName(listing TicketListing, eventName string, similarity float64) bool {
	ticketEventName := normaliseEventName(listing.Event.Name)
	desiredEventName := normaliseEventName(eventName)

	ticketSimilarity := strutil.Similarity(
		ticketEventName, desiredEventName,
		metrics.NewJaroWinkler(),
	)
	if similarity == 0 {
		return ticketSimilarity >= defaultSimilarity/100
	}
	return ticketSimilarity >= similarity/100
}

// matchesRegions determines whether a ticket listing matches any desired regions.
func matchesRegions(listing TicketListing, regions []Region) bool {
	if len(regions) == 0 {
		return true
	}
	return lo.Contains(regions, listing.Event.Venue.Location.Region)
}

// matchesNumTickets determines whether a ticket listing matches a desired number of tickets
func matchesNumTickets(listing TicketListing, numTickets int) bool {
	if numTickets <= 0 {
		return true
	}
	return listing.NumTickets == numTickets
}

// matchesDiscount determines whether a ticket listing matches a desired discount.
func matchesDiscount(listing TicketListing, discount float64) bool {
	if discount <= 0 {
		return true
	}
	return listing.Discount() >= discount
}

// matchesCreatedBefore determines whether a ticket listing matches a desired created before time.
func matchesCreatedBefore(listing TicketListing, createdBefore time.Time) bool {
	if createdBefore.IsZero() {
		return true
	}
	return listing.CreatedAt.Time.Before(createdBefore)
}

// matchesCreatedAfter determines whether a ticket listing matches a desired created after time.
func matchesCreatedAfter(listing TicketListing, createdAfter time.Time) bool {
	if createdAfter.IsZero() {
		return true
	}
	return listing.CreatedAt.Time.After(createdAfter)
}
