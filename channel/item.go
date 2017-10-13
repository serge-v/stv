package channel

type Item struct {
	ID    string // timestamp based id
	Title string
	Link  string // direct http link to file
	Path  string // absolute fs path
}
