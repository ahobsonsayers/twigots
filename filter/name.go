package filter

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/ahobsonsayers/twigots"
	"github.com/hbollon/go-edlib"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	// Default event name similarity if one not specified
	DefaultEventNameSimilarity = 0.9

	// Gap penalty in substring similarity calculation
	substringSimilarityGapPenalty = 1
)

var (
	// Text transformer to remove accents from strings
	accentTransformer = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	// Regular expression to match one or more whitespace characters
	spaceRegex = regexp.MustCompile(`\s+`)
	// Regular expression to match non-alphanumeric characters
	nonAlphaNumericRegex = regexp.MustCompile(`[^a-z0-9]`)
)

// EventRegionIn creates a predicate that matches ticket listings with an event name matching the one specified.
//
// Similarity is a float between 0 and 1 (with 0 represent	ing no similarity and 1 representing an exact match).
//
// Set minimumSimilarity to <=0 to use the default of 0.9 which allows for minor variances in event names.
//
// If eventName is empty, any event name will match.
//
// If minimumSimilarity is set to >1, minimumSimilarity will be set to 1 (exact match only).
func EventName(eventName string, minimumSimilarity float64) TicketListingPredicate {
	// If no event name specified, match any event
	if eventName == "" {
		return alwaysPredicate
	}

	// Use default similarity if not specified or negative
	if minimumSimilarity <= 0 {
		minimumSimilarity = DefaultEventNameSimilarity
	}

	// Clamp similarity to maximum of 1.0
	if minimumSimilarity > 1 {
		minimumSimilarity = 1.0
	}

	return func(listing twigots.TicketListing) bool {
		// Normalise event names
		desiredEventName := normaliseString(eventName)
		listingEventName := normaliseString(listing.Event.Name)

		// Add spaces on either side of event name to help prevent
		// matches of word that is contained within another word
		desiredEventName = fmt.Sprintf(" %s ", desiredEventName)
		listingEventName = fmt.Sprintf(" %s ", listingEventName)

		eventSimilarity := substringSimilarity(desiredEventName, listingEventName)
		return eventSimilarity >= minimumSimilarity
	}
}

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

// substringSimilarity calculates the similarity between a substring and a target string.
// The similarity is calculated using a modified Smith-Waterman local alignment algorithm to align the substring
// with the target string, and optimal string alignment Damerau-Levenshtein to calculate word level similarity.
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
