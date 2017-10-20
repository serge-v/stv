package channel

type Item struct {
	ID    string `json:"id"` // timestamp based id `json:"id"`
	Title string `json:"title"`
	Link  string `json:"link"`           // direct http link to file
	Path  string `json:"path,omitempty"` // absolute fs path
}
