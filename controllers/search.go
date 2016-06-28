package controllers

// SearchController handles request to /search
type SearchController struct {
	BaseController
}

// Get rendlers search bar
func (sc *SearchController) Get() {
	sc.Forward("page_title_search", "search.htm")
}
