package twigots // nolint

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// TicketListing is a listing of ticket(s) on Twickets
type TicketListing struct {
	Id        string   `json:"blockId"`
	CreatedAt UnixTime `json:"created"`
	ExpiresAt UnixTime `json:"expires"`

	// Number of tickets in the listing
	NumTickets int `json:"ticketQuantity"`

	// TotalPriceExclFee is the total price of all tickets, excluding fee.
	// Use TotalPriceInclFee to get the total price of all tickets, including fee.
	// Use TicketPriceExclFee to get the price of a single ticket, excluding fee.
	// Use TicketPriceInclFee to get the price of a single ticket, including fee.
	TotalPriceExclFee Price `json:"totalSellingPrice"`

	// TwicketsFee is the total twickets fee for all tickets.
	// Use TwicketsFeePerTicket to get the twickets fee per ticket.
	TwicketsFee Price `json:"totalTwicketsFee"`

	// OriginalTotalPrice is the original total price of all tickets, including any fee.
	// Use OriginalTicketPrice to get the original price of a single ticket, including any fee.
	OriginalTotalPrice Price `json:"faceValuePrice"`

	SellerWillConsiderOffers bool `json:"sellerWillConsiderOffers"`

	// The type of the ticket e.g. seated, Standing, Box etc.
	TicketType   string `json:"priceTier"`
	SeatAssigned bool   `json:"seatAssigned"`
	Section      string `json:"section"` // Can be empty
	Row          string `json:"row"`     // Can be empty

	Event Event `json:"event"`
	Tour  Tour  `json:"tour"`
}

// URL of the ticket listing
//
// Format is: https://www.twickets.live/app/block/<ticketId>,<quanitity>
func (l TicketListing) URL() string {
	return fmt.Sprintf("https://www.twickets.live/app/block/%s,%d", l.Id, l.NumTickets)
}

// TicketPriceExclFee is price of a single ticket, excluding fee.
//
// Use TotalPriceExclFee to get the total price of all tickets, excluding fee.
//
// Use TotalPriceInclFee to get the total price of all tickets, including fee.
//
// Use TicketPriceInclFee to get the price of a single ticket, including fee.
func (l TicketListing) TicketPriceExclFee() Price {
	return l.TotalPriceExclFee.Divide(l.NumTickets).Add(l.TwicketsFeePerTicket())
}

// TotalPriceExclFee is the total price of all tickets, including fee.
//
// Use TotalPriceExclFee to get the total price of all tickets, excluding fee.
//
// Use TicketPriceExclFee to get the price of a single ticket, excluding fee.
//
// Use TicketPriceInclFee to get the price of a single ticket, including fee.
func (l TicketListing) TotalPriceInclFee() Price {
	return l.TotalPriceExclFee.Add(l.TwicketsFee)
}

// TicketPriceExclFee is price of a single ticket, including fee.
//
// Use TotalPriceExclFee to get the total price of all tickets, excluding fee.
//
// Use TotalPriceInclFee to get the total price of all tickets, including fee.
//
// Use TicketPriceExclFee to get the price of a single ticket, excluding fee.
func (l TicketListing) TicketPriceInclFee() Price {
	return l.TotalPriceInclFee().Divide(l.NumTickets)
}

// TwicketsFeePerTicket is the twickets fee per ticket.
//
// Use TwicketsFee to get the total fee for all tickets.
func (l TicketListing) TwicketsFeePerTicket() Price {
	return l.TwicketsFee.Divide(l.NumTickets)
}

// OriginalTotalPrice is the original price of a single ticket, including any fee.
//
// Use OriginalTotalPrice to get the original total price of all tickets, including any fee.
func (l TicketListing) OriginalTicketPrice() Price {
	return l.OriginalTotalPrice.Divide(l.NumTickets)
}

// Discount is the discount on the original price of a single ticket, including any fee.
//
// Discount is returned as a value between 0 and 1 (with 1 representing 100% off).
// If ticket is being sold at its original price, the addition of the twickets fee will
// cause discount to be < 0 i.e. the total ticket price will have gone up.
func (l TicketListing) Discount() float64 {
	return (1 - l.TotalPriceInclFee().Number()/l.OriginalTotalPrice.Number())
}

// DiscountString is the discount on the original price of a single ticket, including any fee
// as a percentage string between 0-100 with a % suffix.
//
// If ticket is being sold at its original price, the addition of the twickets fee will
// cause discount to be < 0% i.e. the total ticket price will have gone up. If this is the
// / case "none" will be returned instead of a negative percentage
func (l TicketListing) DiscountString() string {
	discount := l.Discount()
	if discount < 0 {
		return "none"
	}
	discountString := strconv.FormatFloat(discount*100, 'f', 2, 64)
	return discountString + "%"
}

// MatchesAnyFilter checks if a ticket listing matches any of the filters provided.
func (l TicketListing) MatchesAnyFilter(filters ...Filter) (bool, error) {
	// Validate filters
	for _, filter := range filters {
		err := filter.Validate()
		if err != nil {
			return false, fmt.Errorf("invalid filter: %w", err)
		}
	}

	return matchesAnyFilter(l, filters...), nil
}

// TicketListings is a slice of ticket listings.
type TicketListings []TicketListing

// GetById gets the first ticket listing with a matching id, or returns nil if one does not exist.
func (l TicketListings) GetById(id string) *TicketListing {
	for _, listing := range l {
		if listing.Id == id {
			return &listing
		}
	}
	return nil
}

// Filter ticket listings based on whether they match any of the filters provided.
func (l TicketListings) Filter(filters ...Filter) (TicketListings, error) {
	if len(filters) == 0 {
		return l, nil
	}

	// Validate filters
	for _, filter := range filters {
		err := filter.Validate()
		if err != nil {
			return nil, fmt.Errorf("invalid filter: %w", err)
		}
	}

	// Filter listings
	filteredListings := make(TicketListings, 0, len(l))
	for _, listing := range l {
		if matchesAnyFilter(listing, filters...) {
			filteredListings = append(filteredListings, listing)
		}
	}

	return filteredListings, nil
}

var spaceRegex = regexp.MustCompile(`\s+`)

func normaliseEventName(eventName string) string {
	eventName = strings.TrimSpace(eventName)
	eventName = spaceRegex.ReplaceAllString(eventName, " ")
	eventName = strings.ToLower(eventName)
	eventName = strings.TrimPrefix(eventName, "the ")
	return eventName
}

func UnmarshalTwicketsFeedJson(data []byte) ([]TicketListing, error) {
	response := struct {
		ResponseData []struct { // nolint
			Listings *TicketListing `json:"catalogBlockSummary"`
		} `json:"responseData"`
	}{}
	err := json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}

	// Get non null listings. Listings are null if they have been delisted
	listings := make([]TicketListing, 0, len(response.ResponseData))
	for _, responseData := range response.ResponseData {
		if responseData.Listings != nil {
			listings = append(listings, *responseData.Listings)
		}
	}

	return listings, nil
}
