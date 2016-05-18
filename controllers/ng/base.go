package ng

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/astaxie/beego"
	"github.com/beego/i18n"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/utils/log"
)

type BaseController struct {
	beego.Controller
	i18n.Locale
	SelfRegistration bool
	IsAdmin          bool
	AuthMode         string
}

type langType struct {
	Lang string
	Name string
}

const (
	viewPath    = "sections"
	prefixNg    = "ng"
	defaultLang = "en-US"
)

var supportLanguages map[string]langType

// Prepare extracts the language information from request and populate data for rendering templates.
func (b *BaseController) Prepare() {

	var lang string
	al := b.Ctx.Request.Header.Get("Accept-Language")

	if len(al) > 4 {
		al = al[:5] // Only compare first 5 letters.
		if i18n.IsExist(al) {
			lang = al
		}
	}

	if _, exist := supportLanguages[lang]; exist == false { //Check if support the request language.
		lang = defaultLang //Set default language if not supported.
	}

	sessionLang := b.GetSession("lang")
	if sessionLang != nil {
		b.SetSession("Lang", lang)
		lang = sessionLang.(string)
	}

	curLang := langType{
		Lang: lang,
	}

	restLangs := make([]*langType, 0, len(langTypes)-1)
	for _, v := range langTypes {
		if lang != v.Lang {
			restLangs = append(restLangs, v)
		} else {
			curLang.Name = v.Name
		}
	}

	// Set language properties.
	b.Lang = lang
	b.Data["Lang"] = curLang.Lang
	b.Data["CurLang"] = curLang.Name
	b.Data["RestLangs"] = restLangs

	authMode := strings.ToLower(os.Getenv("AUTH_MODE"))
	if authMode == "" {
		authMode = "db_auth"
	}
	b.AuthMode = authMode
	b.Data["AuthMode"] = b.AuthMode

	selfRegistration := strings.ToLower(os.Getenv("SELF_REGISTRATION"))

	if selfRegistration == "on" {
		b.SelfRegistration = true
	}

	sessionUserID := b.GetSession("userId")
	if sessionUserID != nil {
		b.Data["Username"] = b.GetSession("username")
		b.Data["UserId"] = sessionUserID.(int)

		var err error
		b.IsAdmin, err = dao.IsAdminRole(sessionUserID.(int))
		if err != nil {
			log.Errorf("Error occurred in IsAdminRole:%v", err)
			b.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
	}

	b.Data["IsAdmin"] = b.IsAdmin
	b.Data["SelfRegistration"] = b.SelfRegistration

}

func (bc *BaseController) Forward(title, templateName string) {
	bc.Layout = filepath.Join(prefixNg, "layout.htm")
	bc.TplName = filepath.Join(prefixNg, templateName)
	bc.Data["Title"] = title
	bc.LayoutSections = make(map[string]string)
	bc.LayoutSections["HeaderInclude"] = filepath.Join(prefixNg, viewPath, "header-include.htm")
	bc.LayoutSections["FooterInclude"] = filepath.Join(prefixNg, viewPath, "footer-include.htm")
	bc.LayoutSections["HeaderContent"] = filepath.Join(prefixNg, viewPath, "header-content.htm")
	bc.LayoutSections["FooterContent"] = filepath.Join(prefixNg, viewPath, "footer-content.htm")

}

var langTypes []*langType

func init() {

	//conf/app.conf -> os.Getenv("config_path")
	configPath := os.Getenv("CONFIG_PATH")
	if len(configPath) != 0 {
		log.Infof("Config path: %s", configPath)
		beego.AppConfigPath = configPath
		if err := beego.ParseConfig(); err != nil {
			log.Warningf("Failed to parse config file: %s, error: %v", configPath, err)
		}
	}

	beego.AddFuncMap("i18n", i18n.Tr)

	langs := strings.Split(beego.AppConfig.String("lang::types"), "|")
	names := strings.Split(beego.AppConfig.String("lang::names"), "|")

	supportLanguages = make(map[string]langType)

	langTypes = make([]*langType, 0, len(langs))
	for i, v := range langs {
		t := langType{
			Lang: v,
			Name: names[i],
		}
		langTypes = append(langTypes, &t)
		supportLanguages[v] = t
	}

	for _, lang := range langs {
		if err := i18n.SetMessage(lang, "static/i18n/"+"locale_"+lang+".ini"); err != nil {
			log.Errorf("Fail to set message file: %s", err.Error())
		}
	}
}
