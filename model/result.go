package model

type Result struct {
	Title    string `json:"t"`
	URL      string `json:"u"`
	Snippet  string `json:"s"`
	Category string `json:"c"`
	Engine   string `json:"e"`
	Body     string `json:"b,omitempty"`
}
