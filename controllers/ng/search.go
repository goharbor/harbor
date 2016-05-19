package ng

type SearchController struct {
	BaseController
}

func (sc *SearchController) Get() {
	sc.Forward("Search", "search.htm")
}
