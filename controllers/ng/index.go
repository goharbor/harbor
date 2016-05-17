package ng

type IndexController struct {
	BaseController
}

func (ic *IndexController) Get() {
	ic.Forward("Index", "index.htm")
}
