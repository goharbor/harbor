/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package controllers

import (
	"net/http"
	"os"
	"strings"

	"github.com/astaxie/beego"
	"github.com/beego/i18n"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/utils/log"
)

// CommonController handles request from UI that doesn't expect a page, such as /login /logout ...
type CommonController struct {
	BaseController
}

// Render returns nil.
func (c *CommonController) Render() error {
	return nil
}

// BaseController wraps common methods such as i18n support, forward,  which can be leveraged by other UI render controllers.
type BaseController struct {
	beego.Controller
	i18n.Locale
}

type langType struct {
	Lang string
	Name string
}

const (
	defaultLang = "en-US"
)

var supportLanguages map[string]langType
var enableAddUserByAdmin bool
var isAdminLoginedUser bool

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

	sessionUserID := b.GetSession("userId")
	if sessionUserID != nil {
		b.Data["Username"] = b.GetSession("username")
		b.Data["UserId"] = sessionUserID.(int)
	}

	authMode := os.Getenv("AUTH_MODE")
	if authMode == "" {
		authMode = "db_auth"
	}
	b.Data["AuthMode"] = authMode

	selfRegistration := strings.ToLower(os.Getenv("SELF_REGISTRATION"))

	if selfRegistration == "off" {
		enableAddUserByAdmin = true
	}

	if sessionUserID != nil {
		var err error
		isAdminLoginedUser, err = dao.IsAdminRole(sessionUserID)
		if err != nil {
			log.Errorf("Error occurred in IsAdminRole:%v", err)
			b.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
	}

	b.Data["IsAdminLoginedUser"] = isAdminLoginedUser
	b.Data["EnableAddUserByAdmin"] = enableAddUserByAdmin

}

// ForwardTo setup layout and template for content for a page.
func (b *BaseController) ForwardTo(pageTitle string, pageName string) {
	b.Layout = "segment/base-layout.tpl"
	b.TplName = "segment/base-layout.tpl"
	b.Data["PageTitle"] = b.Tr(pageTitle)
	b.LayoutSections = make(map[string]string)
	b.LayoutSections["HeaderInc"] = "segment/header-include.tpl"
	b.LayoutSections["HeaderContent"] = "segment/header-content.tpl"
	b.LayoutSections["BodyContent"] = pageName + ".tpl"
	b.LayoutSections["ModalDialog"] = "segment/modal-dialog.tpl"
	b.LayoutSections["FootContent"] = "segment/foot-content.tpl"
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
