package filter

import (
	"time"

	"github.com/ahobsonsayers/twigots"
	"github.com/samber/lo"
)

// TicketListingPredicate is a predicate function that evaluates a TicketListing
// and returns whether or not the listing satisfies a condition.
type TicketListingPredicate func(twigots.TicketListing) bool

// EventRegion creates a predicate that matches ticket listings with an event in any of the specified regions.
//
// Invalid regions will be ignored.
//
// If regions is empty, or all regions are invali,d any region will match.
func EventRegion(regions ...twigots.Region) TicketListingPredicate {
	// Filter out invalid regions
	validRegions := make([]twigots.Region, 0, len(regions))
	for _, region := range regions {
		if twigots.Regions.Contains(region) {
			validRegions = append(validRegions, region)
		}
	}

	// If no valid regions specified, match any region
	if len(validRegions) == 0 {
		return alwaysPredicate
	}

	return func(listing twigots.TicketListing) bool {
		return lo.Contains(validRegions, listing.Event.Venue.Location.Region)
	}
}

// NumTickets creates a predicate that matches ticket listings with the specified number of tickets.
//
// Set numTickets to <=0 to match any number of tickets.
func NumTickets(numTickets int) TicketListingPredicate {
	// If no specific number specified, match any number
	if numTickets <= 0 {
		return alwaysPredicate
	}

	return func(listing twigots.TicketListing) bool {
		return listing.NumTickets == numTickets
	}
}

// ListingNumTickets creates a predicate that matches ticket listings with at least the specified discount.
//
// Discount is specified as a float, between 0 and 1 (with 0 representing no discount and 1 representing 100% off).
//
// Set minDiscount to <=0 to match any discount (including no discount).
//
// If minDiscount is set to >1, minDiscount will be set to 1 (100% discount only).
func MinDiscount(minDiscount float64) TicketListingPredicate {
	// Use no minimum discount if not specified or negative
	if minDiscount <= 0 {
		return alwaysPredicate
	}

	// Clamp discount to maximum of 1.0
	if minDiscount > 1 {
		minDiscount = 1.0
	}

	return func(listing twigots.TicketListing) bool {
		return listing.Discount() >= minDiscount
	}
}

// CreatedBefore creates a predicate that matches ticket listings created before the specified time.
//
// If createdBefore is zero time, any creation time will match.
func CreatedBefore(createdBefore time.Time) TicketListingPredicate {
	// If no time specified, match any creation time
	if createdBefore.IsZero() {
		return alwaysPredicate
	}

	return func(listing twigots.TicketListing) bool {
		return listing.CreatedAt.Before(createdBefore)
	}
}

// CreatedAfter creates a predicate that matches ticket listings created after the specified time.
//
// If createdAfter is zero time, any creation time will match.
func CreatedAfter(createdAfter time.Time) TicketListingPredicate {
	// If no time specified, match any creation time
	if createdAfter.IsZero() {
		return alwaysPredicate
	}

	return func(listing twigots.TicketListing) bool {
		return listing.CreatedAt.After(createdAfter)
	}
}

func alwaysPredicate(_ twigots.TicketListing) bool { return true }
