package controllers

import (
	"net/http"
	"net/http/httptest"
	//"net/url"
	"path/filepath"
	"runtime"
	"testing"

	"fmt"
	"strings"

	"github.com/astaxie/beego"
	//"github.com/dghubble/sling"
	"github.com/stretchr/testify/assert"
)

//const (
//	adminName = "admin"
//	adminPwd  = "Harbor12345"
//)

//type usrInfo struct {
//	Name   string
//	Passwd string
//}

//var admin *usrInfo

func init() {

	_, file, _, _ := runtime.Caller(1)
	apppath, _ := filepath.Abs(filepath.Dir(filepath.Join(file, ".."+string(filepath.Separator))))
	beego.BConfig.WebConfig.Session.SessionOn = true
	beego.TestBeegoInit(apppath)
	beego.AddTemplateExt("htm")

	beego.Router("/", &IndexController{})
	beego.Router("/dashboard", &DashboardController{})
	beego.Router("/project", &ProjectController{})
	beego.Router("/repository", &RepositoryController{})
	beego.Router("/sign_up", &SignUpController{})
	beego.Router("/add_new", &AddNewController{})
	beego.Router("/account_setting", &AccountSettingController{})
	beego.Router("/change_password", &ChangePasswordController{})
	beego.Router("/admin_option", &AdminOptionController{})
	beego.Router("/forgot_password", &ForgotPasswordController{})
	beego.Router("/reset_password", &ResetPasswordController{})
	beego.Router("/search", &SearchController{})

	beego.Router("/login", &CommonController{}, "post:Login")
	beego.Router("/log_out", &CommonController{}, "get:LogOut")
	beego.Router("/reset", &CommonController{}, "post:ResetPassword")
	beego.Router("/userExists", &CommonController{}, "post:UserExists")
	beego.Router("/sendEmail", &CommonController{}, "get:SendEmail")
	beego.Router("/language", &CommonController{}, "get:SwitchLanguage")

	beego.Router("/optional_menu", &OptionalMenuController{})
	beego.Router("/navigation_header", &NavigationHeaderController{})
	beego.Router("/navigation_detail", &NavigationDetailController{})
	beego.Router("/sign_in", &SignInController{})

	//Init user Info
	//admin = &usrInfo{adminName, adminPwd}

}

// TestMain is a sample to run an endpoint test
func TestMain(t *testing.T) {
	assert := assert.New(t)

	//	v := url.Values{}
	//	v.Set("principal", "admin")
	//	v.Add("password", "Harbor12345")

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(200), w.Code, "'/' httpStatusCode should be 200")
	assert.Equal(true, strings.Contains(fmt.Sprintf("%s", w.Body), "<title>page_title_index</title>"), "http respond should have '<title>page_title_index</title>'")

	r, _ = http.NewRequest("GET", "/dashboard", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(200), w.Code, "'/dashboard' httpStatusCode should be 200")
	assert.Equal(true, strings.Contains(fmt.Sprintf("%s", w.Body), "<title>page_title_dashboard</title>"), "http respond should have '<title>page_title_dashboard</title>'")

	r, _ = http.NewRequest("GET", "/project", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(200), w.Code, "'/project' httpStatusCode should be 200")
	assert.Equal(true, strings.Contains(fmt.Sprintf("%s", w.Body), "<title>page_title_project</title>"), "http respond should have '<title>page_title_project</title>'")

	r, _ = http.NewRequest("GET", "/repository", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(200), w.Code, "'/repository' httpStatusCode should be 200")
	assert.Equal(true, strings.Contains(fmt.Sprintf("%s", w.Body), "<title>page_title_repository</title>"), "http respond should have '<title>page_title_repository</title>'")

	r, _ = http.NewRequest("GET", "/sign_up", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(200), w.Code, "'/sign_up' httpStatusCode should be 200")
	assert.Equal(true, strings.Contains(fmt.Sprintf("%s", w.Body), "<title>page_title_sign_up</title>"), "http respond should have '<title>page_title_sign_up</title>'")

	r, _ = http.NewRequest("GET", "/add_new", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(401), w.Code, "'/add_new' httpStatusCode should be 401")

	r, _ = http.NewRequest("GET", "/account_setting", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(200), w.Code, "'/account_setting' httpStatusCode should be 200")
	assert.Equal(true, strings.Contains(fmt.Sprintf("%s", w.Body), "<title>page_title_account_setting</title>"), "http respond should have '<title>page_title_account_setting</title>'")

	r, _ = http.NewRequest("GET", "/change_password", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(200), w.Code, "'/change_password' httpStatusCode should be 200")
	assert.Equal(true, strings.Contains(fmt.Sprintf("%s", w.Body), "<title>page_title_change_password</title>"), "http respond should have '<title>page_title_change_password</title>'")

	r, _ = http.NewRequest("GET", "/admin_option", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(200), w.Code, "'/admin_option' httpStatusCode should be 200")
	assert.Equal(true, strings.Contains(fmt.Sprintf("%s", w.Body), "<title>page_title_admin_option</title>"), "http respond should have '<title>page_title_admin_option</title>'")

	r, _ = http.NewRequest("GET", "/forgot_password", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(200), w.Code, "'/forgot_password' httpStatusCode should be 200")
	assert.Equal(true, strings.Contains(fmt.Sprintf("%s", w.Body), "<title>page_title_forgot_password</title>"), "http respond should have '<title>page_title_forgot_password</title>'")

	r, _ = http.NewRequest("GET", "/reset_password", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(302), w.Code, "'/reset_password' httpStatusCode should be 302")

	r, _ = http.NewRequest("GET", "/search", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(200), w.Code, "'/search' httpStatusCode should be 200")
	assert.Equal(true, strings.Contains(fmt.Sprintf("%s", w.Body), "<title>page_title_search</title>"), "http respond should have '<title>page_title_searc</title>'")

	r, _ = http.NewRequest("POST", "/login", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(401), w.Code, "'/login' httpStatusCode should be 401")

	r, _ = http.NewRequest("GET", "/log_out", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(200), w.Code, "'/log_out' httpStatusCode should be 200")
	assert.Equal(true, strings.Contains(fmt.Sprintf("%s", w.Body), ""), "http respond should be empty")

	r, _ = http.NewRequest("POST", "/reset", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(400), w.Code, "'/reset' httpStatusCode should be 400")

	r, _ = http.NewRequest("POST", "/userExists", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(500), w.Code, "'/userExists' httpStatusCode should be 500")

	r, _ = http.NewRequest("GET", "/sendEmail", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(400), w.Code, "'/sendEmail' httpStatusCode should be 400")

	r, _ = http.NewRequest("GET", "/language", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(302), w.Code, "'/language' httpStatusCode should be 302")

	r, _ = http.NewRequest("GET", "/optional_menu", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	//fmt.Printf("/optional_menu: %s\n", w.Body)
	assert.Equal(int(200), w.Code, "'/optional_menu' httpStatusCode should be 200")
	//assert.Equal(true, strings.Contains(fmt.Sprintf("%s", w.Body), "<title> </title>"), "http respond should have '<title> </title>'")

	r, _ = http.NewRequest("GET", "/navigation_header", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	//fmt.Printf("/navigation_header: %s\n", w.Body)
	assert.Equal(int(200), w.Code, "'/navigation_header' httpStatusCode should be 200")
	//assert.Equal(true, strings.Contains(fmt.Sprintf("%s", w.Body), "<title> </title>"), "http respond should have '<title> </title>'")

	r, _ = http.NewRequest("GET", "/navigation_detail", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	//fmt.Printf("/navigation_detail: %s\n", w.Body)
	assert.Equal(int(200), w.Code, "'/navigation_detail' httpStatusCode should be 200")
	//assert.Equal(true, strings.Contains(fmt.Sprintf("%s", w.Body), "<title> </title>"), "http respond should have '<title> </title>'")

	r, _ = http.NewRequest("GET", "/sign_in", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	//fmt.Printf("/sign_in: %s\n", w.Body)
	assert.Equal(int(200), w.Code, "'/sign_in' httpStatusCode should be 200")
	//assert.Equal(true, strings.Contains(fmt.Sprintf("%s", w.Body), "<title> </title>"), "http respond should have '<title> </title>'")

}
