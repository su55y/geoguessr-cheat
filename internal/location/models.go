package location

type Address struct {
	Hamlet       string `json:"hamlet,omitempty"`
	Borough      string `json:"borough,omitempty"`
	Municipality string `json:"municipality,omitempty"`
	County       string `json:"county,omitempty"`
	Province     string `json:"province,omitempty"`
	Suburb       string `json:"suburb,omitempty"`
	Village      string `json:"village,omitempty"`
	Town         string `json:"town,omitempty"`
	City         string `json:"city,omitempty"`
	StateDist    string `json:"state_district,omitempty"`
	State        string `json:"state,omitempty"`
	Country      string `json:"country"`
}

type Location struct {
	Lat  string  `json:"lat"`
	Lon  string  `json:"lon"`
	Name string  `json:"display_name"`
	Addr Address `json:"address"`
}

type Coords struct {
	Lat string
	Lng string
}
