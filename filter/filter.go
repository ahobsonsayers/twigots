package filter

import "github.com/ahobsonsayers/twigots"

// FilterTicketListings filters ticket listings to those that satisfy all of the provided predicates.
//
// If no predicates are provided, all listings are returned.
func FilterTicketListings( // nolint
	listings []twigots.TicketListing,
	predicates ...TicketListingPredicate,
) []twigots.TicketListing {
	if len(predicates) == 0 {
		return listings
	}

	result := make([]twigots.TicketListing, 0, len(listings))
	for _, listing := range listings {
		if TicketListingMatchesAllPredicates(listing, predicates...) {
			result = append(result, listing)
		}
	}

	return result
}

// TicketListingMatchesAllPredicates checks whether a ticket listing satisfies all of the predicates provided.
//
// Returns true if no predicates are provided.
func TicketListingMatchesAllPredicates(listing twigots.TicketListing, predicates ...TicketListingPredicate) bool {
	for _, predicate := range predicates {
		if !predicate(listing) {
			return false
		}
	}
	return true
}

// TicketListingMatchesAnyPredicate checks whether a ticket listing satisfies any of the predicates provided.
//
// Returns false if no predicates are provided.
func TicketListingMatchesAnyPredicate(listing twigots.TicketListing, predicates ...TicketListingPredicate) bool {
	for _, predicate := range predicates {
		if predicate(listing) {
			return true
		}
	}
	return false
}
