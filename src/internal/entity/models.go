package entity

import "time"

const (
	DefaultTravelClass = "ECONOMY"
	DefaultCurrency    = "USD"
	DefaultAdults      = "1"
)

const (
	AmadeusProvider           = "Amadeus"
	SKyRapidProvider          = "Sky Rapid"
	GoogleFlightRapidProvider = "Google Flight Rapid"
)

type Flight struct {
	ProviderName    string `json:"provider_name,omitempty"`
	Price           float64
	DurationMinutes int       `json:"total_duration_minutes"`
	Segments        []Segment `json:"segments"`
}

type Segment struct {
	DepartureAirport   string `json:"departureAirport"`
	DestinationAirport string `json:"destinationAirport"`
	DepartureTime      string `json:"departureTime"`
	ArrivalTime        string `json:"arrivalTime"`
}

type FlightSearchParam struct {
	Origin        string
	Destination   string
	DateDeparture string
}

type FlightSearchResponse struct {
	Provider string   `json:"provider"`
	Currency string   `json:"currency"`
	Flights  []Flight `json:"flights"`
	//Error    error    `json:"error"`
	Cheapest Flight `json:"cheapest"`
	Fastest  Flight `json:"fastest"`
}

type FlightPriceResponse struct {
	OriginName       string                 `json:"originName"`
	DestinationName  string                 `json:"destinationName"`
	Cheapest         Flight                 `json:"cheapest"`
	Fastest          Flight                 `json:"fastest"`
	FlightByProvider []FlightSearchResponse `json:"flightByProvider"`
}

type FlightAmadeusResp struct {
	Data []FlightOffer `json:"data"`
}

type FlightSkyResp struct {
	Data DataSky `json:"data"`
}

type FlightGoogleResp struct {
	Data DataGoogle `json:"data"`
}

// FlightOffer represent amadeus response of a flight search
type FlightOffer struct {
	ID          string               `json:"id"`
	Itineraries []ItinerariesAmadeus `json:"itineraries"`
	Price       struct {
		Total    string `json:"total"`
		Currency string `json:"currency"`
	} `json:"price"`
}

type ItinerariesAmadeus struct {
	Duration string           `json:"duration"`
	Segments []SegmentAmadeus `json:"segments"`
}

type SegmentAmadeus struct {
	Departure struct {
		IataCode string `json:"iataCode"`
		At       string `json:"at"`
	} `json:"departure"`
	Arrival struct {
		IataCode string `json:"iataCode"`
		At       string `json:"at"`
	} `json:"arrival"`
	CarrierCode string `json:"carrierCode"`
}

// FlightItinerary represent flights-sky response of a flight search
type FlightItinerary struct {
	ID    string `json:"id"`
	Price struct {
		Amount float64 `json:"raw"`
	} `json:"price"`
	Legs []struct {
		Segments  []SegmentSky `json:"segments"`
		Duration  int          `json:"durationInMinutes"`
		Departure string       `json:"departure"`
		Arrival   string       `json:"arrival"`
	} `json:"legs"`
}

// OtherFlight represent flights google response of a flight search
type OtherFlight struct {
	Price    float64          `json:"price"`
	Duration int              `json:"duration"`
	Segments []SegmentGoogleF `json:"segments"`
	Stops    int              `json:"stops"`
}

type SegmentGoogleF struct {
	DepartureAirportName string `json:"departureAirportName"`
	ArrivalAirportName   string `json:"arrivalAirportName"`
	DepartureDate        string `json:"departureDate"`
	ArrivalDate          string `json:"arrivalDate"`
	DepartureTime        string `json:"departureTime"`
	ArrivalTime          string `json:"arrivalTime"`
}

type SegmentSky struct {
	Origin struct {
		Name string `json:"name"`
	} `json:"origin"`
	Destination struct {
		Name string `json:"name"`
	} `json:"destination"`
	DepartureDate string `json:"departure"`
	ArrivalDate   string `json:"arrival"`
}

type Provider struct {
	Name    string
	BaseURL string
	Apikey  string
	Secret  string
	Timeout time.Duration
}

type ProvConfig struct {
	Providers []Provider
}

type DataGoogle struct {
	OtherFlights []OtherFlight `json:"otherFlights"`
}

type DataSky struct {
	Itineraries []FlightItinerary `json:"itineraries"`
}
