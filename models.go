package main

import (
	"encoding/json"
	"time"
)

type Game struct {
	ID           string `json:"id"`
	Names        struct {
		International string `json:"international"`
		Japanese      string `json:"japanese"`
	} `json:"names"`
	Abbreviation string `json:"abbreviation"`
	Released     int    `json:"released"`
	Weblink      string `json:"weblink"`
}

type Category struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Rules   string `json:"rules"`
	Weblink string `json:"weblink"`
}

type Run struct {
	ID       string `json:"id"`
	Weblink  string `json:"weblink"`
	Game     string `json:"game"`
	Category string `json:"category"`
	Date     string `json:"date"`
	Submitted time.Time `json:"submitted"`
	Times    struct {
		Primary        string `json:"primary"`
		Realtime       string `json:"realtime"`
		RealtimeNoLoads string `json:"realtime_noloads"`
		Ingame         string `json:"ingame"`
	} `json:"times"`
	Players []struct {
		Rel  string `json:"rel"`
		ID   string `json:"id"`
		Name string `json:"name"`
		URI  string `json:"uri"`
	} `json:"players"`
	System struct {
		Platform  string `json:"platform"`
		Emulated  bool   `json:"emulated"`
		Region    string `json:"region"`
	} `json:"system"`
	Status struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	} `json:"status"`
	Videos struct {
		Text  string `json:"text"`
		Links []struct {
			URI string `json:"uri"`
		} `json:"links"`
	} `json:"videos"`
	Comment string `json:"comment"`
}

type Leaderboard struct {
	Weblink string `json:"weblink"`
	Game    struct {
		Data Game `json:"data"`
	} `json:"game"`
	Category struct {
		Data Category `json:"data"`
	} `json:"category"`
	Runs []struct {
		Place int `json:"place"`
		Run   Run `json:"run"`
	} `json:"runs"`
	PlatformMap map[string]string `json:"-"`
}

type APIResponse struct {
	Data       json.RawMessage `json:"data"`
	Pagination struct {
		Offset int `json:"offset"`
		Max    int `json:"max"`
		Size   int `json:"size"`
		Links  []struct {
			Rel string `json:"rel"`
			URI string `json:"uri"`
		} `json:"links"`
	} `json:"pagination"`
}

type User struct {
	ID    string `json:"id"`
	Names struct {
		International string `json:"international"`
		Japanese      string `json:"japanese"`
	} `json:"names"`
}

type Platform struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Released int    `json:"released"`
}

type Region struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Variable struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Values   struct {
		Values map[string]struct {
			Label string `json:"label"`
			Rules string `json:"rules"`
		} `json:"values"`
	} `json:"values"`
	IsSubcategory bool `json:"is-subcategory"`
}

type SubCategory struct {
	ID    string
	Label string
	Rules string
}

type APIError struct {
	Message    string
	StatusCode int
	URL        string
	Context    string
}

func (e APIError) Error() string {
	if e.Context != "" {
		return e.Context + ": " + e.Message
	}
	return e.Message
}