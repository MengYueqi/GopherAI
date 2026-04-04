package model

type TravelPlanPayload struct {
	Mode           string            `json:"mode,omitempty"`
	OverallSummary string            `json:"overall_summary,omitempty"`
	FlightPrice    TravelFlightPrice `json:"flight_price,omitempty"`
	DailyPlans     []TravelDayPlan   `json:"daily_plans,omitempty"`
	Sources        []string          `json:"sources,omitempty"`
	Notice         string            `json:"notice,omitempty"`
	RawText        string            `json:"raw_text,omitempty"`
}

type TravelFlightPrice struct {
	Summary     string   `json:"summary,omitempty"`
	Currency    string   `json:"currency,omitempty"`
	PriceRange  string   `json:"price_range,omitempty"`
	BookingTips []string `json:"booking_tips,omitempty"`
	RawText     string   `json:"raw_text,omitempty"`
}

type TravelDayPlan struct {
	Day         int                 `json:"day,omitempty"`
	Title       string              `json:"title,omitempty"`
	Route       string              `json:"route,omitempty"`
	Transport   string              `json:"transport,omitempty"`
	Summary     string              `json:"summary,omitempty"`
	Attractions []TravelAttraction  `json:"attractions,omitempty"`
	Tips        []string            `json:"tips,omitempty"`
}

type TravelAttraction struct {
	Name        string              `json:"name,omitempty"`
	Description string              `json:"description,omitempty"`
	Highlights  []string            `json:"highlights,omitempty"`
	Images      []TravelImageAsset  `json:"images,omitempty"`
}

type TravelImageAsset struct {
	Title     string `json:"title,omitempty"`
	URL       string `json:"url,omitempty"`
	Source    string `json:"source,omitempty"`
	SourceURL string `json:"source_url,omitempty"`
}
