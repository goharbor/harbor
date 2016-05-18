package ng

type AdminOptionController struct {
	BaseController
}

func (aoc *AdminOptionController) Get() {
	aoc.Forward("Admin Options", "admin-options.htm")
}
