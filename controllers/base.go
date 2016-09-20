package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/astaxie/beego"
	"github.com/beego/i18n"
	"github.com/vmware/harbor/auth"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

// BaseController wraps common methods such as i18n support, forward,  which can be leveraged by other UI render controllers.
type BaseController struct {
	beego.Controller
	i18n.Locale
	SelfRegistration bool
	IsAdmin          bool
	AuthMode         string
	UseCompressedJS  bool
}

type langType struct {
	Lang string
	Name string
}

const (
	viewPath    = "sections"
	prefixNg    = ""
	defaultLang = "en-US"
)

var supportLanguages map[string]langType
var mappingLangNames map[string]string

// Prepare extracts the language information from request and populate data for rendering templates.
func (b *BaseController) Prepare() {

	var lang string

	langCookie, err := b.Ctx.Request.Cookie("language")
	if err != nil {
		log.Errorf("Error occurred in Request.Cookie: %v", err)
	}
	if langCookie != nil {
		lang = langCookie.Value
	}
	if len(lang) == 0 {
		sessionLang := b.GetSession("lang")
		if sessionLang != nil {
			b.SetSession("Lang", lang)
			lang = sessionLang.(string)
		} else {
			al := b.Ctx.Request.Header.Get("Accept-Language")
			if len(al) > 4 {
				al = al[:5] // Only compare first 5 letters.
				if i18n.IsExist(al) {
					lang = al
				}
			}
		}
	}

	if _, exist := supportLanguages[lang]; !exist { //Check if support the request language.
		lang = defaultLang //Set default language if not supported.
	}

	b.Ctx.SetCookie("language", lang, 0, "/")
	b.SetSession("Lang", lang)

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

	useCompressedJS := os.Getenv("USE_COMPRESSED_JS")
	if useCompressedJS == "on" {
		b.UseCompressedJS = true
	}

	if _, err := os.Stat(filepath.Join("static", "resources", "js", "harbor.app.min.js")); os.IsNotExist(err) {
		b.UseCompressedJS = false
	}

	selfRegistration := strings.ToLower(os.Getenv("SELF_REGISTRATION"))
	if selfRegistration == "on" {
		b.SelfRegistration = true
	}
	b.Data["SelfRegistration"] = b.SelfRegistration
}

// Forward to setup layout and template for content for a page.
func (b *BaseController) Forward(title, templateName string) {
	b.Layout = filepath.Join(prefixNg, "layout.htm")
	b.TplName = filepath.Join(prefixNg, templateName)
	b.Data["Title"] = b.Tr(title)
	b.LayoutSections = make(map[string]string)
	b.LayoutSections["HeaderInclude"] = filepath.Join(prefixNg, viewPath, "header-include.htm")

	if b.UseCompressedJS {
		b.LayoutSections["HeaderScriptInclude"] = filepath.Join(prefixNg, viewPath, "script-min-include.htm")
	} else {
		b.LayoutSections["HeaderScriptInclude"] = filepath.Join(prefixNg, viewPath, "script-include.htm")
	}

	log.Debugf("Loaded HeaderScriptInclude file: %s", b.LayoutSections["HeaderScriptInclude"])

	b.LayoutSections["FooterInclude"] = filepath.Join(prefixNg, viewPath, "footer-include.htm")
	b.LayoutSections["HeaderContent"] = filepath.Join(prefixNg, viewPath, "header-content.htm")
	b.LayoutSections["FooterContent"] = filepath.Join(prefixNg, viewPath, "footer-content.htm")

}

var langTypes []*langType

// CommonController handles request from UI that doesn't expect a page, such as /SwitchLanguage /logout ...
type CommonController struct {
	BaseController
}

// Render returns nil.
func (cc *CommonController) Render() error {
	return nil
}

// Login handles login request from UI.
func (cc *CommonController) Login() {
	principal := cc.GetString("principal")
	password := cc.GetString("password")

	user, err := auth.Login(models.AuthModel{
		Principal: principal,
		Password:  password,
	})
	if err != nil {
		log.Errorf("Error occurred in UserLogin: %v", err)
		cc.CustomAbort(http.StatusUnauthorized, "")
	}

	if user == nil {
		cc.CustomAbort(http.StatusUnauthorized, "")
	}

	cc.SetSession("userId", user.UserID)
	cc.SetSession("username", user.Username)
}

// LogOut Habor UI
func (cc *CommonController) LogOut() {
	cc.DestroySession()
}

// SwitchLanguage User can swith to prefered language
func (cc *CommonController) SwitchLanguage() {
	lang := cc.GetString("lang")
	hash := cc.GetString("hash")
	if _, exist := supportLanguages[lang]; !exist {
		lang = defaultLang
	}
	cc.SetSession("lang", lang)
	cc.Data["Lang"] = lang
	cc.Redirect(cc.Ctx.Request.Header.Get("Referer")+hash, http.StatusFound)
}

// UserExists checks if user exists when user input value in sign in form.
func (cc *CommonController) UserExists() {
	target := cc.GetString("target")
	value := cc.GetString("value")

	user := models.User{}
	switch target {
	case "username":
		user.Username = value
	case "email":
		user.Email = value
	}

	exist, err := dao.UserExists(user, target)
	if err != nil {
		log.Errorf("Error occurred in UserExists: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	cc.Data["json"] = exist
	cc.ServeJSON()
}

func init() {

	//conf/app.conf -> os.Getenv("config_path")
	configPath := os.Getenv("CONFIG_PATH")
	if len(configPath) != 0 {
		log.Infof("Config path: %s", configPath)
		beego.LoadAppConfig("ini", configPath)
	}

	beego.AddFuncMap("i18n", i18n.Tr)

	langs := strings.Split(beego.AppConfig.String("lang::types"), "|")
	names := strings.Split(beego.AppConfig.String("lang::names"), "|")

	supportLanguages = make(map[string]langType)

	langTypes = make([]*langType, 0, len(langs))

	for i, lang := range langs {
		t := langType{
			Lang: lang,
			Name: names[i],
		}
		langTypes = append(langTypes, &t)
		supportLanguages[lang] = t
		if err := i18n.SetMessage(lang, "static/i18n/"+"locale_"+lang+".ini"); err != nil {
			log.Errorf("Fail to set message file: %s", err.Error())
		}
	}

}
