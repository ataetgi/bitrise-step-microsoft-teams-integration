package main

// AdaptiveCard to send to Microsoft Teams.
type AdaptiveCard struct {
	Type    string              `json:"type"`
	Version string              `json:"version"`
	Body    []AdaptiveCardBody  `json:"body"`
	Actions []AdaptiveCardAction `json:"actions"`
}

// AdaptiveCardBody represents the body of the AdaptiveCard
type AdaptiveCardBody struct {
	Type     string              `json:"type"`
	Text     string              `json:"text,omitempty"`
	Size     string              `json:"size,omitempty"`
	Weight   string              `json:"weight,omitempty"`
	Color    string              `json:"color,omitempty"`
	Wrap     bool                `json:"wrap,omitempty"`
	IsSubtle bool                `json:"isSubtle,omitempty"`
	Items    []AdaptiveCardBody  `json:"items,omitempty"`
	URL      string              `json:"url,omitempty"`
	AltText  string              `json:"altText,omitempty"`
	Facts    []AdaptiveCardFact  `json:"facts,omitempty"`
}

// AdaptiveCardFact represents a fact in the AdaptiveCard
type AdaptiveCardFact struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

// AdaptiveCardAction represents an action in the AdaptiveCard
type AdaptiveCardAction struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	URL   string `json:"url"`
}
