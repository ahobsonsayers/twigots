![twigots](assets/twigots.png)

[![Go Reference](https://pkg.go.dev/badge/github.com/ahobsonsayers/twigots.svg)](https://pkg.go.dev/github.com/ahobsonsayers/twigots)
[![Go Report
Card](https://goreportcard.com/badge/github.com/ahobsonsayers/twigots)](https://goreportcard.com/report/github.com/ahobsonsayers/twigots)
[![License - MIT](https://img.shields.io/badge/License-MIT-9C27B0)](LICENSE)

A go package to fetch ticket listings from the [Twickets](https://www.twickets.live) Live Feed.

Includes utilities to help filter the ticket listings and find the ones you want!

Powers (the similarly creatively named)
[twitchets](https://github.com/ahobsonsayers/twitchets), a tool to watch for event ticket listings on Twickets and notify you so you can snap them up! 🫰

- [Installation- Installation](#installation--installation)
- [Getting an API Key](#getting-an-api-key)
- [Example Usage](#example-usage)
- [How does the event name matching/similarity work?](#how-does-the-event-name-matchingsimilarity-work)
	- [Normalization](#normalization)
- [Why the name twigots?](#why-the-name-twigots)

## Installation- [Installation](#installation)

```bash
go get -u github.com/ahobsonsayers/twigots
```

## Getting an API Key

To use this tool, you will need a Twickets API key. Although Twickets doesn't provide a free API, you can easily obtain a key by following these steps:

1.  Visit the [Twickets Live Feed](https://www.twickets.live/app/catalog/browse)
2.  Open your browser's Developer Tools (F12) and navigate to the Network tab
3.  Look for the GET request to `https://www.twickets.live/services/catalogue` and copy the `api_key` query parameter. You might need to refresh the page first if nothing appears in this tab.

This API key is not provided here due to liability concerns, but it appears to be a fixed, unchanging value.

## Example Usage

Example can be seen in [`example/main.go`](example/main.go) or below:

```go
package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/ahobsonsayers/twigots"
	"github.com/ahobsonsayers/twigots/filter"
)

func main() {
	apiKey := "my_api_key"

	// Fetch ticket listings
	client := twigots.NewClient() // Or use a custom http client
	listings, err := client.FetchTicketListings(
		context.Background(),
		twigots.FetchTicketListingsInput{
			// Required
			APIKey:  apiKey,
			Country: twigots.CountryUnitedKingdom, // Only UK is supported at the moment
			// Optional. See all options in godoc
			CreatedBefore: time.Now(),
			CreatedAfter:  time.Now().Add(time.Duration(-5 * time.Minute)), // 5 mins ago
			MaxNumber:     100,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Fetched %d ticket listings", len(listings))

	// Filter ticket listings just by name
	// Use the default event name similarity (0.9) to allow minor mismatches
	hamiltonListings := filter.FilterTicketListings(
		listings,
		filter.EventName("Hamilton", filter.DefaultEventNameSimilarity),
	)
	for _, listing := range hamiltonListings {
		slog.Info(
			"Found Hamilton ticket listing",
			"Event", listing.Event.Name,
			"NumTickets", listing.NumTickets,
			"Price", listing.TotalPriceInclFee().String(),
			"OriginalPrice", listing.OriginalTicketPrice().String(),
			"URL", listing.URL(),
		)
	}

	// Filter ticket listings just by several filters
	coldplayListings := filter.FilterTicketListings(
		listings,
		filter.EventName("Coldplay", 1), // Event name similarity of 1 - exact match only
		filter.EventRegion( // Only in specific regions
			twigots.RegionLondon,
			twigots.RegionSouth,
		),
		filter.NumTickets(2),    // Exactly 2 tickets in the listing
		filter.MinDiscount(0.1), // Discount of > 10%
	)
	for _, listing := range coldplayListings {
		slog.Info(
			"Found Coldplay ticket listing",
			"Event", listing.Event.Name,
			"NumTickets", listing.NumTickets,
			"Price", listing.TotalPriceInclFee().String(),
			"OriginalPrice", listing.OriginalTicketPrice().String(),
			"URL", listing.URL(),
		)
	}
}
```

## How does the event name matching/similarity work?

Event name similarity is calculated using the [Smith-Waterman-Gotoh algorithm](https://en.wikipedia.org/wiki/Smith%E2%80%93Waterman_algorithm). The complexity behind this algorithm does not need to be understood, but for all intents and purposes, it can be thought of as fuzzy substring matching.

If the desired event name appears within the actual event name returned by twickets (as a substring), the event similarity will be 1. Equally, if the desired event name does not appear at all, the similarity will be 0.

Setting a required similarity below, but close to 1, will allow for small inconsistencies due to misspelling etc., but can return false positives. We recommend (and default to) a value of `0.9`.

False positives can also occur if your desired event name appears in the actual event name, but the event is not the one you want. This can often happen with things like tribute bands - see the example below.

> [!NOTE]
> This is actually bidirectional substring matching, so either the desired event name or the actual event name can be the substring. This is important to take into account. See note below the false positive example for more info.

**Example:**

```
Desired event: Taylor Swift
Actual event: Taylor Swift: The Eras Tour
Similarity score: 1
```

**Example of a false positive:**

```
Desired event: Taylor Swift
Actual event: Miss Americana: A Tribute to Taylor Swift
Similarity score: 1 <- This is a exact match, but it is probably not the event we want
```

> [!NOTE]
> In the false positive example above, as the substring matching is bidirectional, if you were looking explicitly for the **tribute** event, and an actual event called **Taylor Swift** event came up, the event similarity would still be 1.

### Normalization

To help with matching, both the desired and actual event names are normalized before similarity is calculated.

This is done by:

- Converting to lower case
- Removing all symbols/non-alphanumeric characters (except **&** - see below)
- Replacing all **&** symbols with **and**
- Removing any **the** prefix
- Trimming leading and trailing whitespace and replacing all 2+ whitespace with a single space
- Replacing accented characters with their non-accented characters
- Spaces are added to either side of the string, to help avoid cases where the word appears inside another word e.g. grate shouldn't match un*grate*ful

## Why the name twigots?

Because its a stupid mash up of Tickets and Go... and also why not?

[![Hits](https://hits.sh/github.com/ahobsonsayers/twigots.svg?view=today-total&label=Visitors%20Day%20%2F%20Total)](https://hits.sh/github.com/ahobsonsayers/twigots/)
