package ng

// SearchController handles request to ng/search
type SearchController struct {
	BaseController
}

// Get rendlers search bar
func (sc *SearchController) Get() {
	sc.Forward("Search", "search.htm")
}
