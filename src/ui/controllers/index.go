package controllers

// IndexController handles request to /
type IndexController struct {
	BaseController
}

// Get renders the index page
func (ic *IndexController) Get() {
	ic.Forward("page_title_index", "index.htm")
}
