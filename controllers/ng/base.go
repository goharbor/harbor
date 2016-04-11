package ng

import (
	"path/filepath"

	"github.com/astaxie/beego"
)

type BaseController struct {
	beego.Controller
}

const (
	viewPath = "sections"
	prefixNg = "ng"
)

func (bc *BaseController) Forward(title, templateName string) {
	bc.Layout = filepath.Join(prefixNg, "layout.htm")
	bc.TplName = filepath.Join(prefixNg, templateName)
	bc.Data["Title"] = title
	bc.LayoutSections = make(map[string]string)
	bc.LayoutSections["HeaderInclude"] = filepath.Join(prefixNg, viewPath, "headerInclude.htm")
	bc.LayoutSections["FooterInclude"] = filepath.Join(prefixNg, viewPath, "footerInclude.htm")
	bc.LayoutSections["HeaderContent"] = filepath.Join(prefixNg, viewPath, "headerContent.htm")
	bc.LayoutSections["FooterContent"] = filepath.Join(prefixNg, viewPath, "footerContent.htm")

}
