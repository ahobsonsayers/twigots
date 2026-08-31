![twigots](assets/twigots.png)

[![Go Reference](https://pkg.go.dev/badge/github.com/ahobsonsayers/twigots.svg)](https://pkg.go.dev/github.com/ahobsonsayers/twigots)
[![Go Report Card](https://goreportcard.com/badge/github.com/ahobsonsayers/twigots)](https://goreportcard.com/report/github.com/ahobsonsayers/twigots)
[![License - MIT](https://img.shields.io/badge/License-MIT-9C27B0)](LICENSE)
[![Artisan README - Not LLM](https://img.shields.io/static/v1?label=Artisan+README&message=Not+LLM&labelColor=37474F&color=D97757)](#arnl---artisan-readme-not-llm)

> [!NOTE]
> We're back! 💪
>
> After Twickets introduced measures to prevent unofficial access to their data - this project was broken for quite a while.
>
> After a significant amount of time and effort tinkering and hitting my head against a wall (more than I care to admit) I have now found a way to bring this project back to life and get it working again!
>
> Enjoy! 🎟️

A Go package to fetch ticket listings from the [Twickets](https://www.twickets.live) Live Feed.

Includes utilities to help filter the ticket listings and find the ones you want!

Powers (the similarly creatively named)
[twitchets](https://github.com/ahobsonsayers/twitchets), a tool to watch for event ticket listings on Twickets and notify you so you can snap them up! 🫰

This package utilises the API used by the Twickets Android app, built via decompiling and reverse engineering its code.

To use this API (and therefore this package), you will need to obtain API keys. See the [Getting Keys](#getting-keys) section

- [Installation](#installation)
- [Getting Keys](#getting-keys)
	- [Loading Keys](#loading-keys)
- [Example Usage](#example-usage)
- [How does the event name matching/similarity work?](#how-does-the-event-name-matchingsimilarity-work)
	- [Normalization](#normalization)
- [Why the name twigots?](#why-the-name-twigots)
- [AR;NL - Artisan Readme; Not LLM](#arnl---artisan-readme-not-llm)

## Installation

```bash
go get -u github.com/ahobsonsayers/twigots
```

## Getting Keys

This package uses the API used by the Android app. To use this API, three keys are required (as well as a valid `User-Agent`):

- `api_key`
- `x-prosopo-site-key`
- `x-prosopo-android-integrity-token`

These must be supplied to the package via a `keys.json` file that looks like:

```json
{
  "api_key": "your_api_key",
  "User-Agent": "your_user_agent",
  "x-prosopo-site-key": "your_site_key",
  "x-prosopo-android-integrity-token": "your_integrity_token"
}
```

The first two keys are static, but the third is rotated and can only be obtained from a real (or emulated) Android device, on a scheduled basis.

Thankfully I built the [`twickets-key-extractor`](https://github.com/ahobsonsayers/twickets-key-extractor) project to do exactly this with an emulator, and regularly extract these keys.

You can run this yourself, but I have also set up a [public, regularly extracted keys.json here](https://gist.githubusercontent.com/ahobsonsayers/773acb763aafc8a39ac260e12a9b39d5/raw/keys.json)

### Loading Keys

Once you have a valid source for the `keys.json` you can load them for use in this package in two ways, both of which allow hot reloading

Load keys from a URL:

```go
keys, err := keys.LoadKeysFromURL("https://example.com/keys.json")
```

Load keys from a file:

```go
keys, err := keys.LoadKeysFromFile("/path/to/keys.json")
```

## Example Usage

> [!WARNING]
> Although this package is functional and ready for use, it is still a work in progress and is subject to change without notice - the API and usage may be modified at any time.
>
> Use with caution and check for updates regularly.

Example can be seen in [`example/main.go`](example/main.go) or below:

```go
package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/ahobsonsayers/twigots"
	"github.com/ahobsonsayers/twigots/filter"
	"github.com/ahobsonsayers/twigots/keys"
)

func main() {
	keysURL := os.Getenv("TWICKETS_KEYS_URL")
	if keysURL == "" {
		log.Fatal("TWICKETS_KEYS_URL is not set")
	}

	twicketsKeys, err := keys.LoadKeysFromURL(keysURL)
	if err != nil {
		log.Fatal(err)
	}

	// Create twickets client
	client, err := twigots.NewClient(twicketsKeys)
	if err != nil {
		log.Fatal(err)
	}

	// Fetch ticket listings
	listings, err := client.FetchTicketListings(
		context.Background(),
		twigots.FetchTicketListingsInput{
			// Required
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

Event name similarity is calculated using a modified [Smith-Waterman-Gotoh algorithm](https://en.wikipedia.org/wiki/Smith%E2%80%93Waterman_algorithm). The complexity behind this algorithm does not need to be understood, but for all intents and purposes, it can be thought of as fuzzy substring matching.

If the desired event name appears within the actual event name returned by twickets (as a substring), the event similarity will be 1. Equally, if the desired event name does not appear at all, the similarity will be 0.

Setting a required similarity below, but close to 1, will allow for small inconsistencies due to misspellings etc., but can return false positives. We recommend (and default to) a value of `0.9`.

False positives can also occur if your desired event name appears in the actual event name, but the event is not the one you want. This can often happen with things like tribute bands - see the example below.

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
Similarity score: 1 <- This is an exact match, but it is probably not the event we want
```

For a more in-depth explanation of the string matching algorithm, [see this PR](https://github.com/ahobsonsayers/twigots/pull/2).

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

Because it's a stupid mash up of Tickets and Go... and also why not?

[![Hits](https://hits.sh/github.com/ahobsonsayers/twigots.svg?view=today-total&label=Visitors%20Day%20%2F%20Total)](https://hits.sh/github.com/ahobsonsayers/twigots/)

## AR;NL - Artisan Readme; Not LLM

In the age of LLMs and coding agents, code is now cheap - for better or for worse. Your time however, is not ⌛

Therefore this project, like most of my projects, uses a hand written "artisan" README to ensure it is clear, correct and concise. This makes it easy to read and in my opinion encourages reading and engagement - no one likes AI slop!

As someone wiser than me once told a colleague:

"if you can't be bothered to take the time to write these words, then why should I be bothered to read them"

Enjoy!
