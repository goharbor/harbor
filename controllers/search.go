package controllers

// SearchController handles request to /search
type SearchController struct {
	BaseController
}

// Get rendlers search bar
func (sc *SearchController) Get() {
	sc.Forward("Search", "search.htm")
}
