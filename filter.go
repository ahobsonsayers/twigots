package twigots

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/hbollon/go-edlib"
	"github.com/samber/lo"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const defaultSimilarity = 0.9

// A filter to use on ticket listing(s). A ticket listing can either match the filter or not.
//
// A ticket listing must satisfy all specified filter fields to match, making this an AND filter.
type Filter struct {
	// Name of event on ticket listing to match.
	// Required.
	Event string

	// Similarity of event name on ticket listing to the event name specified in the filter.
	// Specified as a float, between 0 and 1 (with 1 representing an exact match).
	// Leave this unset or set to <=0 to use the default.
	// Default is 0.85 which accounts for variances in event names.
	// e.g. Taylor Swift and Taylor Swift: The Eras Tour should match.
	EventSimilarity float64

	// Regions on ticket listings to match.
	// Leave this unset or empty to match any region.
	// Defaults to unset.
	Regions []Region

	// Number of tickets in a listing to match.
	// Leave this unset or set to <=0 to match any number of tickets.
	// Defaults to unset.
	NumTickets int

	// Minimum discount (including fee) of tickets in a listing to match.
	// Specified as a float, between 0 and 1 (with 1 representing 100% off).
	// Leave this unset or set to <=0 to match any discount (including no discount).
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

// Validate the filter.
// This is used internally to check a filter, but can also be used externally.
func (f Filter) Validate() error {
	if f.Event == "" {
		return errors.New("event name must be set")
	}

	if f.EventSimilarity > 1 {
		return errors.New("similarity cannot be > 1")
	}

	for _, region := range f.Regions {
		if !Regions.Contains(region) {
			return fmt.Errorf("region '%s' is not valid", region)
		}
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
	ticketEventName := normaliseString(listing.Event.Name)
	desiredEventName := normaliseString(eventName)

	ticketSimilarity := substringSimilarity(desiredEventName, ticketEventName)
	if similarity <= 0 {
		return ticketSimilarity >= defaultSimilarity
	}
	return ticketSimilarity >= similarity
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

var (
	// Text transformer to remove accents from strings
	accentTransformer = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	// Regular expression to match one or more whitespace characters
	spaceRegex = regexp.MustCompile(`\s+`)
	// Regular expression to match non-alphanumeric characters
	nonAlphaNumericRegex = regexp.MustCompile(`[^a-z0-9]`)
)

// normaliseString normalizes a given string by removing accents, converting to lowercase,
// removing leading/trailing whitespace, replacing '&' with 'and', and replacing special characters with spaces.
func normaliseString(eventName string) string {
	// TODO: This function could be improved.
	// TODO: The accent transformer does not currently support ł, Ł, ø, Æ, which will be removed.

	// Remove leading and trailing whitespace
	eventName = strings.TrimSpace(eventName)
	// Remove all accented characters
	eventName, _, _ = transform.String(accentTransformer, eventName)
	// Convert to lower case
	eventName = strings.ToLower(eventName)
	// Remove leading 'the'
	eventName = strings.TrimPrefix(eventName, "the ")
	// Replace '&' with 'and', ensuring spaces
	eventName = strings.ReplaceAll(eventName, "&", " and ")
	// Replace all special characters with spaces
	eventName = nonAlphaNumericRegex.ReplaceAllString(eventName, " ")
	// Replace all 2+ whitespaces with a single space
	eventName = spaceRegex.ReplaceAllString(eventName, " ")
	// Remove leading and trailing whitespace again
	eventName = strings.TrimSpace(eventName)

	return eventName
}

// Gap penalty in substring similarity calculation
const substringSimilarityGapPenalty = 1

// substringSimilarity calculates the similarity between a substring and a target string.
// The similarity is calculated using a modified Smith-Waterman local alignment algorithm to align the substring
// with the target string, and optimal string alignment Damerau-Levenshtein to calculate word level similarity.
//
// See:
// https://en.wikipedia.org/wiki/Smith%E2%80%93Waterman_algorithm
// https://en.wikipedia.org/wiki/Damerau%E2%80%93Levenshtein_distance#Optimal_string_alignment_distance
func substringSimilarity(subString, targetString string) float64 {
	// Split strings up into words
	subWords := strings.Fields(subString)
	targetWords := strings.Fields(targetString)

	numSubWords := len(subWords)
	numTargetWords := len(targetWords)

	// If both or one string has no words, exit early
	if numSubWords == 0 && numTargetWords == 0 {
		return 1
	}
	if numSubWords == 0 || numTargetWords == 0 {
		return 0
	}

	// Create a matrix (initialised with 0's) to store the similarity scores
	numRows := numSubWords + 1
	numCols := numTargetWords + 1
	matrix := make([][]float64, numRows)
	for i := range matrix {
		matrix[i] = make([]float64, numCols)
	}

	// Do similarity calculations
	for i := 1; i < numRows; i++ {
		for j := 1; j < numCols; j++ {
			similarity, err := edlib.StringsSimilarity(subWords[i-1], targetWords[j-1], edlib.DamerauLevenshtein)
			if err != nil {
				// An error will never occur if a valid similarity algorithm is used.
				// If an error does occur (due to an error in the code), panic so we catch it.
				panic(err)
			}

			// Calculate the match score
			matchScore := matrix[i-1][j-1] + float64(similarity)
			// Calculate the delete score (penalize missing words in substring)
			deleteScore := matrix[i-1][j] - substringSimilarityGapPenalty
			// Calculate the insert score (penalize additional words in substring)
			insertScore := matrix[i][j-1] - substringSimilarityGapPenalty

			// Store the maximum score in the matrix
			matrix[i][j] = maxUtil(0, matchScore, insertScore, deleteScore)
		}
	}

	// Find the maximum score in the last row (all substring words consumed)
	maxScore := maxUtil(matrix[numSubWords]...)

	// Return the average similarity across all words
	avgSimilarity := maxScore / float64(numSubWords)
	return avgSimilarity
}

func maxUtil(nums ...float64) float64 {
	maxNum := nums[0]
	for _, num := range nums[1:] {
		if num > maxNum {
			maxNum = num
		}
	}
	return maxNum
}
