package ng

type IndexController struct {
	BaseController
}

func (c *IndexController) Get() {
	c.Forward("Index", "index.htm")
}
