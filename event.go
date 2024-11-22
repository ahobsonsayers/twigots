package twigots

// Event contains the details of an event.
type Event struct {
	Id       string `json:"id"`
	Name     string `json:"eventName"`
	Category string `json:"category"`

	Date      Date      `json:"date"`
	Time      Time      `json:"showStartingTime"`
	OnSale    *DateTime `json:"onSaleTime"` // 2023-11-17T10:00:00Z
	Announced *DateTime `json:"created"`    // 2023-11-17T10:00:00Z

	Venue  Venue    `json:"venue"`
	Lineup []Lineup `json:"participants"`
}

// Lineup contains the details of the event lineup.
type Lineup struct {
	Artist  Artist `json:"participant"`
	Billing int    `json:"billing"`
}

// Artist contains the details of an artist.
type Artist struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"linkName"`
}

// Venue contains the details of an event venue.
type Venue struct {
	Id       string   `json:"id"`
	Name     string   `json:"name"`
	Location Location `json:"location"`
	Postcode string   `json:"postcode"`
}

// Venue contains the details of an event location.
type Location struct {
	Id       string  `json:"id"`
	Name     string  `json:"shortName"`
	FullName string  `json:"name"`
	Country  Country `json:"countryCode"`
	Region   Region  `json:"regionCode"`
}

// Event contains the details of a tour.
type Tour struct {
	Id         string   `json:"tourId"`
	Name       string   `json:"tourName"`
	Slug       string   `json:"slug"`
	FirstEvent *Date    `json:"minDate"`      // 2024-06-06
	LastEvent  *Date    `json:"maxDate"`      // 2024-11-14
	Countries  []string `json:"countryCodes"` // TODO use enum - requires all countries to be added
}
