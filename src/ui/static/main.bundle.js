webpackJsonp([0,4],{

/***/ 10:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var Subject_1 = __webpack_require__(25);
var message_1 = __webpack_require__(389);
var MessageService = (function () {
    function MessageService() {
        this.messageAnnouncedSource = new Subject_1.Subject();
        this.appLevelAnnouncedSource = new Subject_1.Subject();
        this.messageAnnounced$ = this.messageAnnouncedSource.asObservable();
        this.appLevelAnnounced$ = this.appLevelAnnouncedSource.asObservable();
    }
    MessageService.prototype.announceMessage = function (statusCode, message, alertType) {
        this.messageAnnouncedSource.next(message_1.Message.newMessage(statusCode, message, alertType));
    };
    MessageService.prototype.announceAppLevelMessage = function (statusCode, message, alertType) {
        this.appLevelAnnouncedSource.next(message_1.Message.newMessage(statusCode, message, alertType));
    };
    MessageService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [])
    ], MessageService);
    return MessageService;
}());
exports.MessageService = MessageService;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/message.service.js.map

/***/ }),

/***/ 14:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var http_1 = __webpack_require__(20);
__webpack_require__(58);
var shared_const_1 = __webpack_require__(2);
var signInUrl = '/login';
var currentUserEndpint = "/api/users/current";
var signOffEndpoint = "/log_out";
var accountEndpoint = "/api/users/:id";
var langEndpoint = "/language";
var langMap = {
    "zh": "zh-CN",
    "en": "en-US"
};
/**
 * Define related methods to handle account and session corresponding things
 *
 * @export
 * @class SessionService
 */
var SessionService = (function () {
    function SessionService(http) {
        this.http = http;
        this.currentUser = null;
        this.headers = new http_1.Headers({
            "Content-Type": 'application/json'
        });
        this.formHeaders = new http_1.Headers({
            "Content-Type": 'application/x-www-form-urlencoded'
        });
    }
    //Handle the related exceptions
    SessionService.prototype.handleError = function (error) {
        return Promise.reject(error.message || error);
    };
    //Submit signin form to backend (NOT restful service)
    SessionService.prototype.signIn = function (signInCredential) {
        var _this = this;
        //Build the form package
        var body = new http_1.URLSearchParams();
        body.set('principal', signInCredential.principal);
        body.set('password', signInCredential.password);
        //Trigger Http
        return this.http.post(signInUrl, body.toString(), { headers: this.formHeaders })
            .toPromise()
            .then(function () { return null; })
            .catch(function (error) { return _this.handleError(error); });
    };
    /**
     * Get the related information of current signed in user from backend
     *
     * @returns {Promise<SessionUser>}
     *
     * @memberOf SessionService
     */
    SessionService.prototype.retrieveUser = function () {
        var _this = this;
        return this.http.get(currentUserEndpint, { headers: this.headers }).toPromise()
            .then(function (response) { return _this.currentUser = response.json(); })
            .catch(function (error) { return _this.handleError(error); });
    };
    /**
     * For getting info
     */
    SessionService.prototype.getCurrentUser = function () {
        return this.currentUser;
    };
    /**
     * Log out the system
     */
    SessionService.prototype.signOff = function () {
        var _this = this;
        return this.http.get(signOffEndpoint, { headers: this.headers }).toPromise()
            .then(function () {
            //Destroy current session cache
            _this.currentUser = null;
        }) //Nothing returned
            .catch(function (error) { return _this.handleError(error); });
    };
    /**
     *
     * Update accpunt settings
     *
     * @param {SessionUser} account
     * @returns {Promise<any>}
     *
     * @memberOf SessionService
     */
    SessionService.prototype.updateAccountSettings = function (account) {
        var _this = this;
        if (!account) {
            return Promise.reject("Invalid account settings");
        }
        var putUrl = accountEndpoint.replace(":id", account.user_id + "");
        return this.http.put(putUrl, JSON.stringify(account), { headers: this.headers }).toPromise()
            .then(function () {
            //Retrieve current session user
            return _this.retrieveUser();
        })
            .catch(function (error) { return _this.handleError(error); });
    };
    /**
     * Switch the backend language profile
     */
    SessionService.prototype.switchLanguage = function (lang) {
        var _this = this;
        if (!lang) {
            return Promise.reject("Invalid language");
        }
        var backendLang = langMap[lang];
        if (!backendLang) {
            backendLang = langMap[shared_const_1.enLang];
        }
        var getUrl = langEndpoint + "?lang=" + backendLang;
        return this.http.get(getUrl).toPromise()
            .then(function () { return null; })
            .catch(function (error) { return _this.handleError(error); });
    };
    SessionService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof http_1.Http !== 'undefined' && http_1.Http) === 'function' && _a) || Object])
    ], SessionService);
    return SessionService;
    var _a;
}());
exports.SessionService = SessionService;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/session.service.js.map

/***/ }),

/***/ 171:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var http_1 = __webpack_require__(20);
__webpack_require__(58);
var passwordChangeEndpoint = "/api/users/:user_id/password";
var sendEmailEndpoint = "/sendEmail";
var resetPasswordEndpoint = "/reset";
var PasswordSettingService = (function () {
    function PasswordSettingService(http) {
        this.http = http;
        this.headers = new http_1.Headers({
            "Accept": 'application/json',
            "Content-Type": 'application/json'
        });
        this.options = new http_1.RequestOptions({
            'headers': this.headers
        });
    }
    PasswordSettingService.prototype.changePassword = function (userId, setting) {
        if (!setting || setting.new_password.trim() === "" || setting.old_password.trim() === "") {
            return Promise.reject("Invalid data");
        }
        var putUrl = passwordChangeEndpoint.replace(":user_id", userId + "");
        return this.http.put(putUrl, JSON.stringify(setting), this.options)
            .toPromise()
            .then(function () { return null; })
            .catch(function (error) {
            return Promise.reject(error);
        });
    };
    PasswordSettingService.prototype.sendResetPasswordMail = function (email) {
        if (!email) {
            return Promise.reject("Invalid email");
        }
        var getUrl = sendEmailEndpoint + "?email=" + email;
        return this.http.get(getUrl, this.options).toPromise()
            .then(function (response) { return response; })
            .catch(function (error) {
            return Promise.reject(error);
        });
    };
    PasswordSettingService.prototype.resetPassword = function (uuid, newPassword) {
        if (!uuid || !newPassword) {
            return Promise.reject("Invalid reset uuid or password");
        }
        var formHeaders = new http_1.Headers({
            "Content-Type": 'application/x-www-form-urlencoded'
        });
        var formOptions = new http_1.RequestOptions({
            headers: formHeaders
        });
        var body = new http_1.URLSearchParams();
        body.set("reset_uuid", uuid);
        body.set("password", newPassword);
        return this.http.post(resetPasswordEndpoint, body.toString(), formOptions)
            .toPromise()
            .then(function (response) { return response; })
            .catch(function (error) {
            return Promise.reject(error);
        });
    };
    PasswordSettingService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof http_1.Http !== 'undefined' && http_1.Http) === 'function' && _a) || Object])
    ], PasswordSettingService);
    return PasswordSettingService;
    var _a;
}());
exports.PasswordSettingService = PasswordSettingService;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/password-setting.service.js.map

/***/ }),

/***/ 172:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var http_1 = __webpack_require__(20);
__webpack_require__(58);
var app_config_1 = __webpack_require__(244);
exports.systemInfoEndpoint = "/api/systeminfo";
/**
 * Declare service to handle the bootstrap options
 *
 *
 * @export
 * @class GlobalSearchService
 */
var AppConfigService = (function () {
    function AppConfigService(http) {
        this.http = http;
        this.headers = new http_1.Headers({
            "Content-Type": 'application/json'
        });
        this.options = new http_1.RequestOptions({
            headers: this.headers
        });
        //Store the application configuration
        this.configurations = new app_config_1.AppConfig();
    }
    AppConfigService.prototype.load = function () {
        var _this = this;
        return this.http.get(exports.systemInfoEndpoint, this.options).toPromise()
            .then(function (response) { return _this.configurations = response.json(); })
            .catch(function (error) {
            //Catch the error
            console.error("Failed to load bootstrap options with error: ", error);
        });
    };
    AppConfigService.prototype.getConfig = function () {
        return this.configurations;
    };
    AppConfigService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof http_1.Http !== 'undefined' && http_1.Http) === 'function' && _a) || Object])
    ], AppConfigService);
    return AppConfigService;
    var _a;
}());
exports.AppConfigService = AppConfigService;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/app-config.service.js.map

/***/ }),

/***/ 173:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var StringValueItem = (function () {
    function StringValueItem(v, e) {
        this.value = v;
        this.editable = e;
    }
    return StringValueItem;
}());
exports.StringValueItem = StringValueItem;
var NumberValueItem = (function () {
    function NumberValueItem(v, e) {
        this.value = v;
        this.editable = e;
    }
    return NumberValueItem;
}());
exports.NumberValueItem = NumberValueItem;
var BoolValueItem = (function () {
    function BoolValueItem(v, e) {
        this.value = v;
        this.editable = e;
    }
    return BoolValueItem;
}());
exports.BoolValueItem = BoolValueItem;
var Configuration = (function () {
    function Configuration() {
        this.auth_mode = new StringValueItem("db_auth", true);
        this.project_creation_restriction = new StringValueItem("everyone", true);
        this.self_registration = new BoolValueItem(false, true);
        this.ldap_base_dn = new StringValueItem("", true);
        this.ldap_filter = new StringValueItem("", true);
        this.ldap_scope = new NumberValueItem(0, true);
        this.ldap_search_dn = new StringValueItem("", true);
        this.ldap_search_password = new StringValueItem("", true);
        this.ldap_timeout = new NumberValueItem(5, true);
        this.ldap_uid = new StringValueItem("", true);
        this.ldap_url = new StringValueItem("", true);
        this.email_host = new StringValueItem("", true);
        this.email_identity = new StringValueItem("", true);
        this.email_from = new StringValueItem("", true);
        this.email_port = new NumberValueItem(25, true);
        this.email_ssl = new BoolValueItem(false, true);
        this.email_username = new StringValueItem("", true);
        this.email_password = new StringValueItem("", true);
        this.token_expiration = new NumberValueItem(5, true);
        this.cfg_expiration = new NumberValueItem(30, true);
        this.verify_remote_cert = new BoolValueItem(false, true);
    }
    return Configuration;
}());
exports.Configuration = Configuration;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/config.js.map

/***/ }),

/***/ 174:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var http_1 = __webpack_require__(20);
var Observable_1 = __webpack_require__(3);
__webpack_require__(133);
__webpack_require__(85);
__webpack_require__(132);
var ProjectService = (function () {
    function ProjectService(http) {
        this.http = http;
        this.headers = new http_1.Headers({ 'Content-type': 'application/json' });
        this.options = new http_1.RequestOptions({ 'headers': this.headers });
    }
    ProjectService.prototype.getProject = function (projectId) {
        return this.http
            .get("/api/projects/" + projectId)
            .toPromise()
            .then(function (response) { return response.json(); })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    ProjectService.prototype.listProjects = function (name, isPublic, page, pageSize) {
        var params = new http_1.URLSearchParams();
        params.set('page', page + '');
        params.set('page_size', pageSize + '');
        return this.http
            .get("/api/projects?project_name=" + name + "&is_public=" + isPublic, { search: params })
            .map(function (response) { return response; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    ProjectService.prototype.createProject = function (name, isPublic) {
        return this.http
            .post("/api/projects", JSON.stringify({ 'project_name': name, 'public': isPublic }), this.options)
            .map(function (response) { return response.status; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    ProjectService.prototype.toggleProjectPublic = function (projectId, isPublic) {
        return this.http
            .put("/api/projects/" + projectId + "/publicity", { 'public': isPublic }, this.options)
            .map(function (response) { return response.status; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    ProjectService.prototype.deleteProject = function (projectId) {
        return this.http
            .delete("/api/projects/" + projectId)
            .map(function (response) { return response.status; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    ProjectService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof http_1.Http !== 'undefined' && http_1.Http) === 'function' && _a) || Object])
    ], ProjectService);
    return ProjectService;
    var _a;
}());
exports.ProjectService = ProjectService;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/project.service.js.map

/***/ }),

/***/ 175:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var http_1 = __webpack_require__(20);
__webpack_require__(58);
var userMgmtEndpoint = '/api/users';
/**
 * Define related methods to handle account and session corresponding things
 *
 * @export
 * @class SessionService
 */
var UserService = (function () {
    function UserService(http) {
        this.http = http;
        this.httpOptions = new http_1.RequestOptions({
            headers: new http_1.Headers({
                "Content-Type": 'application/json'
            })
        });
    }
    //Handle the related exceptions
    UserService.prototype.handleError = function (error) {
        return Promise.reject(error.message || error);
    };
    //Get the user list
    UserService.prototype.getUsers = function () {
        var _this = this;
        return this.http.get(userMgmtEndpoint, this.httpOptions).toPromise()
            .then(function (response) { return response.json(); })
            .catch(function (error) { return _this.handleError(error); });
    };
    //Add new user
    UserService.prototype.addUser = function (user) {
        var _this = this;
        return this.http.post(userMgmtEndpoint, JSON.stringify(user), this.httpOptions).toPromise()
            .then(function () { return null; })
            .catch(function (error) { return _this.handleError(error); });
    };
    //Delete the specified user
    UserService.prototype.deleteUser = function (userId) {
        var _this = this;
        return this.http.delete(userMgmtEndpoint + "/" + userId, this.httpOptions)
            .toPromise()
            .then(function () { return null; })
            .catch(function (error) { return _this.handleError(error); });
    };
    //Update user to enable/disable the admin role
    UserService.prototype.updateUser = function (user) {
        var _this = this;
        return this.http.put(userMgmtEndpoint + "/" + user.user_id, JSON.stringify(user), this.httpOptions)
            .toPromise()
            .then(function () { return null; })
            .catch(function (error) { return _this.handleError(error); });
    };
    //Set user admin role
    UserService.prototype.updateUserRole = function (user) {
        var _this = this;
        return this.http.put(userMgmtEndpoint + "/" + user.user_id + "/sysadmin", JSON.stringify(user), this.httpOptions)
            .toPromise()
            .then(function () { return null; })
            .catch(function (error) { return _this.handleError(error); });
    };
    UserService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof http_1.Http !== 'undefined' && http_1.Http) === 'function' && _a) || Object])
    ], UserService);
    return UserService;
    var _a;
}());
exports.UserService = UserService;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/user.service.js.map

/***/ }),

/***/ 2:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

exports.supportedLangs = ['en', 'zh'];
exports.enLang = "en";
exports.languageNames = {
    "en": "English",
    "zh": "中文简体"
};
(function (AlertType) {
    AlertType[AlertType["DANGER"] = 0] = "DANGER";
    AlertType[AlertType["WARNING"] = 1] = "WARNING";
    AlertType[AlertType["INFO"] = 2] = "INFO";
    AlertType[AlertType["SUCCESS"] = 3] = "SUCCESS";
})(exports.AlertType || (exports.AlertType = {}));
var AlertType = exports.AlertType;
;
exports.dismissInterval = 15 * 1000;
exports.httpStatusCode = {
    "Unauthorized": 401,
    "Forbidden": 403
};
(function (DeletionTargets) {
    DeletionTargets[DeletionTargets["EMPTY"] = 0] = "EMPTY";
    DeletionTargets[DeletionTargets["PROJECT"] = 1] = "PROJECT";
    DeletionTargets[DeletionTargets["PROJECT_MEMBER"] = 2] = "PROJECT_MEMBER";
    DeletionTargets[DeletionTargets["USER"] = 3] = "USER";
    DeletionTargets[DeletionTargets["POLICY"] = 4] = "POLICY";
    DeletionTargets[DeletionTargets["TARGET"] = 5] = "TARGET";
    DeletionTargets[DeletionTargets["REPOSITORY"] = 6] = "REPOSITORY";
    DeletionTargets[DeletionTargets["TAG"] = 7] = "TAG";
})(exports.DeletionTargets || (exports.DeletionTargets = {}));
var DeletionTargets = exports.DeletionTargets;
;
exports.harborRootRoute = "/harbor/dashboard";
exports.signInRoute = "/sign-in";
(function (ActionType) {
    ActionType[ActionType["ADD_NEW"] = 0] = "ADD_NEW";
    ActionType[ActionType["EDIT"] = 1] = "EDIT";
})(exports.ActionType || (exports.ActionType = {}));
var ActionType = exports.ActionType;
;
exports.ListMode = {
    READONLY: "readonly",
    FULL: "full"
};
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/shared.const.js.map

/***/ }),

/***/ 243:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var new_user_form_component_1 = __webpack_require__(253);
var session_service_1 = __webpack_require__(14);
var user_service_1 = __webpack_require__(175);
var inline_alert_component_1 = __webpack_require__(80);
var clarity_angular_1 = __webpack_require__(419);
var SignUpComponent = (function () {
    function SignUpComponent(session, userService) {
        this.session = session;
        this.userService = userService;
        this.opened = false;
        this.staticBackdrop = true;
        this.onGoing = false;
        this.formValueChanged = false;
    }
    SignUpComponent.prototype.getNewUser = function () {
        return this.newUserForm.getData();
    };
    Object.defineProperty(SignUpComponent.prototype, "inProgress", {
        get: function () {
            return this.onGoing;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(SignUpComponent.prototype, "isValid", {
        get: function () {
            return this.newUserForm.isValid && this.error == null;
        },
        enumerable: true,
        configurable: true
    });
    SignUpComponent.prototype.formValueChange = function (flag) {
        if (flag) {
            this.formValueChanged = true;
        }
        if (this.error != null) {
            this.error = null; //clear error
        }
        this.inlienAlert.close(); //Close alert if being shown
    };
    SignUpComponent.prototype.open = function () {
        this.newUserForm.reset(); //Reset form
        this.formValueChanged = false;
        this.modal.open();
    };
    SignUpComponent.prototype.close = function () {
        if (this.formValueChanged) {
            if (this.newUserForm.isEmpty()) {
                this.opened = false;
            }
            else {
                //Need user confirmation
                this.inlienAlert.showInlineConfirmation({
                    message: "ALERT.FORM_CHANGE_CONFIRMATION"
                });
            }
        }
        else {
            this.opened = false;
        }
    };
    SignUpComponent.prototype.confirmCancel = function () {
        this.modal.close();
    };
    //Create new user
    SignUpComponent.prototype.create = function () {
        var _this = this;
        //Double confirm everything is ok
        //Form is valid
        if (!this.isValid) {
            return;
        }
        //We have new user data
        var u = this.getNewUser();
        if (!u) {
            return;
        }
        //Start process
        this.onGoing = true;
        this.userService.addUser(u)
            .then(function () {
            _this.onGoing = false;
            _this.modal.close();
        })
            .catch(function (error) {
            _this.onGoing = false;
            _this.error = error;
            _this.inlienAlert.showInlineError(error);
        });
    };
    __decorate([
        core_1.ViewChild(new_user_form_component_1.NewUserFormComponent), 
        __metadata('design:type', (typeof (_a = typeof new_user_form_component_1.NewUserFormComponent !== 'undefined' && new_user_form_component_1.NewUserFormComponent) === 'function' && _a) || Object)
    ], SignUpComponent.prototype, "newUserForm", void 0);
    __decorate([
        core_1.ViewChild(inline_alert_component_1.InlineAlertComponent), 
        __metadata('design:type', (typeof (_b = typeof inline_alert_component_1.InlineAlertComponent !== 'undefined' && inline_alert_component_1.InlineAlertComponent) === 'function' && _b) || Object)
    ], SignUpComponent.prototype, "inlienAlert", void 0);
    __decorate([
        core_1.ViewChild(clarity_angular_1.Modal), 
        __metadata('design:type', (typeof (_c = typeof clarity_angular_1.Modal !== 'undefined' && clarity_angular_1.Modal) === 'function' && _c) || Object)
    ], SignUpComponent.prototype, "modal", void 0);
    SignUpComponent = __decorate([
        core_1.Component({
            selector: 'sign-up',
            template: __webpack_require__(825)
        }), 
        __metadata('design:paramtypes', [(typeof (_d = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _d) || Object, (typeof (_e = typeof user_service_1.UserService !== 'undefined' && user_service_1.UserService) === 'function' && _e) || Object])
    ], SignUpComponent);
    return SignUpComponent;
    var _a, _b, _c, _d, _e;
}());
exports.SignUpComponent = SignUpComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/sign-up.component.js.map

/***/ }),

/***/ 244:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var AppConfig = (function () {
    function AppConfig() {
        //Set default value
        this.with_notary = false;
        this.with_admiral = false;
        this.admiral_endpoint = "";
        this.auth_mode = "db_auth";
        this.registry_url = "";
        this.project_creation_restriction = "everyone";
        this.self_registration = true;
    }
    return AppConfig;
}());
exports.AppConfig = AppConfig;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/app-config.js.map

/***/ }),

/***/ 245:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var session_service_1 = __webpack_require__(14);
var StartPageComponent = (function () {
    function StartPageComponent(session) {
        this.session = session;
        this.isSessionValid = false;
    }
    StartPageComponent.prototype.ngOnInit = function () {
        this.isSessionValid = this.session.getCurrentUser() != null;
    };
    StartPageComponent = __decorate([
        core_1.Component({
            selector: 'start-page',
            template: __webpack_require__(832),
            styles: [__webpack_require__(806)]
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _a) || Object])
    ], StartPageComponent);
    return StartPageComponent;
    var _a;
}());
exports.StartPageComponent = StartPageComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/start.component.js.map

/***/ }),

/***/ 246:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var platform_browser_1 = __webpack_require__(121);
var core_1 = __webpack_require__(0);
var forms_1 = __webpack_require__(26);
var http_1 = __webpack_require__(20);
var clarity_angular_1 = __webpack_require__(419);
var CoreModule = (function () {
    function CoreModule() {
    }
    CoreModule = __decorate([
        core_1.NgModule({
            imports: [
                platform_browser_1.BrowserModule,
                forms_1.FormsModule,
                http_1.HttpModule,
                clarity_angular_1.ClarityModule.forRoot()
            ],
            exports: [
                platform_browser_1.BrowserModule,
                forms_1.FormsModule,
                http_1.HttpModule,
                clarity_angular_1.ClarityModule
            ]
        }), 
        __metadata('design:paramtypes', [])
    ], CoreModule);
    return CoreModule;
}());
exports.CoreModule = CoreModule;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/core.module.js.map

/***/ }),

/***/ 247:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var http_1 = __webpack_require__(20);
var base_service_1 = __webpack_require__(251);
__webpack_require__(133);
__webpack_require__(85);
__webpack_require__(132);
exports.logEndpoint = "/api/logs";
var AuditLogService = (function (_super) {
    __extends(AuditLogService, _super);
    function AuditLogService(http) {
        _super.call(this);
        this.http = http;
        this.httpOptions = new http_1.RequestOptions({
            headers: new http_1.Headers({
                "Content-Type": 'application/json',
                "Accept": 'application/json'
            })
        });
    }
    AuditLogService.prototype.listAuditLogs = function (queryParam) {
        var _this = this;
        return this.http
            .post("/api/projects/" + queryParam.project_id + "/logs/filter?page=" + queryParam.page + "&page_size=" + queryParam.page_size, {
            begin_timestamp: queryParam.begin_timestamp,
            end_timestamp: queryParam.end_timestamp,
            keywords: queryParam.keywords,
            operation: queryParam.operation,
            project_id: queryParam.project_id,
            username: queryParam.username
        })
            .map(function (response) { return response; })
            .catch(function (error) { return _this.handleError(error); });
    };
    AuditLogService.prototype.getRecentLogs = function (lines) {
        var _this = this;
        return this.http.get(exports.logEndpoint + "?lines=" + lines, this.httpOptions)
            .map(function (response) { return response.json(); })
            .catch(function (error) { return _this.handleError(error); });
    };
    AuditLogService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof http_1.Http !== 'undefined' && http_1.Http) === 'function' && _a) || Object])
    ], AuditLogService);
    return AuditLogService;
    var _a;
}(base_service_1.BaseService));
exports.AuditLogService = AuditLogService;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/audit-log.service.js.map

/***/ }),

/***/ 248:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var http_1 = __webpack_require__(20);
var Observable_1 = __webpack_require__(3);
__webpack_require__(133);
__webpack_require__(85);
__webpack_require__(132);
var base_service_1 = __webpack_require__(251);
var MemberService = (function (_super) {
    __extends(MemberService, _super);
    function MemberService(http) {
        _super.call(this);
        this.http = http;
    }
    MemberService.prototype.listMembers = function (projectId, username) {
        var _this = this;
        console.log('Get member from project_id:' + projectId + ', username:' + username);
        return this.http
            .get("/api/projects/" + projectId + "/members?username=" + username)
            .map(function (response) { return response.json(); })
            .catch(function (error) { return _this.handleError(error); });
    };
    MemberService.prototype.addMember = function (projectId, username, roleId) {
        console.log('Adding member with username:' + username + ', roleId:' + roleId + ' under projectId:' + projectId);
        return this.http
            .post("/api/projects/" + projectId + "/members", { username: username, roles: [roleId] })
            .map(function (response) { return response.status; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    MemberService.prototype.changeMemberRole = function (projectId, userId, roleId) {
        console.log('Changing member role with userId:' + ' to roleId:' + roleId + ' under projectId:' + projectId);
        return this.http
            .put("/api/projects/" + projectId + "/members/" + userId, { roles: [roleId] })
            .map(function (response) { return response.status; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    MemberService.prototype.deleteMember = function (projectId, userId) {
        console.log('Deleting member role with userId:' + userId + ' under projectId:' + projectId);
        return this.http
            .delete("/api/projects/" + projectId + "/members/" + userId)
            .map(function (response) { return response.status; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    MemberService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof http_1.Http !== 'undefined' && http_1.Http) === 'function' && _a) || Object])
    ], MemberService);
    return MemberService;
    var _a;
}(base_service_1.BaseService));
exports.MemberService = MemberService;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/member.service.js.map

/***/ }),

/***/ 249:
/***/ (function(module, exports, __webpack_require__) {

"use strict";
/*
 {
    "id": 1,
    "endpoint": "http://10.117.4.151",
    "name": "target_01",
    "username": "admin",
    "password": "Harbor12345",
    "type": 0,
    "creation_time": "2017-02-24T06:41:52Z",
    "update_time": "2017-02-24T06:41:52Z"
  }
*/

var Target = (function () {
    function Target() {
    }
    return Target;
}());
exports.Target = Target;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/target.js.map

/***/ }),

/***/ 250:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var http_1 = __webpack_require__(20);
var Observable_1 = __webpack_require__(3);
__webpack_require__(460);
__webpack_require__(463);
var RepositoryService = (function () {
    function RepositoryService(http) {
        this.http = http;
    }
    RepositoryService.prototype.listRepositories = function (projectId, repoName, page, pageSize) {
        console.log('List repositories with project ID:' + projectId);
        var params = new http_1.URLSearchParams();
        params.set('page', page + '');
        params.set('page_size', pageSize + '');
        return this.http
            .get("/api/repositories?project_id=" + projectId + "&q=" + repoName + "&detail=1", { search: params })
            .map(function (response) { return response; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    RepositoryService.prototype.listTags = function (repoName) {
        return this.http
            .get("/api/repositories/tags?repo_name=" + repoName + "&detail=1")
            .map(function (response) { return response.json(); })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    RepositoryService.prototype.listNotarySignatures = function (repoName) {
        return this.http
            .get("/api/repositories/signatures?repo_name=" + repoName)
            .map(function (response) { return response.json(); })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    RepositoryService.prototype.listTagsWithVerifiedSignatures = function (repoName) {
        var _this = this;
        return this.http
            .get("/api/repositories/signatures?repo_name=" + repoName)
            .map(function (response) { return response; })
            .flatMap(function (res) {
            return _this.listTags(repoName)
                .map(function (tags) {
                var signatures = res.json();
                tags.forEach(function (t) {
                    for (var i = 0; i < signatures.length; i++) {
                        if (signatures[i].tag === t.tag) {
                            t.verified = true;
                            break;
                        }
                    }
                });
                return tags;
            })
                .catch(function (error) { return Observable_1.Observable.throw(error); });
        })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    RepositoryService.prototype.deleteRepository = function (repoName) {
        console.log('Delete repository with repo name:' + repoName);
        return this.http
            .delete("/api/repositories?repo_name=" + repoName)
            .map(function (response) { return response.status; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    RepositoryService.prototype.deleteRepoByTag = function (repoName, tag) {
        console.log('Delete repository with repo name:' + repoName + ', tag:' + tag);
        return this.http
            .delete("/api/repositories?repo_name=" + repoName + "&tag=" + tag)
            .map(function (response) { return response.status; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    RepositoryService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof http_1.Http !== 'undefined' && http_1.Http) === 'function' && _a) || Object])
    ], RepositoryService);
    return RepositoryService;
    var _a;
}());
exports.RepositoryService = RepositoryService;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/repository.service.js.map

/***/ }),

/***/ 251:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var http_1 = __webpack_require__(20);
var BaseService = (function () {
    function BaseService() {
    }
    BaseService.prototype.handleError = function (error) {
        // In a real world app, we might use a remote logging infrastructure
        var errMsg;
        console.log(typeof error);
        if (error instanceof http_1.Response) {
            var body = error.json() || '';
            var err = body.error || JSON.stringify(body);
            errMsg = error.status + " - " + (error.statusText || '') + " " + err;
        }
        else {
            errMsg = error.message ? error.message : error.toString();
        }
        return Promise.reject(error);
    };
    return BaseService;
}());
exports.BaseService = BaseService;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/base.service.js.map

/***/ }),

/***/ 252:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var create_edit_policy_1 = __webpack_require__(624);
var replication_service_1 = __webpack_require__(79);
var message_service_1 = __webpack_require__(10);
var shared_const_1 = __webpack_require__(2);
var policy_1 = __webpack_require__(617);
var target_1 = __webpack_require__(249);
var core_2 = __webpack_require__(34);
var CreateEditPolicyComponent = (function () {
    function CreateEditPolicyComponent(replicationService, messageService, translateService) {
        this.replicationService = replicationService;
        this.messageService = messageService;
        this.translateService = translateService;
        this.createEditPolicy = new create_edit_policy_1.CreateEditPolicy();
        this.reload = new core_1.EventEmitter();
    }
    CreateEditPolicyComponent.prototype.prepareTargets = function (targetId) {
        var _this = this;
        this.replicationService
            .listTargets('')
            .subscribe(function (targets) {
            _this.targets = targets;
            if (_this.targets && _this.targets.length > 0) {
                var initialTarget = void 0;
                (targetId) ? initialTarget = _this.targets.find(function (t) { return t.id == targetId; }) : initialTarget = _this.targets[0];
                _this.createEditPolicy.targetId = initialTarget.id;
                _this.createEditPolicy.targetName = initialTarget.name;
                _this.createEditPolicy.endpointUrl = initialTarget.endpoint;
                _this.createEditPolicy.username = initialTarget.username;
                _this.createEditPolicy.password = initialTarget.password;
            }
        }, function (error) { return _this.messageService.announceMessage(error.status, 'Error occurred while get targets.', shared_const_1.AlertType.DANGER); });
    };
    CreateEditPolicyComponent.prototype.ngOnInit = function () { };
    CreateEditPolicyComponent.prototype.openCreateEditPolicy = function (policyId) {
        var _this = this;
        this.createEditPolicyOpened = true;
        this.createEditPolicy = new create_edit_policy_1.CreateEditPolicy();
        this.isCreateDestination = false;
        this.errorMessageOpened = false;
        this.errorMessage = '';
        this.pingTestMessage = '';
        this.pingStatus = true;
        this.testOngoing = false;
        if (policyId) {
            this.actionType = shared_const_1.ActionType.EDIT;
            this.translateService.get('REPLICATION.EDIT_POLICY').subscribe(function (res) { return _this.modalTitle = res; });
            this.replicationService
                .getPolicy(policyId)
                .subscribe(function (policy) {
                _this.createEditPolicy.policyId = policyId;
                _this.createEditPolicy.name = policy.name;
                _this.createEditPolicy.description = policy.description;
                _this.createEditPolicy.enable = policy.enabled === 1 ? true : false;
                _this.prepareTargets(policy.target_id);
            });
        }
        else {
            this.actionType = shared_const_1.ActionType.ADD_NEW;
            this.translateService.get('REPLICATION.ADD_POLICY').subscribe(function (res) { return _this.modalTitle = res; });
            this.prepareTargets();
        }
    };
    CreateEditPolicyComponent.prototype.newDestination = function (checkedAddNew) {
        console.log('CheckedAddNew:' + checkedAddNew);
        this.isCreateDestination = checkedAddNew;
        if (this.isCreateDestination) {
            this.createEditPolicy.targetName = '';
            this.createEditPolicy.endpointUrl = '';
            this.createEditPolicy.username = '';
            this.createEditPolicy.password = '';
        }
        else {
            this.prepareTargets();
        }
    };
    CreateEditPolicyComponent.prototype.selectTarget = function () {
        var _this = this;
        var result = this.targets.find(function (target) { return target.id == _this.createEditPolicy.targetId; });
        if (result) {
            this.createEditPolicy.targetId = result.id;
            this.createEditPolicy.endpointUrl = result.endpoint;
            this.createEditPolicy.username = result.username;
            this.createEditPolicy.password = result.password;
        }
    };
    CreateEditPolicyComponent.prototype.onErrorMessageClose = function () {
        this.errorMessageOpened = false;
        this.errorMessage = '';
    };
    CreateEditPolicyComponent.prototype.getPolicyByForm = function () {
        var policy = new policy_1.Policy();
        policy.project_id = this.projectId;
        policy.id = this.createEditPolicy.policyId;
        policy.name = this.createEditPolicy.name;
        policy.description = this.createEditPolicy.description;
        policy.enabled = this.createEditPolicy.enable ? 1 : 0;
        policy.target_id = this.createEditPolicy.targetId;
        return policy;
    };
    CreateEditPolicyComponent.prototype.getTargetByForm = function () {
        var target = new target_1.Target();
        target.id = this.createEditPolicy.targetId;
        target.name = this.createEditPolicy.targetName;
        target.endpoint = this.createEditPolicy.endpointUrl;
        target.username = this.createEditPolicy.username;
        target.password = this.createEditPolicy.password;
        return target;
    };
    CreateEditPolicyComponent.prototype.createPolicy = function () {
        var _this = this;
        console.log('Create policy with existing target in component.');
        this.replicationService
            .createPolicy(this.getPolicyByForm())
            .subscribe(function (response) {
            console.log('Successful created policy: ' + response);
            _this.createEditPolicyOpened = false;
            _this.reload.emit(true);
        }, function (error) {
            _this.errorMessageOpened = true;
            _this.errorMessage = error['_body'];
            console.log('Failed to create policy:' + error.status + ', error message:' + JSON.stringify(error['_body']));
        });
    };
    CreateEditPolicyComponent.prototype.createOrUpdatePolicyAndCreateTarget = function () {
        var _this = this;
        console.log('Creating policy with new created target.');
        this.replicationService
            .createOrUpdatePolicyWithNewTarget(this.getPolicyByForm(), this.getTargetByForm())
            .subscribe(function (response) {
            console.log('Successful created policy and target:' + response);
            _this.createEditPolicyOpened = false;
            _this.reload.emit(true);
        }, function (error) {
            _this.errorMessageOpened = true;
            _this.errorMessage = error['_body'];
            console.log('Failed to create policy and target:' + error.status + ', error message:' + JSON.stringify(error['_body']));
        });
    };
    CreateEditPolicyComponent.prototype.updatePolicy = function () {
        var _this = this;
        console.log('Creating policy with existing target.');
        this.replicationService
            .updatePolicy(this.getPolicyByForm())
            .subscribe(function (response) {
            console.log('Successful created policy and target:' + response);
            _this.createEditPolicyOpened = false;
            _this.reload.emit(true);
        }, function (error) {
            _this.errorMessageOpened = true;
            _this.errorMessage = error['_body'];
            console.log('Failed to create policy and target:' + error.status + ', error message:' + JSON.stringify(error['_body']));
        });
    };
    CreateEditPolicyComponent.prototype.onSubmit = function () {
        if (this.isCreateDestination) {
            this.createOrUpdatePolicyAndCreateTarget();
        }
        else {
            if (this.actionType === shared_const_1.ActionType.ADD_NEW) {
                this.createPolicy();
            }
            else if (this.actionType === shared_const_1.ActionType.EDIT) {
                this.updatePolicy();
            }
        }
        this.errorMessageOpened = false;
        this.errorMessage = '';
    };
    CreateEditPolicyComponent.prototype.testConnection = function () {
        var _this = this;
        this.pingStatus = true;
        this.translateService.get('REPLICATION.TESTING_CONNECTION').subscribe(function (res) { return _this.pingTestMessage = res; });
        this.testOngoing = !this.testOngoing;
        var pingTarget = new target_1.Target();
        pingTarget.endpoint = this.createEditPolicy.endpointUrl;
        pingTarget.username = this.createEditPolicy.username;
        pingTarget.password = this.createEditPolicy.password;
        this.replicationService
            .pingTarget(pingTarget)
            .subscribe(function (response) {
            _this.testOngoing = !_this.testOngoing;
            _this.translateService.get('REPLICATION.TEST_CONNECTION_SUCCESS').subscribe(function (res) { return _this.pingTestMessage = res; });
            _this.pingStatus = true;
        }, function (error) {
            _this.testOngoing = !_this.testOngoing;
            _this.translateService.get('REPLICATION.TEST_CONNECTION_FAILURE').subscribe(function (res) { return _this.pingTestMessage = res; });
            _this.pingStatus = false;
        });
    };
    __decorate([
        core_1.Input(), 
        __metadata('design:type', Number)
    ], CreateEditPolicyComponent.prototype, "projectId", void 0);
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], CreateEditPolicyComponent.prototype, "reload", void 0);
    CreateEditPolicyComponent = __decorate([
        core_1.Component({
            selector: 'create-edit-policy',
            template: __webpack_require__(856)
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof replication_service_1.ReplicationService !== 'undefined' && replication_service_1.ReplicationService) === 'function' && _a) || Object, (typeof (_b = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _b) || Object, (typeof (_c = typeof core_2.TranslateService !== 'undefined' && core_2.TranslateService) === 'function' && _c) || Object])
    ], CreateEditPolicyComponent);
    return CreateEditPolicyComponent;
    var _a, _b, _c;
}());
exports.CreateEditPolicyComponent = CreateEditPolicyComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/create-edit-policy.component.js.map

/***/ }),

/***/ 253:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var forms_1 = __webpack_require__(26);
var user_1 = __webpack_require__(638);
var shared_utils_1 = __webpack_require__(33);
var NewUserFormComponent = (function () {
    function NewUserFormComponent() {
        this.newUser = new user_1.User();
        this.confirmedPwd = "";
        this.isSelfRegistration = false;
        //Notify the form value changes
        this.valueChange = new core_1.EventEmitter();
    }
    Object.defineProperty(NewUserFormComponent.prototype, "isValid", {
        get: function () {
            var pwdEqualStatus = true;
            if (this.newUserForm.controls["confirmPassword"] &&
                this.newUserForm.controls["newPassword"]) {
                pwdEqualStatus = this.newUserForm.controls["confirmPassword"].value === this.newUserForm.controls["newPassword"].value;
            }
            return this.newUserForm &&
                this.newUserForm.valid && pwdEqualStatus;
        },
        enumerable: true,
        configurable: true
    });
    NewUserFormComponent.prototype.ngAfterViewChecked = function () {
        var _this = this;
        if (this.newUserFormRef != this.newUserForm) {
            this.newUserFormRef = this.newUserForm;
            if (this.newUserFormRef) {
                this.newUserFormRef.valueChanges.subscribe(function (data) {
                    _this.valueChange.emit(true);
                });
            }
        }
    };
    //Return the current user data
    NewUserFormComponent.prototype.getData = function () {
        return this.newUser;
    };
    //Reset form
    NewUserFormComponent.prototype.reset = function () {
        if (this.newUserForm) {
            this.newUserForm.reset();
        }
    };
    //To check if form is empty
    NewUserFormComponent.prototype.isEmpty = function () {
        return shared_utils_1.isEmptyForm(this.newUserForm);
    };
    __decorate([
        core_1.Input(), 
        __metadata('design:type', Boolean)
    ], NewUserFormComponent.prototype, "isSelfRegistration", void 0);
    __decorate([
        core_1.ViewChild("newUserFrom"), 
        __metadata('design:type', (typeof (_a = typeof forms_1.NgForm !== 'undefined' && forms_1.NgForm) === 'function' && _a) || Object)
    ], NewUserFormComponent.prototype, "newUserForm", void 0);
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], NewUserFormComponent.prototype, "valueChange", void 0);
    NewUserFormComponent = __decorate([
        core_1.Component({
            selector: 'new-user-form',
            template: __webpack_require__(862),
            styles: [__webpack_require__(816)]
        }), 
        __metadata('design:paramtypes', [])
    ], NewUserFormComponent);
    return NewUserFormComponent;
    var _a;
}());
exports.NewUserFormComponent = NewUserFormComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/new-user-form.component.js.map

/***/ }),

/***/ 277:
/***/ (function(module, exports) {

module.exports = ""

/***/ }),

/***/ 33:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var shared_const_1 = __webpack_require__(2);
/**
 * To handle the error message body
 *
 * @export
 * @returns {string}
 */
exports.errorHandler = function (error) {
    if (error) {
        if (error.message) {
            return error.message;
        }
        else if (error._body) {
            return error._body;
        }
        else if (error.statusText) {
            return error.statusText;
        }
        else {
            return error;
        }
    }
    return "UNKNOWN_ERROR";
};
/**
 * To check if form is empty
 */
exports.isEmptyForm = function (ngForm) {
    if (ngForm && ngForm.form) {
        var values = ngForm.form.value;
        if (values) {
            for (var key in values) {
                if (values[key]) {
                    return false;
                }
            }
        }
    }
    return true;
};
/**
 * Hanlde the 401 and 403 code
 *
 * If handled the 401 or 403, then return true otherwise false
 */
exports.accessErrorHandler = function (error, msgService) {
    if (error && error.status && msgService) {
        if (error.status === shared_const_1.httpStatusCode.Unauthorized) {
            msgService.announceAppLevelMessage(error.status, "UNAUTHORIZED_ERROR", shared_const_1.AlertType.DANGER);
            return true;
        }
        else if (error.status === shared_const_1.httpStatusCode.Forbidden) {
            msgService.announceAppLevelMessage(error.status, "FORBIDDEN_ERROR", shared_const_1.AlertType.DANGER);
            return true;
        }
    }
    return false;
};
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/shared.utils.js.map

/***/ }),

/***/ 374:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var forms_1 = __webpack_require__(26);
var session_service_1 = __webpack_require__(14);
var message_service_1 = __webpack_require__(10);
var shared_const_1 = __webpack_require__(2);
var shared_utils_1 = __webpack_require__(33);
var inline_alert_component_1 = __webpack_require__(80);
var AccountSettingsModalComponent = (function () {
    function AccountSettingsModalComponent(session, msgService) {
        this.session = session;
        this.msgService = msgService;
        this.opened = false;
        this.staticBackdrop = true;
        this.error = null;
        this.isOnCalling = false;
        this.formValueChanged = false;
    }
    AccountSettingsModalComponent.prototype.ngOnInit = function () {
        //Value copy
        this.account = Object.assign({}, this.session.getCurrentUser());
    };
    AccountSettingsModalComponent.prototype.isUserDataChange = function () {
        if (!this.originalStaticData || !this.account) {
            return false;
        }
        for (var prop in this.originalStaticData) {
            if (this.originalStaticData[prop]) {
                if (this.account[prop]) {
                    if (this.originalStaticData[prop] != this.account[prop]) {
                        return true;
                    }
                }
            }
        }
        return false;
    };
    Object.defineProperty(AccountSettingsModalComponent.prototype, "isValid", {
        get: function () {
            return this.accountForm && this.accountForm.valid && this.error === null;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(AccountSettingsModalComponent.prototype, "showProgress", {
        get: function () {
            return this.isOnCalling;
        },
        enumerable: true,
        configurable: true
    });
    AccountSettingsModalComponent.prototype.ngAfterViewChecked = function () {
        var _this = this;
        if (this.accountFormRef != this.accountForm) {
            this.accountFormRef = this.accountForm;
            if (this.accountFormRef) {
                this.accountFormRef.valueChanges.subscribe(function (data) {
                    if (_this.error) {
                        _this.error = null;
                    }
                    _this.formValueChanged = true;
                    _this.inlineAlert.close();
                });
            }
        }
    };
    AccountSettingsModalComponent.prototype.open = function () {
        //Keep the initial data for future diff
        this.originalStaticData = Object.assign({}, this.session.getCurrentUser());
        this.account = Object.assign({}, this.session.getCurrentUser());
        this.formValueChanged = false;
        this.opened = true;
    };
    AccountSettingsModalComponent.prototype.close = function () {
        if (this.formValueChanged) {
            if (!this.isUserDataChange()) {
                this.opened = false;
            }
            else {
                //Need user confirmation
                this.inlineAlert.showInlineConfirmation({
                    message: "ALERT.FORM_CHANGE_CONFIRMATION"
                });
            }
        }
        else {
            this.opened = false;
        }
    };
    AccountSettingsModalComponent.prototype.submit = function () {
        var _this = this;
        if (!this.isValid || this.isOnCalling) {
            return;
        }
        //Double confirm session is valid
        var cUser = this.session.getCurrentUser();
        if (!cUser) {
            return;
        }
        this.isOnCalling = true;
        this.session.updateAccountSettings(this.account)
            .then(function () {
            _this.isOnCalling = false;
            _this.opened = false;
            _this.msgService.announceMessage(200, "PROFILE.SAVE_SUCCESS", shared_const_1.AlertType.SUCCESS);
        })
            .catch(function (error) {
            _this.isOnCalling = false;
            _this.error = error;
            if (shared_utils_1.accessErrorHandler(error, _this.msgService)) {
                _this.opened = false;
            }
            else {
                _this.inlineAlert.showInlineError(error);
            }
        });
    };
    AccountSettingsModalComponent.prototype.confirmCancel = function () {
        this.inlineAlert.close();
        this.opened = false;
    };
    __decorate([
        core_1.ViewChild("accountSettingsFrom"), 
        __metadata('design:type', (typeof (_a = typeof forms_1.NgForm !== 'undefined' && forms_1.NgForm) === 'function' && _a) || Object)
    ], AccountSettingsModalComponent.prototype, "accountForm", void 0);
    __decorate([
        core_1.ViewChild(inline_alert_component_1.InlineAlertComponent), 
        __metadata('design:type', (typeof (_b = typeof inline_alert_component_1.InlineAlertComponent !== 'undefined' && inline_alert_component_1.InlineAlertComponent) === 'function' && _b) || Object)
    ], AccountSettingsModalComponent.prototype, "inlineAlert", void 0);
    AccountSettingsModalComponent = __decorate([
        core_1.Component({
            selector: "account-settings-modal",
            template: __webpack_require__(820)
        }), 
        __metadata('design:paramtypes', [(typeof (_c = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _c) || Object, (typeof (_d = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _d) || Object])
    ], AccountSettingsModalComponent);
    return AccountSettingsModalComponent;
    var _a, _b, _c, _d;
}());
exports.AccountSettingsModalComponent = AccountSettingsModalComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/account-settings-modal.component.js.map

/***/ }),

/***/ 375:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var core_module_1 = __webpack_require__(246);
var sign_in_component_1 = __webpack_require__(379);
var password_setting_component_1 = __webpack_require__(377);
var account_settings_modal_component_1 = __webpack_require__(374);
var shared_module_1 = __webpack_require__(52);
var sign_up_component_1 = __webpack_require__(243);
var forgot_password_component_1 = __webpack_require__(376);
var reset_password_component_1 = __webpack_require__(378);
var password_setting_service_1 = __webpack_require__(171);
var AccountModule = (function () {
    function AccountModule() {
    }
    AccountModule = __decorate([
        core_1.NgModule({
            imports: [
                core_module_1.CoreModule,
                router_1.RouterModule,
                shared_module_1.SharedModule
            ],
            declarations: [
                sign_in_component_1.SignInComponent,
                password_setting_component_1.PasswordSettingComponent,
                account_settings_modal_component_1.AccountSettingsModalComponent,
                sign_up_component_1.SignUpComponent,
                forgot_password_component_1.ForgotPasswordComponent,
                reset_password_component_1.ResetPasswordComponent],
            exports: [
                sign_in_component_1.SignInComponent,
                password_setting_component_1.PasswordSettingComponent,
                account_settings_modal_component_1.AccountSettingsModalComponent,
                reset_password_component_1.ResetPasswordComponent],
            providers: [password_setting_service_1.PasswordSettingService]
        }), 
        __metadata('design:paramtypes', [])
    ], AccountModule);
    return AccountModule;
}());
exports.AccountModule = AccountModule;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/account.module.js.map

/***/ }),

/***/ 376:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var forms_1 = __webpack_require__(26);
var password_setting_service_1 = __webpack_require__(171);
var inline_alert_component_1 = __webpack_require__(80);
var ForgotPasswordComponent = (function () {
    function ForgotPasswordComponent(pwdService) {
        this.pwdService = pwdService;
        this.opened = false;
        this.onGoing = false;
        this.email = "";
        this.validationState = true;
        this.forceValid = true;
    }
    Object.defineProperty(ForgotPasswordComponent.prototype, "showProgress", {
        get: function () {
            return this.onGoing;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(ForgotPasswordComponent.prototype, "isValid", {
        get: function () {
            return this.forgotPwdForm && this.forgotPwdForm.valid && this.forceValid;
        },
        enumerable: true,
        configurable: true
    });
    ForgotPasswordComponent.prototype.open = function () {
        this.opened = true;
        this.validationState = true;
        this.forceValid = true;
        this.forgotPwdForm.resetForm();
    };
    ForgotPasswordComponent.prototype.close = function () {
        this.opened = false;
    };
    ForgotPasswordComponent.prototype.send = function () {
        var _this = this;
        //Double confirm to avoid improper situations
        if (!this.email) {
            return;
        }
        if (!this.isValid) {
            return;
        }
        this.onGoing = true;
        this.pwdService.sendResetPasswordMail(this.email)
            .then(function (response) {
            _this.onGoing = false;
            _this.forceValid = false; //diable the send button
            _this.inlineAlert.showInlineSuccess({
                message: "RESET_PWD.SUCCESS"
            });
        })
            .catch(function (error) {
            _this.onGoing = false;
            _this.inlineAlert.showInlineError(error);
        });
    };
    ForgotPasswordComponent.prototype.handleValidation = function (flag) {
        if (flag) {
            this.validationState = true;
        }
        else {
            this.validationState = this.isValid;
        }
    };
    __decorate([
        core_1.ViewChild("forgotPasswordFrom"), 
        __metadata('design:type', (typeof (_a = typeof forms_1.NgForm !== 'undefined' && forms_1.NgForm) === 'function' && _a) || Object)
    ], ForgotPasswordComponent.prototype, "forgotPwdForm", void 0);
    __decorate([
        core_1.ViewChild(inline_alert_component_1.InlineAlertComponent), 
        __metadata('design:type', (typeof (_b = typeof inline_alert_component_1.InlineAlertComponent !== 'undefined' && inline_alert_component_1.InlineAlertComponent) === 'function' && _b) || Object)
    ], ForgotPasswordComponent.prototype, "inlineAlert", void 0);
    ForgotPasswordComponent = __decorate([
        core_1.Component({
            selector: 'forgot-password',
            template: __webpack_require__(821),
            styles: [__webpack_require__(457)]
        }), 
        __metadata('design:paramtypes', [(typeof (_c = typeof password_setting_service_1.PasswordSettingService !== 'undefined' && password_setting_service_1.PasswordSettingService) === 'function' && _c) || Object])
    ], ForgotPasswordComponent);
    return ForgotPasswordComponent;
    var _a, _b, _c;
}());
exports.ForgotPasswordComponent = ForgotPasswordComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/forgot-password.component.js.map

/***/ }),

/***/ 377:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var forms_1 = __webpack_require__(26);
var password_setting_service_1 = __webpack_require__(171);
var session_service_1 = __webpack_require__(14);
var shared_const_1 = __webpack_require__(2);
var message_service_1 = __webpack_require__(10);
var shared_utils_1 = __webpack_require__(33);
var inline_alert_component_1 = __webpack_require__(80);
var PasswordSettingComponent = (function () {
    function PasswordSettingComponent(passwordService, session, msgService) {
        this.passwordService = passwordService;
        this.session = session;
        this.msgService = msgService;
        this.opened = false;
        this.oldPwd = "";
        this.newPwd = "";
        this.reNewPwd = "";
        this.error = null;
        this.formValueChanged = false;
        this.onCalling = false;
    }
    Object.defineProperty(PasswordSettingComponent.prototype, "isValid", {
        //If form is valid
        get: function () {
            if (this.pwdForm && this.pwdForm.form.get("newPassword")) {
                return this.pwdForm.valid &&
                    (this.pwdForm.form.get("newPassword").value === this.pwdForm.form.get("reNewPassword").value) &&
                    this.error === null;
            }
            return false;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(PasswordSettingComponent.prototype, "valueChanged", {
        get: function () {
            return this.formValueChanged;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(PasswordSettingComponent.prototype, "showProgress", {
        get: function () {
            return this.onCalling;
        },
        enumerable: true,
        configurable: true
    });
    PasswordSettingComponent.prototype.ngAfterViewChecked = function () {
        var _this = this;
        if (this.pwdFormRef != this.pwdForm) {
            this.pwdFormRef = this.pwdForm;
            if (this.pwdFormRef) {
                this.pwdFormRef.valueChanges.subscribe(function (data) {
                    _this.formValueChanged = true;
                    _this.error = null;
                    _this.inlineAlert.close();
                });
            }
        }
    };
    //Open modal dialog
    PasswordSettingComponent.prototype.open = function () {
        this.opened = true;
        this.pwdForm.reset();
        this.formValueChanged = false;
    };
    //Close the moal dialog
    PasswordSettingComponent.prototype.close = function () {
        if (this.formValueChanged) {
            if (shared_utils_1.isEmptyForm(this.pwdForm)) {
                this.opened = false;
            }
            else {
                //Need user confirmation
                this.inlineAlert.showInlineConfirmation({
                    message: "ALERT.FORM_CHANGE_CONFIRMATION"
                });
            }
        }
        else {
            this.opened = false;
        }
    };
    PasswordSettingComponent.prototype.confirmCancel = function () {
        this.opened = false;
    };
    //handle the ok action
    PasswordSettingComponent.prototype.doOk = function () {
        var _this = this;
        if (this.onCalling) {
            return; //To avoid duplicate click events
        }
        if (!this.isValid) {
            return; //Double confirm
        }
        //Double confirm session is valid
        var cUser = this.session.getCurrentUser();
        if (!cUser) {
            return;
        }
        //Call service
        this.onCalling = true;
        this.passwordService.changePassword(cUser.user_id, {
            new_password: this.pwdForm.value.newPassword,
            old_password: this.pwdForm.value.oldPassword
        })
            .then(function () {
            _this.onCalling = false;
            _this.opened = false;
            _this.msgService.announceMessage(200, "CHANGE_PWD.SAVE_SUCCESS", shared_const_1.AlertType.SUCCESS);
        })
            .catch(function (error) {
            _this.onCalling = false;
            _this.error = error;
            if (shared_utils_1.accessErrorHandler(error, _this.msgService)) {
                _this.opened = false;
            }
            else {
                _this.inlineAlert.showInlineError(error);
            }
        });
    };
    __decorate([
        core_1.ViewChild("changepwdForm"), 
        __metadata('design:type', (typeof (_a = typeof forms_1.NgForm !== 'undefined' && forms_1.NgForm) === 'function' && _a) || Object)
    ], PasswordSettingComponent.prototype, "pwdForm", void 0);
    __decorate([
        core_1.ViewChild(inline_alert_component_1.InlineAlertComponent), 
        __metadata('design:type', (typeof (_b = typeof inline_alert_component_1.InlineAlertComponent !== 'undefined' && inline_alert_component_1.InlineAlertComponent) === 'function' && _b) || Object)
    ], PasswordSettingComponent.prototype, "inlineAlert", void 0);
    PasswordSettingComponent = __decorate([
        core_1.Component({
            selector: 'password-setting',
            template: __webpack_require__(822)
        }), 
        __metadata('design:paramtypes', [(typeof (_c = typeof password_setting_service_1.PasswordSettingService !== 'undefined' && password_setting_service_1.PasswordSettingService) === 'function' && _c) || Object, (typeof (_d = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _d) || Object, (typeof (_e = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _e) || Object])
    ], PasswordSettingComponent);
    return PasswordSettingComponent;
    var _a, _b, _c, _d, _e;
}());
exports.PasswordSettingComponent = PasswordSettingComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/password-setting.component.js.map

/***/ }),

/***/ 378:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var forms_1 = __webpack_require__(26);
var password_setting_service_1 = __webpack_require__(171);
var inline_alert_component_1 = __webpack_require__(80);
var shared_utils_1 = __webpack_require__(33);
var message_service_1 = __webpack_require__(10);
var ResetPasswordComponent = (function () {
    function ResetPasswordComponent(pwdService, route, msgService, router) {
        this.pwdService = pwdService;
        this.route = route;
        this.msgService = msgService;
        this.router = router;
        this.opened = true;
        this.onGoing = false;
        this.password = "";
        this.validationState = {};
        this.resetUuid = "";
        this.resetOk = false;
    }
    ResetPasswordComponent.prototype.ngOnInit = function () {
        var _this = this;
        this.route.queryParams.subscribe(function (params) { return _this.resetUuid = params["reset_uuid"] || ""; });
    };
    Object.defineProperty(ResetPasswordComponent.prototype, "showProgress", {
        get: function () {
            return this.onGoing;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(ResetPasswordComponent.prototype, "isValid", {
        get: function () {
            return this.resetPwdForm && this.resetPwdForm.valid && this.samePassword();
        },
        enumerable: true,
        configurable: true
    });
    ResetPasswordComponent.prototype.getValidationState = function (key) {
        return this.validationState &&
            this.validationState[key] &&
            key === 'reNewPassword' ? this.samePassword() : true;
    };
    ResetPasswordComponent.prototype.open = function () {
        this.resetOk = false;
        this.opened = true;
        this.resetPwdForm.resetForm();
    };
    ResetPasswordComponent.prototype.close = function () {
        this.opened = false;
    };
    ResetPasswordComponent.prototype.send = function () {
        var _this = this;
        //If already reset password ok, navigator to sign-in
        if (this.resetOk) {
            this.router.navigate(['sign-in']);
            return;
        }
        //Double confirm to avoid improper situations
        if (!this.password) {
            return;
        }
        if (!this.isValid) {
            return;
        }
        this.onGoing = true;
        this.pwdService.resetPassword(this.resetUuid, this.password)
            .then(function () {
            _this.onGoing = false;
            _this.resetOk = true;
            _this.inlineAlert.showInlineSuccess({ message: 'RESET_PWD.RESET_OK' });
        })
            .catch(function (error) {
            _this.onGoing = false;
            if (shared_utils_1.accessErrorHandler(error, _this.msgService)) {
                _this.close();
            }
            else {
                _this.inlineAlert.showInlineError(shared_utils_1.errorHandler(error));
            }
        });
    };
    ResetPasswordComponent.prototype.handleValidation = function (key, flag) {
        if (flag) {
            if (!this.validationState[key]) {
                this.validationState[key] = true;
            }
        }
        else {
            this.validationState[key] = this.getControlValidationState(key);
        }
    };
    ResetPasswordComponent.prototype.getControlValidationState = function (key) {
        if (this.resetPwdForm) {
            var control = this.resetPwdForm.controls[key];
            if (control) {
                return control.valid;
            }
        }
        return false;
    };
    ResetPasswordComponent.prototype.samePassword = function () {
        if (this.resetPwdForm) {
            var control1 = this.resetPwdForm.controls["newPassword"];
            var control2 = this.resetPwdForm.controls["reNewPassword"];
            if (control1 && control2) {
                return control1.value == control2.value;
            }
        }
        return false;
    };
    __decorate([
        core_1.ViewChild("resetPwdForm"), 
        __metadata('design:type', (typeof (_a = typeof forms_1.NgForm !== 'undefined' && forms_1.NgForm) === 'function' && _a) || Object)
    ], ResetPasswordComponent.prototype, "resetPwdForm", void 0);
    __decorate([
        core_1.ViewChild(inline_alert_component_1.InlineAlertComponent), 
        __metadata('design:type', (typeof (_b = typeof inline_alert_component_1.InlineAlertComponent !== 'undefined' && inline_alert_component_1.InlineAlertComponent) === 'function' && _b) || Object)
    ], ResetPasswordComponent.prototype, "inlineAlert", void 0);
    ResetPasswordComponent = __decorate([
        core_1.Component({
            selector: 'reset-password',
            template: __webpack_require__(823),
            styles: [__webpack_require__(457)]
        }), 
        __metadata('design:paramtypes', [(typeof (_c = typeof password_setting_service_1.PasswordSettingService !== 'undefined' && password_setting_service_1.PasswordSettingService) === 'function' && _c) || Object, (typeof (_d = typeof router_1.ActivatedRoute !== 'undefined' && router_1.ActivatedRoute) === 'function' && _d) || Object, (typeof (_e = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _e) || Object, (typeof (_f = typeof router_1.Router !== 'undefined' && router_1.Router) === 'function' && _f) || Object])
    ], ResetPasswordComponent);
    return ResetPasswordComponent;
    var _a, _b, _c, _d, _e, _f;
}());
exports.ResetPasswordComponent = ResetPasswordComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/reset-password.component.js.map

/***/ }),

/***/ 379:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var core_2 = __webpack_require__(0);
var forms_1 = __webpack_require__(26);
var session_service_1 = __webpack_require__(14);
var sign_in_credential_1 = __webpack_require__(632);
var sign_up_component_1 = __webpack_require__(243);
var shared_const_1 = __webpack_require__(2);
var forgot_password_component_1 = __webpack_require__(376);
var app_config_service_1 = __webpack_require__(172);
var app_config_1 = __webpack_require__(244);
//Define status flags for signing in states
exports.signInStatusNormal = 0;
exports.signInStatusOnGoing = 1;
exports.signInStatusError = -1;
var SignInComponent = (function () {
    function SignInComponent(router, session, route, appConfigService) {
        this.router = router;
        this.session = session;
        this.route = route;
        this.appConfigService = appConfigService;
        this.redirectUrl = "";
        this.appConfig = new app_config_1.AppConfig();
        //Status flag
        this.signInStatus = exports.signInStatusNormal;
        //Initialize sign in credential
        this.signInCredential = {
            principal: "",
            password: ""
        };
    }
    SignInComponent.prototype.ngOnInit = function () {
        var _this = this;
        this.appConfig = this.appConfigService.getConfig();
        this.route.queryParams
            .subscribe(function (params) {
            _this.redirectUrl = params["redirect_url"] || "";
            var isSignUp = params["sign_up"] || "";
            if (isSignUp != "") {
                _this.signUp(); //Open sign up
            }
        });
    };
    Object.defineProperty(SignInComponent.prototype, "isError", {
        //For template accessing
        get: function () {
            return this.signInStatus === exports.signInStatusError;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(SignInComponent.prototype, "isOnGoing", {
        get: function () {
            return this.signInStatus === exports.signInStatusOnGoing;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(SignInComponent.prototype, "isValid", {
        //Validate the related fields
        get: function () {
            return this.currentForm.form.valid;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(SignInComponent.prototype, "selfSignUp", {
        //Whether show the 'sign up' link
        get: function () {
            return this.appConfig.auth_mode === 'db_auth'
                && this.appConfig.self_registration;
        },
        enumerable: true,
        configurable: true
    });
    //General error handler
    SignInComponent.prototype.handleError = function (error) {
        //Set error status
        this.signInStatus = exports.signInStatusError;
        var message = error.status ? error.status + ":" + error.statusText : error;
        console.error("An error occurred when signing in:", message);
    };
    //Hande form values changes
    SignInComponent.prototype.formChanged = function () {
        var _this = this;
        if (this.currentForm === this.signInForm) {
            return;
        }
        this.signInForm = this.currentForm;
        if (this.signInForm) {
            this.signInForm.valueChanges
                .subscribe(function (data) {
                _this.updateState();
            });
        }
    };
    //Implement interface
    //Watch the view change only when view is in error state
    SignInComponent.prototype.ngAfterViewChecked = function () {
        if (this.signInStatus === exports.signInStatusError) {
            this.formChanged();
        }
    };
    //Update the status if we have done some changes
    SignInComponent.prototype.updateState = function () {
        if (this.signInStatus === exports.signInStatusError) {
            this.signInStatus = exports.signInStatusNormal; //reset
        }
    };
    //Trigger the signin action
    SignInComponent.prototype.signIn = function () {
        var _this = this;
        //Should validate input firstly
        if (!this.isValid || this.isOnGoing) {
            return;
        }
        //Start signing in progress
        this.signInStatus = exports.signInStatusOnGoing;
        //Call the service to send out the http request
        this.session.signIn(this.signInCredential)
            .then(function () {
            //Set status
            _this.signInStatus = exports.signInStatusNormal;
            //Redirect to the right route
            if (_this.redirectUrl === "") {
                //Routing to the default location
                _this.router.navigateByUrl(shared_const_1.harborRootRoute);
            }
            else {
                _this.router.navigateByUrl(_this.redirectUrl);
            }
        })
            .catch(function (error) {
            _this.handleError(error);
        });
    };
    //Open sign up dialog
    SignInComponent.prototype.signUp = function () {
        this.signUpDialog.open();
    };
    //Open forgot password dialog
    SignInComponent.prototype.forgotPassword = function () {
        this.forgotPwdDialog.open();
    };
    __decorate([
        core_2.ViewChild('signInForm'), 
        __metadata('design:type', (typeof (_a = typeof forms_1.NgForm !== 'undefined' && forms_1.NgForm) === 'function' && _a) || Object)
    ], SignInComponent.prototype, "currentForm", void 0);
    __decorate([
        core_2.ViewChild('signupDialog'), 
        __metadata('design:type', (typeof (_b = typeof sign_up_component_1.SignUpComponent !== 'undefined' && sign_up_component_1.SignUpComponent) === 'function' && _b) || Object)
    ], SignInComponent.prototype, "signUpDialog", void 0);
    __decorate([
        core_2.ViewChild('forgotPwdDialog'), 
        __metadata('design:type', (typeof (_c = typeof forgot_password_component_1.ForgotPasswordComponent !== 'undefined' && forgot_password_component_1.ForgotPasswordComponent) === 'function' && _c) || Object)
    ], SignInComponent.prototype, "forgotPwdDialog", void 0);
    __decorate([
        core_2.Input(), 
        __metadata('design:type', (typeof (_d = typeof sign_in_credential_1.SignInCredential !== 'undefined' && sign_in_credential_1.SignInCredential) === 'function' && _d) || Object)
    ], SignInComponent.prototype, "signInCredential", void 0);
    SignInComponent = __decorate([
        core_1.Component({
            selector: 'sign-in',
            template: __webpack_require__(824),
            styles: [__webpack_require__(802)]
        }), 
        __metadata('design:paramtypes', [(typeof (_e = typeof router_1.Router !== 'undefined' && router_1.Router) === 'function' && _e) || Object, (typeof (_f = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _f) || Object, (typeof (_g = typeof router_1.ActivatedRoute !== 'undefined' && router_1.ActivatedRoute) === 'function' && _g) || Object, (typeof (_h = typeof app_config_service_1.AppConfigService !== 'undefined' && app_config_service_1.AppConfigService) === 'function' && _h) || Object])
    ], SignInComponent);
    return SignInComponent;
    var _a, _b, _c, _d, _e, _f, _g, _h;
}());
exports.SignInComponent = SignInComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/sign-in.component.js.map

/***/ }),

/***/ 380:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var core_2 = __webpack_require__(34);
var core_3 = __webpack_require__(257);
var shared_const_1 = __webpack_require__(2);
var session_service_1 = __webpack_require__(14);
var AppComponent = (function () {
    function AppComponent(translate, cookie, session) {
        this.translate = translate;
        this.cookie = cookie;
        this.session = session;
        translate.addLangs(shared_const_1.supportedLangs);
        translate.setDefaultLang(shared_const_1.enLang);
        //If user has selected lang, then directly use it
        var langSetting = this.cookie.get("harbor-lang");
        if (!langSetting || langSetting.trim() === "") {
            //Use browser lang
            langSetting = translate.getBrowserLang();
        }
        var selectedLang = this.isLangMatch(langSetting, shared_const_1.supportedLangs) ? langSetting : shared_const_1.enLang;
        translate.use(selectedLang);
        //this.session.switchLanguage(selectedLang).catch(error => console.error(error));
    }
    AppComponent.prototype.isLangMatch = function (browserLang, supportedLangs) {
        if (supportedLangs && supportedLangs.length > 0) {
            return supportedLangs.find(function (lang) { return lang === browserLang; });
        }
    };
    AppComponent = __decorate([
        core_1.Component({
            selector: 'harbor-app',
            template: __webpack_require__(826),
            styleUrls: []
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof core_2.TranslateService !== 'undefined' && core_2.TranslateService) === 'function' && _a) || Object, (typeof (_b = typeof core_3.CookieService !== 'undefined' && core_3.CookieService) === 'function' && _b) || Object, (typeof (_c = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _c) || Object])
    ], AppComponent);
    return AppComponent;
    var _a, _b, _c;
}());
exports.AppComponent = AppComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/app.component.js.map

/***/ }),

/***/ 381:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var global_search_service_1 = __webpack_require__(604);
var search_results_1 = __webpack_require__(605);
var shared_utils_1 = __webpack_require__(33);
var shared_const_1 = __webpack_require__(2);
var message_service_1 = __webpack_require__(10);
var search_trigger_service_1 = __webpack_require__(95);
var SearchResultComponent = (function () {
    function SearchResultComponent(search, msgService, searchTrigger) {
        this.search = search;
        this.msgService = msgService;
        this.searchTrigger = searchTrigger;
        this.searchResults = new search_results_1.SearchResults();
        this.currentTerm = "";
        //Open or close
        this.stateIndicator = false;
        //Search in progress
        this.onGoing = false;
        //Whether or not mouse point is onto the close indicator
        this.mouseOn = false;
    }
    SearchResultComponent.prototype.doFilterProjects = function (event) {
        this.searchResults.project = this.originalCopy.project.filter(function (pro) { return pro.name.indexOf(event) != -1; });
    };
    SearchResultComponent.prototype.clone = function (src) {
        var res = new search_results_1.SearchResults();
        if (src) {
            src.project.forEach(function (pro) { return res.project.push(Object.assign({}, pro)); });
            src.repository.forEach(function (repo) { return res.repository.push(Object.assign({}, repo)); });
            return res;
        }
        return res; //Empty object
    };
    Object.defineProperty(SearchResultComponent.prototype, "listMode", {
        get: function () {
            return shared_const_1.ListMode.READONLY;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(SearchResultComponent.prototype, "state", {
        get: function () {
            return this.stateIndicator;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(SearchResultComponent.prototype, "done", {
        get: function () {
            return !this.onGoing;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(SearchResultComponent.prototype, "hover", {
        get: function () {
            return this.mouseOn;
        },
        enumerable: true,
        configurable: true
    });
    //Handle mouse event of close indicator
    SearchResultComponent.prototype.mouseAction = function (over) {
        this.mouseOn = over;
    };
    //Show the results
    SearchResultComponent.prototype.show = function () {
        this.stateIndicator = true;
        this.searchTrigger.searchInputStat(true);
    };
    //Close the result page
    SearchResultComponent.prototype.close = function () {
        //Tell shell close
        this.searchTrigger.closeSearch(true);
        this.searchTrigger.searchInputStat(false);
        this.stateIndicator = false;
    };
    //Call search service to complete the search request
    SearchResultComponent.prototype.doSearch = function (term) {
        var _this = this;
        //Do nothing if search is ongoing
        if (this.onGoing) {
            return;
        }
        //Confirm page is displayed
        if (!this.stateIndicator) {
            this.show();
        }
        this.currentTerm = term;
        //If term is empty, then clear the results
        if (term === "") {
            this.searchResults.project = [];
            this.searchResults.repository = [];
            return;
        }
        //Show spinner
        this.onGoing = true;
        this.search.doSearch(term)
            .then(function (searchResults) {
            _this.onGoing = false;
            _this.originalCopy = searchResults; //Keeo the original data
            _this.searchResults = _this.clone(searchResults);
        })
            .catch(function (error) {
            _this.onGoing = false;
            if (!shared_utils_1.accessErrorHandler(error, _this.msgService)) {
                _this.msgService.announceMessage(error.status, shared_utils_1.errorHandler(error), shared_const_1.AlertType.DANGER);
            }
        });
    };
    SearchResultComponent = __decorate([
        core_1.Component({
            selector: "search-result",
            template: __webpack_require__(829),
            styles: [__webpack_require__(803)],
            providers: [global_search_service_1.GlobalSearchService]
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof global_search_service_1.GlobalSearchService !== 'undefined' && global_search_service_1.GlobalSearchService) === 'function' && _a) || Object, (typeof (_b = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _b) || Object, (typeof (_c = typeof search_trigger_service_1.SearchTriggerService !== 'undefined' && search_trigger_service_1.SearchTriggerService) === 'function' && _c) || Object])
    ], SearchResultComponent);
    return SearchResultComponent;
    var _a, _b, _c;
}());
exports.SearchResultComponent = SearchResultComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/search-result.component.js.map

/***/ }),

/***/ 382:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var modal_events_const_1 = __webpack_require__(383);
var account_settings_modal_component_1 = __webpack_require__(374);
var search_result_component_1 = __webpack_require__(381);
var password_setting_component_1 = __webpack_require__(377);
var navigator_component_1 = __webpack_require__(384);
var session_service_1 = __webpack_require__(14);
var about_dialog_component_1 = __webpack_require__(407);
var start_component_1 = __webpack_require__(245);
var search_trigger_service_1 = __webpack_require__(95);
var shared_const_1 = __webpack_require__(2);
var HarborShellComponent = (function () {
    function HarborShellComponent(route, router, session, searchTrigger) {
        this.route = route;
        this.router = router;
        this.session = session;
        this.searchTrigger = searchTrigger;
        //To indicator whwther or not the search results page is displayed
        //We need to use this property to do some overriding work
        this.isSearchResultsOpened = false;
    }
    HarborShellComponent.prototype.ngOnInit = function () {
        var _this = this;
        this.searchSub = this.searchTrigger.searchTriggerChan$.subscribe(function (searchEvt) {
            _this.doSearch(searchEvt);
        });
        this.searchCloseSub = this.searchTrigger.searchCloseChan$.subscribe(function (close) {
            if (close) {
                _this.searchClose();
            }
            else {
                _this.watchClickEvt(); //reuse
            }
        });
    };
    HarborShellComponent.prototype.ngOnDestroy = function () {
        if (this.searchSub) {
            this.searchSub.unsubscribe();
        }
        if (this.searchCloseSub) {
            this.searchCloseSub.unsubscribe();
        }
    };
    Object.defineProperty(HarborShellComponent.prototype, "isStartPage", {
        get: function () {
            return this.router.routerState.snapshot.url.toString() === shared_const_1.harborRootRoute;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(HarborShellComponent.prototype, "showSearch", {
        get: function () {
            return this.isSearchResultsOpened;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(HarborShellComponent.prototype, "isSystemAdmin", {
        get: function () {
            var account = this.session.getCurrentUser();
            return account != null && account.has_admin_role > 0;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(HarborShellComponent.prototype, "isUserExisting", {
        get: function () {
            var account = this.session.getCurrentUser();
            return account != null;
        },
        enumerable: true,
        configurable: true
    });
    //Open modal dialog
    HarborShellComponent.prototype.openModal = function (event) {
        switch (event.modalName) {
            case modal_events_const_1.modalEvents.USER_PROFILE:
                this.accountSettingsModal.open();
                break;
            case modal_events_const_1.modalEvents.CHANGE_PWD:
                this.pwdSetting.open();
                break;
            case modal_events_const_1.modalEvents.ABOUT:
                this.aboutDialog.open();
                break;
            default:
                break;
        }
    };
    //Handle the global search event and then let the result page to trigger api
    HarborShellComponent.prototype.doSearch = function (event) {
        if (event === "") {
            if (!this.isSearchResultsOpened) {
                //Will not open search result panel if term is empty
                return;
            }
            else {
                //If opened, then close the search result panel
                this.isSearchResultsOpened = false;
                this.searchResultComponet.close();
                return;
            }
        }
        //Once this method is called
        //the search results page must be opened
        this.isSearchResultsOpened = true;
        //Call the child component to do the real work
        this.searchResultComponet.doSearch(event);
    };
    //Search results page closed
    //remove the related ovevriding things
    HarborShellComponent.prototype.searchClose = function () {
        this.isSearchResultsOpened = false;
    };
    //Close serch result panel if existing
    HarborShellComponent.prototype.watchClickEvt = function () {
        this.searchResultComponet.close();
        this.isSearchResultsOpened = false;
    };
    __decorate([
        core_1.ViewChild(account_settings_modal_component_1.AccountSettingsModalComponent), 
        __metadata('design:type', (typeof (_a = typeof account_settings_modal_component_1.AccountSettingsModalComponent !== 'undefined' && account_settings_modal_component_1.AccountSettingsModalComponent) === 'function' && _a) || Object)
    ], HarborShellComponent.prototype, "accountSettingsModal", void 0);
    __decorate([
        core_1.ViewChild(search_result_component_1.SearchResultComponent), 
        __metadata('design:type', (typeof (_b = typeof search_result_component_1.SearchResultComponent !== 'undefined' && search_result_component_1.SearchResultComponent) === 'function' && _b) || Object)
    ], HarborShellComponent.prototype, "searchResultComponet", void 0);
    __decorate([
        core_1.ViewChild(password_setting_component_1.PasswordSettingComponent), 
        __metadata('design:type', (typeof (_c = typeof password_setting_component_1.PasswordSettingComponent !== 'undefined' && password_setting_component_1.PasswordSettingComponent) === 'function' && _c) || Object)
    ], HarborShellComponent.prototype, "pwdSetting", void 0);
    __decorate([
        core_1.ViewChild(navigator_component_1.NavigatorComponent), 
        __metadata('design:type', (typeof (_d = typeof navigator_component_1.NavigatorComponent !== 'undefined' && navigator_component_1.NavigatorComponent) === 'function' && _d) || Object)
    ], HarborShellComponent.prototype, "navigator", void 0);
    __decorate([
        core_1.ViewChild(about_dialog_component_1.AboutDialogComponent), 
        __metadata('design:type', (typeof (_e = typeof about_dialog_component_1.AboutDialogComponent !== 'undefined' && about_dialog_component_1.AboutDialogComponent) === 'function' && _e) || Object)
    ], HarborShellComponent.prototype, "aboutDialog", void 0);
    __decorate([
        core_1.ViewChild(start_component_1.StartPageComponent), 
        __metadata('design:type', (typeof (_f = typeof start_component_1.StartPageComponent !== 'undefined' && start_component_1.StartPageComponent) === 'function' && _f) || Object)
    ], HarborShellComponent.prototype, "searchSatrt", void 0);
    HarborShellComponent = __decorate([
        core_1.Component({
            selector: 'harbor-shell',
            template: __webpack_require__(830),
            styles: [__webpack_require__(804)]
        }), 
        __metadata('design:paramtypes', [(typeof (_g = typeof router_1.ActivatedRoute !== 'undefined' && router_1.ActivatedRoute) === 'function' && _g) || Object, (typeof (_h = typeof router_1.Router !== 'undefined' && router_1.Router) === 'function' && _h) || Object, (typeof (_j = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _j) || Object, (typeof (_k = typeof search_trigger_service_1.SearchTriggerService !== 'undefined' && search_trigger_service_1.SearchTriggerService) === 'function' && _k) || Object])
    ], HarborShellComponent);
    return HarborShellComponent;
    var _a, _b, _c, _d, _e, _f, _g, _h, _j, _k;
}());
exports.HarborShellComponent = HarborShellComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/harbor-shell.component.js.map

/***/ }),

/***/ 383:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

exports.modalEvents = {
    USER_PROFILE: "USER_PROFILE",
    CHANGE_PWD: "CHANGE_PWD",
    ABOUT: "ABOUT"
};
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/modal-events.const.js.map

/***/ }),

/***/ 384:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var core_2 = __webpack_require__(34);
var modal_events_const_1 = __webpack_require__(383);
var session_service_1 = __webpack_require__(14);
var core_3 = __webpack_require__(257);
var shared_const_1 = __webpack_require__(2);
var app_config_service_1 = __webpack_require__(172);
var app_config_1 = __webpack_require__(244);
var NavigatorComponent = (function () {
    function NavigatorComponent(session, router, translate, cookie, appConfigService) {
        this.session = session;
        this.router = router;
        this.translate = translate;
        this.cookie = cookie;
        this.appConfigService = appConfigService;
        // constructor(private router: Router){}
        this.showAccountSettingsModal = new core_1.EventEmitter();
        this.showPwdChangeModal = new core_1.EventEmitter();
        this.sessionUser = null;
        this.selectedLang = shared_const_1.enLang;
        this.appConfig = new app_config_1.AppConfig();
    }
    NavigatorComponent.prototype.ngOnInit = function () {
        var _this = this;
        this.sessionUser = this.session.getCurrentUser();
        this.selectedLang = this.translate.currentLang;
        this.translate.onLangChange.subscribe(function (langChange) {
            _this.selectedLang = langChange.lang;
            //Keep in cookie for next use
            _this.cookie.put("harbor-lang", langChange.lang);
        });
        this.appConfig = this.appConfigService.getConfig();
    };
    Object.defineProperty(NavigatorComponent.prototype, "isSessionValid", {
        get: function () {
            return this.sessionUser != null;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(NavigatorComponent.prototype, "accountName", {
        get: function () {
            return this.sessionUser ? this.sessionUser.username : "";
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(NavigatorComponent.prototype, "currentLang", {
        get: function () {
            return shared_const_1.languageNames[this.selectedLang];
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(NavigatorComponent.prototype, "isIntegrationMode", {
        get: function () {
            return this.appConfig.with_admiral && this.appConfig.admiral_endpoint.trim() != "";
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(NavigatorComponent.prototype, "admiralLink", {
        get: function () {
            var routeSegments = [this.appConfig.admiral_endpoint,
                "?registry_url=",
                encodeURIComponent(window.location.href)
            ];
            return routeSegments.join("");
        },
        enumerable: true,
        configurable: true
    });
    NavigatorComponent.prototype.matchLang = function (lang) {
        return lang.trim() === this.selectedLang;
    };
    //Open the account setting dialog
    NavigatorComponent.prototype.openAccountSettingsModal = function () {
        this.showAccountSettingsModal.emit({
            modalName: modal_events_const_1.modalEvents.USER_PROFILE,
            modalFlag: true
        });
    };
    //Open change password dialog
    NavigatorComponent.prototype.openChangePwdModal = function () {
        this.showPwdChangeModal.emit({
            modalName: modal_events_const_1.modalEvents.CHANGE_PWD,
            modalFlag: true
        });
    };
    //Open about dialog
    NavigatorComponent.prototype.openAboutDialog = function () {
        this.showPwdChangeModal.emit({
            modalName: modal_events_const_1.modalEvents.ABOUT,
            modalFlag: true
        });
    };
    //Log out system
    NavigatorComponent.prototype.logOut = function () {
        var _this = this;
        this.session.signOff()
            .then(function () {
            _this.sessionUser = null;
            //Naviagte to the sign in route
            _this.router.navigate(["/sign-in"]);
        })
            .catch(); //TODO:
    };
    //Switch languages
    NavigatorComponent.prototype.switchLanguage = function (lang) {
        if (shared_const_1.supportedLangs.find(function (supportedLang) { return supportedLang === lang.trim(); })) {
            this.translate.use(lang);
        }
        else {
            this.translate.use(shared_const_1.enLang); //Use default
            //TODO:
            console.error('Language ' + lang.trim() + ' is not suppoted');
        }
        //Try to switch backend lang
        //this.session.switchLanguage(lang).catch(error => console.error(error));
    };
    //Handle the home action
    NavigatorComponent.prototype.homeAction = function () {
        if (this.sessionUser != null) {
            //Navigate to default page
            this.router.navigate(['harbor']);
        }
        else {
            //Naviagte to signin page
            this.router.navigate(['sign-in']);
        }
    };
    NavigatorComponent.prototype.openSignUp = function () {
        var navigatorExtra = {
            queryParams: { "sign_up": true }
        };
        this.router.navigate([shared_const_1.signInRoute], navigatorExtra);
    };
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], NavigatorComponent.prototype, "showAccountSettingsModal", void 0);
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], NavigatorComponent.prototype, "showPwdChangeModal", void 0);
    NavigatorComponent = __decorate([
        core_1.Component({
            selector: 'navigator',
            template: __webpack_require__(831),
            styles: [__webpack_require__(805)]
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _a) || Object, (typeof (_b = typeof router_1.Router !== 'undefined' && router_1.Router) === 'function' && _b) || Object, (typeof (_c = typeof core_2.TranslateService !== 'undefined' && core_2.TranslateService) === 'function' && _c) || Object, (typeof (_d = typeof core_3.CookieService !== 'undefined' && core_3.CookieService) === 'function' && _d) || Object, (typeof (_e = typeof app_config_service_1.AppConfigService !== 'undefined' && app_config_service_1.AppConfigService) === 'function' && _e) || Object])
    ], NavigatorComponent);
    return NavigatorComponent;
    var _a, _b, _c, _d, _e;
}());
exports.NavigatorComponent = NavigatorComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/navigator.component.js.map

/***/ }),

/***/ 385:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var forms_1 = __webpack_require__(26);
var config_1 = __webpack_require__(173);
var ConfigurationAuthComponent = (function () {
    function ConfigurationAuthComponent() {
        this.currentConfig = new config_1.Configuration();
    }
    Object.defineProperty(ConfigurationAuthComponent.prototype, "showLdap", {
        get: function () {
            return this.currentConfig &&
                this.currentConfig.auth_mode &&
                this.currentConfig.auth_mode.value === 'ldap_auth';
        },
        enumerable: true,
        configurable: true
    });
    ConfigurationAuthComponent.prototype.disabled = function (prop) {
        return !(prop && prop.editable);
    };
    ConfigurationAuthComponent.prototype.isValid = function () {
        return this.authForm && this.authForm.valid;
    };
    __decorate([
        core_1.Input("ldapConfig"), 
        __metadata('design:type', (typeof (_a = typeof config_1.Configuration !== 'undefined' && config_1.Configuration) === 'function' && _a) || Object)
    ], ConfigurationAuthComponent.prototype, "currentConfig", void 0);
    __decorate([
        core_1.ViewChild("authConfigFrom"), 
        __metadata('design:type', (typeof (_b = typeof forms_1.NgForm !== 'undefined' && forms_1.NgForm) === 'function' && _b) || Object)
    ], ConfigurationAuthComponent.prototype, "authForm", void 0);
    ConfigurationAuthComponent = __decorate([
        core_1.Component({
            selector: 'config-auth',
            template: __webpack_require__(833),
            styles: [__webpack_require__(277)]
        }), 
        __metadata('design:paramtypes', [])
    ], ConfigurationAuthComponent);
    return ConfigurationAuthComponent;
    var _a, _b;
}());
exports.ConfigurationAuthComponent = ConfigurationAuthComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/config-auth.component.js.map

/***/ }),

/***/ 386:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var forms_1 = __webpack_require__(26);
var config_service_1 = __webpack_require__(387);
var config_1 = __webpack_require__(173);
var message_service_1 = __webpack_require__(10);
var shared_const_1 = __webpack_require__(2);
var shared_utils_1 = __webpack_require__(33);
var config_2 = __webpack_require__(173);
var deletion_dialog_service_1 = __webpack_require__(48);
var deletion_message_1 = __webpack_require__(68);
var config_auth_component_1 = __webpack_require__(385);
var config_email_component_1 = __webpack_require__(388);
var app_config_service_1 = __webpack_require__(172);
var fakePass = "fakepassword";
var ConfigurationComponent = (function () {
    function ConfigurationComponent(msgService, configService, confirmService, appConfigService) {
        this.msgService = msgService;
        this.configService = configService;
        this.confirmService = confirmService;
        this.appConfigService = appConfigService;
        this.onGoing = false;
        this.allConfig = new config_1.Configuration();
        this.currentTabId = "";
        this.testingOnGoing = false;
    }
    ConfigurationComponent.prototype.ngOnInit = function () {
        var _this = this;
        //First load
        this.retrieveConfig();
        this.confirmSub = this.confirmService.deletionConfirm$.subscribe(function (confirmation) {
            _this.reset(confirmation.data);
        });
    };
    ConfigurationComponent.prototype.ngOnDestroy = function () {
        if (this.confirmSub) {
            this.confirmSub.unsubscribe();
        }
    };
    Object.defineProperty(ConfigurationComponent.prototype, "inProgress", {
        get: function () {
            return this.onGoing;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(ConfigurationComponent.prototype, "testingInProgress", {
        get: function () {
            return this.testingOnGoing;
        },
        enumerable: true,
        configurable: true
    });
    ConfigurationComponent.prototype.isValid = function () {
        return this.repoConfigForm &&
            this.repoConfigForm.valid &&
            this.systemConfigForm &&
            this.systemConfigForm.valid &&
            this.mailConfig &&
            this.mailConfig.isValid() &&
            this.authConfig &&
            this.authConfig.isValid();
    };
    ConfigurationComponent.prototype.hasChanges = function () {
        return !this.isEmpty(this.getChanges());
    };
    ConfigurationComponent.prototype.isMailConfigValid = function () {
        return this.mailConfig &&
            this.mailConfig.isValid();
    };
    Object.defineProperty(ConfigurationComponent.prototype, "showTestServerBtn", {
        get: function () {
            return this.currentTabId === 'config-email';
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(ConfigurationComponent.prototype, "showLdapServerBtn", {
        get: function () {
            return this.currentTabId === 'config-auth' &&
                this.allConfig.auth_mode &&
                this.allConfig.auth_mode.value === "ldap_auth";
        },
        enumerable: true,
        configurable: true
    });
    ConfigurationComponent.prototype.isLDAPConfigValid = function () {
        return this.authConfig && this.authConfig.isValid();
    };
    ConfigurationComponent.prototype.tabLinkChanged = function (tabLink) {
        this.currentTabId = tabLink.id;
    };
    /**
     *
     * Save the changed values
     *
     * @memberOf ConfigurationComponent
     */
    ConfigurationComponent.prototype.save = function () {
        var _this = this;
        var changes = this.getChanges();
        if (!this.isEmpty(changes)) {
            this.onGoing = true;
            this.configService.saveConfiguration(changes)
                .then(function (response) {
                _this.onGoing = false;
                //API should return the updated configurations here
                //Unfortunately API does not do that
                //To refresh the view, we can clone the original data copy
                //or force refresh by calling service.
                //HERE we choose force way
                _this.retrieveConfig();
                //Reload bootstrap option
                _this.appConfigService.load().catch(function (error) { return console.error("Failed to reload bootstrap option with error: ", error); });
                _this.msgService.announceMessage(response.status, "CONFIG.SAVE_SUCCESS", shared_const_1.AlertType.SUCCESS);
            })
                .catch(function (error) {
                _this.onGoing = false;
                if (!shared_utils_1.accessErrorHandler(error, _this.msgService)) {
                    _this.msgService.announceMessage(error.status, shared_utils_1.errorHandler(error), shared_const_1.AlertType.DANGER);
                }
            });
        }
        else {
            //Inprop situation, should not come here
            console.error("Save obort becasue nothing changed");
        }
    };
    /**
     *
     * Discard current changes if have and reset
     *
     * @memberOf ConfigurationComponent
     */
    ConfigurationComponent.prototype.cancel = function () {
        var changes = this.getChanges();
        if (!this.isEmpty(changes)) {
            var msg = new deletion_message_1.DeletionMessage("CONFIG.CONFIRM_TITLE", "CONFIG.CONFIRM_SUMMARY", "", changes, shared_const_1.DeletionTargets.EMPTY);
            this.confirmService.openComfirmDialog(msg);
        }
        else {
            //Inprop situation, should not come here
            console.error("Nothing changed");
        }
    };
    /**
     *
     * Test the connection of specified mail server
     *
     *
     * @memberOf ConfigurationComponent
     */
    ConfigurationComponent.prototype.testMailServer = function () {
        var _this = this;
        var mailSettings = {};
        var allChanges = this.getChanges();
        for (var prop in allChanges) {
            if (prop.startsWith("email_")) {
                mailSettings[prop] = allChanges[prop];
            }
        }
        this.testingOnGoing = true;
        this.configService.testMailServer(mailSettings)
            .then(function (response) {
            _this.testingOnGoing = false;
            _this.msgService.announceMessage(200, "CONFIG.TEST_MAIL_SUCCESS", shared_const_1.AlertType.SUCCESS);
        })
            .catch(function (error) {
            _this.testingOnGoing = false;
            _this.msgService.announceMessage(error.status, shared_utils_1.errorHandler(error), shared_const_1.AlertType.WARNING);
        });
    };
    ConfigurationComponent.prototype.testLDAPServer = function () {
        var _this = this;
        var ldapSettings = {};
        var allChanges = this.getChanges();
        for (var prop in allChanges) {
            if (prop.startsWith("ldap_")) {
                ldapSettings[prop] = allChanges[prop];
            }
        }
        console.info(ldapSettings);
        this.testingOnGoing = true;
        this.configService.testLDAPServer(ldapSettings)
            .then(function (respone) {
            _this.testingOnGoing = false;
            _this.msgService.announceMessage(200, "CONFIG.TEST_LDAP_SUCCESS", shared_const_1.AlertType.SUCCESS);
        })
            .catch(function (error) {
            _this.testingOnGoing = false;
            _this.msgService.announceMessage(error.status, shared_utils_1.errorHandler(error), shared_const_1.AlertType.WARNING);
        });
    };
    ConfigurationComponent.prototype.retrieveConfig = function () {
        var _this = this;
        this.onGoing = true;
        this.configService.getConfiguration()
            .then(function (configurations) {
            _this.onGoing = false;
            //Add two password fields
            configurations.email_password = new config_2.StringValueItem(fakePass, true);
            configurations.ldap_search_password = new config_2.StringValueItem(fakePass, true);
            _this.allConfig = configurations;
            //Keep the original copy of the data
            _this.originalCopy = _this.clone(configurations);
        })
            .catch(function (error) {
            _this.onGoing = false;
            if (!shared_utils_1.accessErrorHandler(error, _this.msgService)) {
                _this.msgService.announceMessage(error.status, shared_utils_1.errorHandler(error), shared_const_1.AlertType.DANGER);
            }
        });
    };
    /**
     *
     * Get the changed fields and return a map
     *
     * @private
     * @returns {*}
     *
     * @memberOf ConfigurationComponent
     */
    ConfigurationComponent.prototype.getChanges = function () {
        var changes = {};
        if (!this.allConfig || !this.originalCopy) {
            return changes;
        }
        for (var prop in this.allConfig) {
            var field = this.originalCopy[prop];
            if (field && field.editable) {
                if (field.value != this.allConfig[prop].value) {
                    changes[prop] = this.allConfig[prop].value;
                    //Fix boolean issue
                    if (typeof field.value === "boolean") {
                        changes[prop] = changes[prop] ? "1" : "0";
                    }
                }
            }
        }
        return changes;
    };
    /**
     *
     * Deep clone the configuration object
     *
     * @private
     * @param {Configuration} src
     * @returns {Configuration}
     *
     * @memberOf ConfigurationComponent
     */
    ConfigurationComponent.prototype.clone = function (src) {
        var dest = new config_1.Configuration();
        if (!src) {
            return dest; //Empty
        }
        for (var prop in src) {
            if (src[prop]) {
                dest[prop] = Object.assign({}, src[prop]); //Deep copy inner object
            }
        }
        return dest;
    };
    /**
     *
     * Reset the configuration form
     *
     * @private
     * @param {*} changes
     *
     * @memberOf ConfigurationComponent
     */
    ConfigurationComponent.prototype.reset = function (changes) {
        if (!this.isEmpty(changes)) {
            for (var prop in changes) {
                if (this.originalCopy[prop]) {
                    this.allConfig[prop] = Object.assign({}, this.originalCopy[prop]);
                }
            }
        }
        else {
            //force reset
            this.retrieveConfig();
        }
    };
    ConfigurationComponent.prototype.isEmpty = function (obj) {
        for (var key in obj) {
            if (obj.hasOwnProperty(key))
                return false;
        }
        return true;
    };
    ConfigurationComponent.prototype.disabled = function (prop) {
        return !(prop && prop.editable);
    };
    __decorate([
        core_1.ViewChild("repoConfigFrom"), 
        __metadata('design:type', (typeof (_a = typeof forms_1.NgForm !== 'undefined' && forms_1.NgForm) === 'function' && _a) || Object)
    ], ConfigurationComponent.prototype, "repoConfigForm", void 0);
    __decorate([
        core_1.ViewChild("systemConfigFrom"), 
        __metadata('design:type', (typeof (_b = typeof forms_1.NgForm !== 'undefined' && forms_1.NgForm) === 'function' && _b) || Object)
    ], ConfigurationComponent.prototype, "systemConfigForm", void 0);
    __decorate([
        core_1.ViewChild(config_email_component_1.ConfigurationEmailComponent), 
        __metadata('design:type', (typeof (_c = typeof config_email_component_1.ConfigurationEmailComponent !== 'undefined' && config_email_component_1.ConfigurationEmailComponent) === 'function' && _c) || Object)
    ], ConfigurationComponent.prototype, "mailConfig", void 0);
    __decorate([
        core_1.ViewChild(config_auth_component_1.ConfigurationAuthComponent), 
        __metadata('design:type', (typeof (_d = typeof config_auth_component_1.ConfigurationAuthComponent !== 'undefined' && config_auth_component_1.ConfigurationAuthComponent) === 'function' && _d) || Object)
    ], ConfigurationComponent.prototype, "authConfig", void 0);
    ConfigurationComponent = __decorate([
        core_1.Component({
            selector: 'config',
            template: __webpack_require__(834),
            styles: [__webpack_require__(277)]
        }), 
        __metadata('design:paramtypes', [(typeof (_e = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _e) || Object, (typeof (_f = typeof config_service_1.ConfigurationService !== 'undefined' && config_service_1.ConfigurationService) === 'function' && _f) || Object, (typeof (_g = typeof deletion_dialog_service_1.DeletionDialogService !== 'undefined' && deletion_dialog_service_1.DeletionDialogService) === 'function' && _g) || Object, (typeof (_h = typeof app_config_service_1.AppConfigService !== 'undefined' && app_config_service_1.AppConfigService) === 'function' && _h) || Object])
    ], ConfigurationComponent);
    return ConfigurationComponent;
    var _a, _b, _c, _d, _e, _f, _g, _h;
}());
exports.ConfigurationComponent = ConfigurationComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/config.component.js.map

/***/ }),

/***/ 387:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var http_1 = __webpack_require__(20);
__webpack_require__(58);
var configEndpoint = "/api/configurations";
var emailEndpoint = "/api/email/ping";
var ldapEndpoint = "/api/ldap/ping";
var ConfigurationService = (function () {
    function ConfigurationService(http) {
        this.http = http;
        this.headers = new http_1.Headers({
            "Accept": 'application/json',
            "Content-Type": 'application/json'
        });
        this.options = new http_1.RequestOptions({
            'headers': this.headers
        });
    }
    ConfigurationService.prototype.getConfiguration = function () {
        return this.http.get(configEndpoint, this.options).toPromise()
            .then(function (response) { return response.json(); })
            .catch(function (error) { return Promise.reject(error); });
    };
    ConfigurationService.prototype.saveConfiguration = function (values) {
        return this.http.put(configEndpoint, JSON.stringify(values), this.options)
            .toPromise()
            .then(function (response) { return response; })
            .catch(function (error) { return Promise.reject(error); });
    };
    ConfigurationService.prototype.testMailServer = function (mailSettings) {
        return this.http.post(emailEndpoint, JSON.stringify(mailSettings), this.options)
            .toPromise()
            .then(function (response) { return response; })
            .catch(function (error) { return Promise.reject(error); });
    };
    ConfigurationService.prototype.testLDAPServer = function (ldapSettings) {
        return this.http.post(ldapEndpoint, JSON.stringify(ldapSettings), this.options)
            .toPromise()
            .then(function (response) { return response; })
            .catch(function (error) { return Promise.reject(error); });
    };
    ConfigurationService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof http_1.Http !== 'undefined' && http_1.Http) === 'function' && _a) || Object])
    ], ConfigurationService);
    return ConfigurationService;
    var _a;
}());
exports.ConfigurationService = ConfigurationService;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/config.service.js.map

/***/ }),

/***/ 388:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var forms_1 = __webpack_require__(26);
var config_1 = __webpack_require__(173);
var ConfigurationEmailComponent = (function () {
    function ConfigurationEmailComponent() {
        this.currentConfig = new config_1.Configuration();
    }
    ConfigurationEmailComponent.prototype.disabled = function (prop) {
        return !(prop && prop.editable);
    };
    ConfigurationEmailComponent.prototype.isValid = function () {
        return this.mailForm && this.mailForm.valid;
    };
    __decorate([
        core_1.Input("mailConfig"), 
        __metadata('design:type', (typeof (_a = typeof config_1.Configuration !== 'undefined' && config_1.Configuration) === 'function' && _a) || Object)
    ], ConfigurationEmailComponent.prototype, "currentConfig", void 0);
    __decorate([
        core_1.ViewChild("mailConfigFrom"), 
        __metadata('design:type', (typeof (_b = typeof forms_1.NgForm !== 'undefined' && forms_1.NgForm) === 'function' && _b) || Object)
    ], ConfigurationEmailComponent.prototype, "mailForm", void 0);
    ConfigurationEmailComponent = __decorate([
        core_1.Component({
            selector: 'config-email',
            template: __webpack_require__(835),
            styles: [__webpack_require__(277)]
        }), 
        __metadata('design:paramtypes', [])
    ], ConfigurationEmailComponent);
    return ConfigurationEmailComponent;
    var _a, _b;
}());
exports.ConfigurationEmailComponent = ConfigurationEmailComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/config-email.component.js.map

/***/ }),

/***/ 389:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var shared_const_1 = __webpack_require__(2);
var Message = (function () {
    function Message() {
        this.isAppLevel = false;
    }
    Object.defineProperty(Message.prototype, "type", {
        get: function () {
            switch (this.alertType) {
                case shared_const_1.AlertType.DANGER:
                    return 'alert-danger';
                case shared_const_1.AlertType.INFO:
                    return 'alert-info';
                case shared_const_1.AlertType.SUCCESS:
                    return 'alert-success';
                case shared_const_1.AlertType.WARNING:
                    return 'alert-warning';
                default:
                    return 'alert-warning';
            }
        },
        enumerable: true,
        configurable: true
    });
    Message.newMessage = function (statusCode, message, alertType) {
        var m = new Message();
        m.statusCode = statusCode;
        m.message = message;
        m.alertType = alertType;
        return m;
    };
    Message.prototype.toString = function () {
        return 'Message with statusCode:' + this.statusCode +
            ', message:' + this.message +
            ', alert type:' + this.type;
    };
    return Message;
}());
exports.Message = Message;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/message.js.map

/***/ }),

/***/ 390:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var audit_log_1 = __webpack_require__(611);
var audit_log_service_1 = __webpack_require__(247);
var message_service_1 = __webpack_require__(10);
var shared_const_1 = __webpack_require__(2);
var optionalSearch = { 0: 'AUDIT_LOG.ADVANCED', 1: 'AUDIT_LOG.SIMPLE' };
var FilterOption = (function () {
    function FilterOption(iKey, iDescription, iChecked) {
        this.iKey = iKey;
        this.iDescription = iDescription;
        this.iChecked = iChecked;
        this.key = iKey;
        this.description = iDescription;
        this.checked = iChecked;
    }
    FilterOption.prototype.toString = function () {
        return 'key:' + this.key + ', description:' + this.description + ', checked:' + this.checked + '\n';
    };
    return FilterOption;
}());
var AuditLogComponent = (function () {
    function AuditLogComponent(route, router, auditLogService, messageService) {
        var _this = this;
        this.route = route;
        this.router = router;
        this.auditLogService = auditLogService;
        this.messageService = messageService;
        this.queryParam = new audit_log_1.AuditLog();
        this.toggleName = optionalSearch;
        this.currentOption = 0;
        this.filterOptions = [
            new FilterOption('all', 'AUDIT_LOG.ALL_OPERATIONS', true),
            new FilterOption('pull', 'AUDIT_LOG.PULL', true),
            new FilterOption('push', 'AUDIT_LOG.PUSH', true),
            new FilterOption('create', 'AUDIT_LOG.CREATE', true),
            new FilterOption('delete', 'AUDIT_LOG.DELETE', true),
            new FilterOption('others', 'AUDIT_LOG.OTHERS', true)
        ];
        this.pageOffset = 1;
        this.pageSize = 2;
        //Get current user from registered resolver.
        this.route.data.subscribe(function (data) { return _this.currentUser = data['auditLogResolver']; });
    }
    AuditLogComponent.prototype.ngOnInit = function () {
        this.projectId = +this.route.snapshot.parent.params['id'];
        console.log('Get projectId from route params snapshot:' + this.projectId);
        this.queryParam.project_id = this.projectId;
        this.queryParam.page_size = this.pageSize;
    };
    AuditLogComponent.prototype.retrieve = function (state) {
        var _this = this;
        if (state) {
            this.queryParam.page = state.page.to + 1;
        }
        this.auditLogService
            .listAuditLogs(this.queryParam)
            .subscribe(function (response) {
            _this.totalRecordCount = response.headers.get('x-total-count');
            _this.totalPage = Math.ceil(_this.totalRecordCount / _this.pageSize);
            console.log('TotalRecordCount:' + _this.totalRecordCount + ', totalPage:' + _this.totalPage);
            _this.auditLogs = response.json();
        }, function (error) {
            _this.router.navigate(['/harbor', 'projects']);
            _this.messageService.announceMessage(error.status, 'Failed to list audit logs with project ID:' + _this.queryParam.project_id, shared_const_1.AlertType.DANGER);
        });
    };
    AuditLogComponent.prototype.doSearchAuditLogs = function (searchUsername) {
        this.queryParam.username = searchUsername;
        this.retrieve();
    };
    AuditLogComponent.prototype.doSearchByTimeRange = function (strDate, target) {
        var oneDayOffset = 3600 * 24;
        switch (target) {
            case 'begin':
                this.queryParam.begin_timestamp = new Date(strDate).getTime() / 1000;
                break;
            case 'end':
                this.queryParam.end_timestamp = new Date(strDate).getTime() / 1000 + oneDayOffset;
                break;
        }
        console.log('Search audit log filtered by time range, begin: ' + this.queryParam.begin_timestamp + ', end:' + this.queryParam.end_timestamp);
        this.retrieve();
    };
    AuditLogComponent.prototype.doSearchByOptions = function () {
        var selectAll = true;
        var operationFilter = [];
        for (var i in this.filterOptions) {
            var filterOption = this.filterOptions[i];
            if (filterOption.checked) {
                operationFilter.push(this.filterOptions[i].key);
            }
            else {
                selectAll = false;
            }
        }
        if (selectAll) {
            operationFilter = [];
        }
        this.queryParam.keywords = operationFilter.join('/');
        this.retrieve();
        console.log('Search option filter:' + operationFilter.join('/'));
    };
    AuditLogComponent.prototype.toggleOptionalName = function (option) {
        (option === 1) ? this.currentOption = 0 : this.currentOption = 1;
    };
    AuditLogComponent.prototype.toggleFilterOption = function (option) {
        var selectedOption = this.filterOptions.find(function (value) { return (value.key === option); });
        selectedOption.checked = !selectedOption.checked;
        if (selectedOption.key === 'all') {
            this.filterOptions.filter(function (value) { return value.key !== selectedOption.key; }).forEach(function (value) { return value.checked = selectedOption.checked; });
        }
        else {
            if (!selectedOption.checked) {
                this.filterOptions.find(function (value) { return value.key === 'all'; }).checked = false;
            }
            var selectAll_1 = true;
            this.filterOptions.filter(function (value) { return value.key !== 'all'; }).forEach(function (value) {
                if (!value.checked) {
                    selectAll_1 = false;
                }
            });
            this.filterOptions.find(function (value) { return value.key === 'all'; }).checked = selectAll_1;
        }
        this.doSearchByOptions();
    };
    AuditLogComponent.prototype.refresh = function () {
        this.retrieve();
    };
    AuditLogComponent = __decorate([
        core_1.Component({
            selector: 'audit-log',
            template: __webpack_require__(837),
            styles: [__webpack_require__(807)]
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof router_1.ActivatedRoute !== 'undefined' && router_1.ActivatedRoute) === 'function' && _a) || Object, (typeof (_b = typeof router_1.Router !== 'undefined' && router_1.Router) === 'function' && _b) || Object, (typeof (_c = typeof audit_log_service_1.AuditLogService !== 'undefined' && audit_log_service_1.AuditLogService) === 'function' && _c) || Object, (typeof (_d = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _d) || Object])
    ], AuditLogComponent);
    return AuditLogComponent;
    var _a, _b, _c, _d;
}());
exports.AuditLogComponent = AuditLogComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/audit-log.component.js.map

/***/ }),

/***/ 391:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var audit_log_service_1 = __webpack_require__(247);
var session_service_1 = __webpack_require__(14);
var message_service_1 = __webpack_require__(10);
var shared_const_1 = __webpack_require__(2);
var shared_utils_1 = __webpack_require__(33);
var RecentLogComponent = (function () {
    function RecentLogComponent(session, msgService, logService) {
        this.session = session;
        this.msgService = msgService;
        this.logService = logService;
        this.sessionUser = null;
        this.onGoing = false;
        this.lines = 10; //Support 10, 25 and 50
        this.sessionUser = this.session.getCurrentUser(); //Initialize session
    }
    RecentLogComponent.prototype.ngOnInit = function () {
        this.retrieveLogs();
    };
    Object.defineProperty(RecentLogComponent.prototype, "inProgress", {
        get: function () {
            return this.onGoing;
        },
        enumerable: true,
        configurable: true
    });
    RecentLogComponent.prototype.setLines = function (lines) {
        this.lines = lines;
        if (this.lines < 10) {
            this.lines = 10;
        }
        this.retrieveLogs();
    };
    RecentLogComponent.prototype.doFilter = function (terms) {
        var _this = this;
        if (terms.trim() === "") {
            this.recentLogs = this.logsCache.filter(function (log) { return log.username != ""; });
            return;
        }
        this.recentLogs = this.logsCache.filter(function (log) { return _this.isMatched(terms, log); });
    };
    RecentLogComponent.prototype.refresh = function () {
        this.retrieveLogs();
    };
    RecentLogComponent.prototype.formatDateTime = function (dateTime) {
        var dt = new Date(dateTime);
        return dt.toLocaleString();
    };
    RecentLogComponent.prototype.retrieveLogs = function () {
        var _this = this;
        if (this.lines < 10) {
            this.lines = 10;
        }
        this.onGoing = true;
        this.logService.getRecentLogs(this.lines)
            .subscribe(function (response) {
            _this.onGoing = false;
            _this.logsCache = response; //Keep the data
            _this.recentLogs = _this.logsCache.filter(function (log) { return log.username != ""; }); //To display
        }, function (error) {
            _this.onGoing = false;
            if (!shared_utils_1.accessErrorHandler(error, _this.msgService)) {
                _this.msgService.announceMessage(error.status, shared_utils_1.errorHandler(error), shared_const_1.AlertType.DANGER);
            }
        });
    };
    RecentLogComponent.prototype.isMatched = function (terms, log) {
        var reg = new RegExp('.*' + terms + '.*', 'i');
        return reg.test(log.username) ||
            reg.test(log.repo_name) ||
            reg.test(log.operation);
    };
    RecentLogComponent = __decorate([
        core_1.Component({
            selector: 'recent-log',
            template: __webpack_require__(838),
            styles: [__webpack_require__(808)]
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _a) || Object, (typeof (_b = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _b) || Object, (typeof (_c = typeof audit_log_service_1.AuditLogService !== 'undefined' && audit_log_service_1.AuditLogService) === 'function' && _c) || Object])
    ], RecentLogComponent);
    return RecentLogComponent;
    var _a, _b, _c;
}());
exports.RecentLogComponent = RecentLogComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/recent-log.component.js.map

/***/ }),

/***/ 392:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var http_1 = __webpack_require__(20);
var project_1 = __webpack_require__(615);
var project_service_1 = __webpack_require__(174);
var message_service_1 = __webpack_require__(10);
var shared_const_1 = __webpack_require__(2);
var core_2 = __webpack_require__(34);
var CreateProjectComponent = (function () {
    function CreateProjectComponent(projectService, messageService, translateService) {
        this.projectService = projectService;
        this.messageService = messageService;
        this.translateService = translateService;
        this.project = new project_1.Project();
        this.create = new core_1.EventEmitter();
    }
    CreateProjectComponent.prototype.onSubmit = function () {
        var _this = this;
        this.projectService
            .createProject(this.project.name, this.project.public ? 1 : 0)
            .subscribe(function (status) {
            _this.create.emit(true);
            _this.createProjectOpened = false;
        }, function (error) {
            _this.errorMessageOpened = true;
            if (error instanceof http_1.Response) {
                switch (error.status) {
                    case 409:
                        _this.translateService.get('PROJECT.NAME_ALREADY_EXISTS').subscribe(function (res) { return _this.errorMessage = res; });
                        break;
                    case 400:
                        _this.translateService.get('PROJECT.NAME_IS_ILLEGAL').subscribe(function (res) { return _this.errorMessage = res; });
                        break;
                    default:
                        _this.translateService.get('PROJECT.UNKNOWN_ERROR').subscribe(function (res) {
                            _this.errorMessage = res;
                            _this.messageService.announceMessage(error.status, _this.errorMessage, shared_const_1.AlertType.DANGER);
                        });
                }
            }
        });
    };
    CreateProjectComponent.prototype.newProject = function () {
        this.project = new project_1.Project();
        this.createProjectOpened = true;
        this.errorMessageOpened = false;
        this.errorMessage = '';
    };
    CreateProjectComponent.prototype.onErrorMessageClose = function () {
        this.errorMessageOpened = false;
        this.errorMessage = '';
    };
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], CreateProjectComponent.prototype, "create", void 0);
    CreateProjectComponent = __decorate([
        core_1.Component({
            selector: 'create-project',
            template: __webpack_require__(839),
            styles: [__webpack_require__(809)]
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof project_service_1.ProjectService !== 'undefined' && project_service_1.ProjectService) === 'function' && _a) || Object, (typeof (_b = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _b) || Object, (typeof (_c = typeof core_2.TranslateService !== 'undefined' && core_2.TranslateService) === 'function' && _c) || Object])
    ], CreateProjectComponent);
    return CreateProjectComponent;
    var _a, _b, _c;
}());
exports.CreateProjectComponent = CreateProjectComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/create-project.component.js.map

/***/ }),

/***/ 393:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var session_service_1 = __webpack_require__(14);
var search_trigger_service_1 = __webpack_require__(95);
var shared_const_1 = __webpack_require__(2);
var ListProjectComponent = (function () {
    function ListProjectComponent(session, router, searchTrigger) {
        this.session = session;
        this.router = router;
        this.searchTrigger = searchTrigger;
        this.pageOffset = 1;
        this.paginate = new core_1.EventEmitter();
        this.toggle = new core_1.EventEmitter();
        this.delete = new core_1.EventEmitter();
        this.mode = shared_const_1.ListMode.FULL;
    }
    ListProjectComponent.prototype.ngOnInit = function () {
    };
    Object.defineProperty(ListProjectComponent.prototype, "listFullMode", {
        get: function () {
            return this.mode === shared_const_1.ListMode.FULL;
        },
        enumerable: true,
        configurable: true
    });
    ListProjectComponent.prototype.goToLink = function (proId) {
        this.searchTrigger.closeSearch(false);
        var linkUrl = ['harbor', 'projects', proId, 'repository'];
        if (!this.session.getCurrentUser()) {
            var navigatorExtra = {
                queryParams: { "redirect_url": linkUrl.join("/") }
            };
            this.router.navigate([shared_const_1.signInRoute], navigatorExtra);
        }
        else {
            this.router.navigate(linkUrl);
        }
    };
    ListProjectComponent.prototype.refresh = function (state) {
        this.paginate.emit(state);
    };
    ListProjectComponent.prototype.toggleProject = function (p) {
        this.toggle.emit(p);
    };
    ListProjectComponent.prototype.deleteProject = function (p) {
        this.delete.emit(p);
    };
    __decorate([
        core_1.Input(), 
        __metadata('design:type', Array)
    ], ListProjectComponent.prototype, "projects", void 0);
    __decorate([
        core_1.Input(), 
        __metadata('design:type', Number)
    ], ListProjectComponent.prototype, "totalPage", void 0);
    __decorate([
        core_1.Input(), 
        __metadata('design:type', Number)
    ], ListProjectComponent.prototype, "totalRecordCount", void 0);
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], ListProjectComponent.prototype, "paginate", void 0);
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], ListProjectComponent.prototype, "toggle", void 0);
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], ListProjectComponent.prototype, "delete", void 0);
    __decorate([
        core_1.Input(), 
        __metadata('design:type', String)
    ], ListProjectComponent.prototype, "mode", void 0);
    ListProjectComponent = __decorate([
        core_1.Component({
            selector: 'list-project',
            template: __webpack_require__(840)
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _a) || Object, (typeof (_b = typeof router_1.Router !== 'undefined' && router_1.Router) === 'function' && _b) || Object, (typeof (_c = typeof search_trigger_service_1.SearchTriggerService !== 'undefined' && search_trigger_service_1.SearchTriggerService) === 'function' && _c) || Object])
    ], ListProjectComponent);
    return ListProjectComponent;
    var _a, _b, _c;
}());
exports.ListProjectComponent = ListProjectComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/list-project.component.js.map

/***/ }),

/***/ 394:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var http_1 = __webpack_require__(20);
var member_service_1 = __webpack_require__(248);
var message_service_1 = __webpack_require__(10);
var shared_const_1 = __webpack_require__(2);
var core_2 = __webpack_require__(34);
var member_1 = __webpack_require__(613);
var AddMemberComponent = (function () {
    function AddMemberComponent(memberService, messageService, translateService) {
        this.memberService = memberService;
        this.messageService = messageService;
        this.translateService = translateService;
        this.member = new member_1.Member();
        this.added = new core_1.EventEmitter();
    }
    AddMemberComponent.prototype.onSubmit = function () {
        var _this = this;
        console.log('Adding member:' + JSON.stringify(this.member));
        this.memberService
            .addMember(this.projectId, this.member.username, this.member.role_id)
            .subscribe(function (response) {
            console.log('Added member successfully.');
            _this.added.emit(true);
            _this.addMemberOpened = false;
        }, function (error) {
            _this.errorMessageOpened = true;
            if (error instanceof http_1.Response) {
                switch (error.status) {
                    case 404:
                        _this.translateService.get('MEMBER.USERNAME_DOES_NOT_EXISTS').subscribe(function (res) { return _this.errorMessage = res; });
                        break;
                    case 409:
                        _this.translateService.get('MEMBER.USERNAME_ALREADY_EXISTS').subscribe(function (res) { return _this.errorMessage = res; });
                        break;
                    default:
                        _this.translateService.get('MEMBER.UNKNOWN_ERROR').subscribe(function (res) {
                            _this.errorMessage = res;
                            _this.messageService.announceMessage(error.status, _this.errorMessage, shared_const_1.AlertType.DANGER);
                        });
                }
            }
            console.log('Failed to add member of project:' + _this.projectId, ' with error:' + error);
        });
    };
    AddMemberComponent.prototype.openAddMemberModal = function () {
        this.errorMessageOpened = false;
        this.errorMessage = '';
        this.member = new member_1.Member();
        this.addMemberOpened = true;
    };
    AddMemberComponent.prototype.onErrorMessageClose = function () {
        this.errorMessageOpened = false;
        this.errorMessage = '';
    };
    __decorate([
        core_1.Input(), 
        __metadata('design:type', Number)
    ], AddMemberComponent.prototype, "projectId", void 0);
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], AddMemberComponent.prototype, "added", void 0);
    AddMemberComponent = __decorate([
        core_1.Component({
            selector: 'add-member',
            template: __webpack_require__(841)
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof member_service_1.MemberService !== 'undefined' && member_service_1.MemberService) === 'function' && _a) || Object, (typeof (_b = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _b) || Object, (typeof (_c = typeof core_2.TranslateService !== 'undefined' && core_2.TranslateService) === 'function' && _c) || Object])
    ], AddMemberComponent);
    return AddMemberComponent;
    var _a, _b, _c;
}());
exports.AddMemberComponent = AddMemberComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/add-member.component.js.map

/***/ }),

/***/ 395:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var member_service_1 = __webpack_require__(248);
var add_member_component_1 = __webpack_require__(394);
var message_service_1 = __webpack_require__(10);
var shared_const_1 = __webpack_require__(2);
var deletion_dialog_service_1 = __webpack_require__(48);
var deletion_message_1 = __webpack_require__(68);
var session_service_1 = __webpack_require__(14);
__webpack_require__(874);
__webpack_require__(133);
__webpack_require__(85);
__webpack_require__(132);
exports.roleInfo = { 1: 'MEMBER.PROJECT_ADMIN', 2: 'MEMBER.DEVELOPER', 3: 'MEMBER.GUEST' };
var MemberComponent = (function () {
    function MemberComponent(route, router, memberService, messageService, deletionDialogService, session) {
        var _this = this;
        this.route = route;
        this.router = router;
        this.memberService = memberService;
        this.messageService = messageService;
        this.deletionDialogService = deletionDialogService;
        this.roleInfo = exports.roleInfo;
        //Get current user from registered resolver.
        this.currentUser = session.getCurrentUser();
        deletionDialogService.deletionConfirm$.subscribe(function (message) {
            if (message && message.targetId === shared_const_1.DeletionTargets.PROJECT_MEMBER) {
                _this.memberService
                    .deleteMember(_this.projectId, message.data)
                    .subscribe(function (response) {
                    console.log('Successful change role with user ' + message.data);
                    _this.retrieve(_this.projectId, '');
                }, function (error) { return _this.messageService.announceMessage(error.status, 'Failed to change role with user ' + message.data, shared_const_1.AlertType.DANGER); });
            }
        });
    }
    MemberComponent.prototype.retrieve = function (projectId, username) {
        var _this = this;
        this.memberService
            .listMembers(projectId, username)
            .subscribe(function (response) { return _this.members = response; }, function (error) {
            _this.router.navigate(['/harbor', 'projects']);
            _this.messageService.announceMessage(error.status, 'Failed to get project member with project ID:' + projectId, shared_const_1.AlertType.DANGER);
        });
    };
    MemberComponent.prototype.ngOnInit = function () {
        //Get projectId from route params snapshot.          
        this.projectId = +this.route.snapshot.parent.params['id'];
        console.log('Get projectId from route params snapshot:' + this.projectId);
        this.retrieve(this.projectId, '');
    };
    MemberComponent.prototype.openAddMemberModal = function () {
        this.addMemberComponent.openAddMemberModal();
    };
    MemberComponent.prototype.addedMember = function () {
        this.retrieve(this.projectId, '');
    };
    MemberComponent.prototype.changeRole = function (userId, roleId) {
        var _this = this;
        this.memberService
            .changeMemberRole(this.projectId, userId, roleId)
            .subscribe(function (response) {
            console.log('Successful change role with user ' + userId + ' to roleId ' + roleId);
            _this.retrieve(_this.projectId, '');
        }, function (error) { return _this.messageService.announceMessage(error.status, 'Failed to change role with user ' + userId + ' to roleId ' + roleId, shared_const_1.AlertType.DANGER); });
    };
    MemberComponent.prototype.deleteMember = function (userId) {
        var deletionMessage = new deletion_message_1.DeletionMessage('MEMBER.DELETION_TITLE', 'MEMBER.DELETION_SUMMARY', userId + "", userId, shared_const_1.DeletionTargets.PROJECT_MEMBER);
        this.deletionDialogService.openComfirmDialog(deletionMessage);
    };
    MemberComponent.prototype.doSearch = function (searchMember) {
        this.retrieve(this.projectId, searchMember);
    };
    MemberComponent.prototype.refresh = function () {
        this.retrieve(this.projectId, '');
    };
    __decorate([
        core_1.ViewChild(add_member_component_1.AddMemberComponent), 
        __metadata('design:type', (typeof (_a = typeof add_member_component_1.AddMemberComponent !== 'undefined' && add_member_component_1.AddMemberComponent) === 'function' && _a) || Object)
    ], MemberComponent.prototype, "addMemberComponent", void 0);
    MemberComponent = __decorate([
        core_1.Component({
            template: __webpack_require__(842)
        }), 
        __metadata('design:paramtypes', [(typeof (_b = typeof router_1.ActivatedRoute !== 'undefined' && router_1.ActivatedRoute) === 'function' && _b) || Object, (typeof (_c = typeof router_1.Router !== 'undefined' && router_1.Router) === 'function' && _c) || Object, (typeof (_d = typeof member_service_1.MemberService !== 'undefined' && member_service_1.MemberService) === 'function' && _d) || Object, (typeof (_e = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _e) || Object, (typeof (_f = typeof deletion_dialog_service_1.DeletionDialogService !== 'undefined' && deletion_dialog_service_1.DeletionDialogService) === 'function' && _f) || Object, (typeof (_g = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _g) || Object])
    ], MemberComponent);
    return MemberComponent;
    var _a, _b, _c, _d, _e, _f, _g;
}());
exports.MemberComponent = MemberComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/member.component.js.map

/***/ }),

/***/ 396:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var session_service_1 = __webpack_require__(14);
var ProjectDetailComponent = (function () {
    function ProjectDetailComponent(route, router, sessionService) {
        var _this = this;
        this.route = route;
        this.router = router;
        this.sessionService = sessionService;
        this.route.data.subscribe(function (data) { return _this.currentProject = data['projectResolver']; });
    }
    Object.defineProperty(ProjectDetailComponent.prototype, "isSystemAdmin", {
        get: function () {
            var account = this.sessionService.getCurrentUser();
            return account != null && account.has_admin_role > 0;
        },
        enumerable: true,
        configurable: true
    });
    ProjectDetailComponent = __decorate([
        core_1.Component({
            selector: 'project-detail',
            template: __webpack_require__(843),
            styles: [__webpack_require__(810)]
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof router_1.ActivatedRoute !== 'undefined' && router_1.ActivatedRoute) === 'function' && _a) || Object, (typeof (_b = typeof router_1.Router !== 'undefined' && router_1.Router) === 'function' && _b) || Object, (typeof (_c = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _c) || Object])
    ], ProjectDetailComponent);
    return ProjectDetailComponent;
    var _a, _b, _c;
}());
exports.ProjectDetailComponent = ProjectDetailComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/project-detail.component.js.map

/***/ }),

/***/ 397:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var project_service_1 = __webpack_require__(174);
var ProjectRoutingResolver = (function () {
    function ProjectRoutingResolver(projectService, router) {
        this.projectService = projectService;
        this.router = router;
    }
    ProjectRoutingResolver.prototype.resolve = function (route, state) {
        var _this = this;
        var projectId = route.params['id'];
        return this.projectService
            .getProject(projectId)
            .then(function (project) {
            if (project) {
                return project;
            }
            else {
                _this.router.navigate(['/harbor', 'projects']);
                return null;
            }
        });
    };
    ProjectRoutingResolver = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof project_service_1.ProjectService !== 'undefined' && project_service_1.ProjectService) === 'function' && _a) || Object, (typeof (_b = typeof router_1.Router !== 'undefined' && router_1.Router) === 'function' && _b) || Object])
    ], ProjectRoutingResolver);
    return ProjectRoutingResolver;
    var _a, _b;
}());
exports.ProjectRoutingResolver = ProjectRoutingResolver;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/project-routing-resolver.service.js.map

/***/ }),

/***/ 398:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var project_service_1 = __webpack_require__(174);
var create_project_component_1 = __webpack_require__(392);
var list_project_component_1 = __webpack_require__(393);
var message_service_1 = __webpack_require__(10);
var shared_const_1 = __webpack_require__(2);
var deletion_dialog_service_1 = __webpack_require__(48);
var deletion_message_1 = __webpack_require__(68);
var shared_const_2 = __webpack_require__(2);
var types = { 0: 'PROJECT.MY_PROJECTS', 1: 'PROJECT.PUBLIC_PROJECTS' };
var ProjectComponent = (function () {
    function ProjectComponent(projectService, messageService, deletionDialogService) {
        var _this = this;
        this.projectService = projectService;
        this.messageService = messageService;
        this.deletionDialogService = deletionDialogService;
        this.selected = [];
        this.projectTypes = types;
        this.currentFilteredType = 0;
        this.page = 1;
        this.pageSize = 3;
        this.subscription = deletionDialogService.deletionConfirm$.subscribe(function (message) {
            if (message && message.targetId === shared_const_2.DeletionTargets.PROJECT) {
                var projectId_1 = message.data;
                _this.projectService
                    .deleteProject(projectId_1)
                    .subscribe(function (response) {
                    console.log('Successful delete project with ID:' + projectId_1);
                    _this.retrieve();
                }, function (error) { return _this.messageService.announceMessage(error.status, error, shared_const_1.AlertType.WARNING); });
            }
        });
    }
    ProjectComponent.prototype.ngOnInit = function () {
        this.projectName = '';
        this.isPublic = 0;
    };
    ProjectComponent.prototype.retrieve = function (state) {
        var _this = this;
        if (state) {
            this.page = state.page.to + 1;
        }
        this.projectService
            .listProjects(this.projectName, this.isPublic, this.page, this.pageSize)
            .subscribe(function (response) {
            _this.totalRecordCount = response.headers.get('x-total-count');
            _this.totalPage = Math.ceil(_this.totalRecordCount / _this.pageSize);
            console.log('TotalRecordCount:' + _this.totalRecordCount + ', totalPage:' + _this.totalPage);
            _this.changedProjects = response.json();
        }, function (error) { return _this.messageService.announceAppLevelMessage(error.status, error, shared_const_1.AlertType.WARNING); });
    };
    ProjectComponent.prototype.openModal = function () {
        this.creationProject.newProject();
    };
    ProjectComponent.prototype.createProject = function (created) {
        if (created) {
            this.retrieve();
        }
    };
    ProjectComponent.prototype.doSearchProjects = function (projectName) {
        console.log('Search for project name:' + projectName);
        this.projectName = projectName;
        this.retrieve();
    };
    ProjectComponent.prototype.doFilterProjects = function (filteredType) {
        console.log('Filter projects with type:' + types[filteredType]);
        this.isPublic = filteredType;
        this.retrieve();
    };
    ProjectComponent.prototype.toggleProject = function (p) {
        var _this = this;
        if (p) {
            p.public === 0 ? p.public = 1 : p.public = 0;
            this.projectService
                .toggleProjectPublic(p.project_id, p.public)
                .subscribe(function (response) { return console.log('Successful toggled project_id:' + p.project_id); }, function (error) { return _this.messageService.announceMessage(error.status, error, shared_const_1.AlertType.WARNING); });
        }
    };
    ProjectComponent.prototype.deleteProject = function (p) {
        var deletionMessage = new deletion_message_1.DeletionMessage('PROJECT.DELETION_TITLE', 'PROJECT.DELETION_SUMMARY', p.name, p.project_id, shared_const_2.DeletionTargets.PROJECT);
        this.deletionDialogService.openComfirmDialog(deletionMessage);
    };
    ProjectComponent.prototype.refresh = function () {
        this.retrieve();
    };
    __decorate([
        core_1.ViewChild(create_project_component_1.CreateProjectComponent), 
        __metadata('design:type', (typeof (_a = typeof create_project_component_1.CreateProjectComponent !== 'undefined' && create_project_component_1.CreateProjectComponent) === 'function' && _a) || Object)
    ], ProjectComponent.prototype, "creationProject", void 0);
    __decorate([
        core_1.ViewChild(list_project_component_1.ListProjectComponent), 
        __metadata('design:type', (typeof (_b = typeof list_project_component_1.ListProjectComponent !== 'undefined' && list_project_component_1.ListProjectComponent) === 'function' && _b) || Object)
    ], ProjectComponent.prototype, "listProject", void 0);
    ProjectComponent = __decorate([
        core_1.Component({
            selector: 'project',
            template: __webpack_require__(844),
            styles: [__webpack_require__(811)]
        }), 
        __metadata('design:paramtypes', [(typeof (_c = typeof project_service_1.ProjectService !== 'undefined' && project_service_1.ProjectService) === 'function' && _c) || Object, (typeof (_d = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _d) || Object, (typeof (_e = typeof deletion_dialog_service_1.DeletionDialogService !== 'undefined' && deletion_dialog_service_1.DeletionDialogService) === 'function' && _e) || Object])
    ], ProjectComponent);
    return ProjectComponent;
    var _a, _b, _c, _d, _e;
}());
exports.ProjectComponent = ProjectComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/project.component.js.map

/***/ }),

/***/ 399:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var replication_service_1 = __webpack_require__(79);
var message_service_1 = __webpack_require__(10);
var shared_const_1 = __webpack_require__(2);
var target_1 = __webpack_require__(249);
var core_2 = __webpack_require__(34);
var CreateEditDestinationComponent = (function () {
    function CreateEditDestinationComponent(replicationService, messageService, translateService) {
        this.replicationService = replicationService;
        this.messageService = messageService;
        this.translateService = translateService;
        this.target = new target_1.Target();
        this.reload = new core_1.EventEmitter();
    }
    CreateEditDestinationComponent.prototype.openCreateEditTarget = function (targetId) {
        var _this = this;
        this.target = new target_1.Target();
        this.createEditDestinationOpened = true;
        this.errorMessageOpened = false;
        this.errorMessage = '';
        this.pingTestMessage = '';
        this.pingStatus = true;
        this.testOngoing = false;
        if (targetId) {
            this.actionType = shared_const_1.ActionType.EDIT;
            this.translateService.get('DESTINATION.TITLE_EDIT').subscribe(function (res) { return _this.modalTitle = res; });
            this.replicationService
                .getTarget(targetId)
                .subscribe(function (target) { return _this.target = target; }, function (error) { return _this.messageService
                .announceMessage(error.status, 'DESTINATION.FAILED_TO_GET_TARGET', shared_const_1.AlertType.DANGER); });
        }
        else {
            this.actionType = shared_const_1.ActionType.ADD_NEW;
            this.translateService.get('DESTINATION.TITLE_ADD').subscribe(function (res) { return _this.modalTitle = res; });
        }
    };
    CreateEditDestinationComponent.prototype.testConnection = function () {
        var _this = this;
        this.translateService.get('DESTINATION.TESTING_CONNECTION').subscribe(function (res) { return _this.pingTestMessage = res; });
        this.pingStatus = true;
        this.testOngoing = !this.testOngoing;
        this.replicationService
            .pingTarget(this.target)
            .subscribe(function (response) {
            _this.pingStatus = true;
            _this.translateService.get('DESTINATION.TEST_CONNECTION_SUCCESS').subscribe(function (res) { return _this.pingTestMessage = res; });
            _this.testOngoing = !_this.testOngoing;
        }, function (error) {
            _this.pingStatus = false;
            _this.translateService.get('DESTINATION.TEST_CONNECTION_FAILURE').subscribe(function (res) { return _this.pingTestMessage = res; });
            _this.testOngoing = !_this.testOngoing;
        });
    };
    CreateEditDestinationComponent.prototype.onSubmit = function () {
        var _this = this;
        this.errorMessage = '';
        this.errorMessageOpened = false;
        switch (this.actionType) {
            case shared_const_1.ActionType.ADD_NEW:
                this.replicationService
                    .createTarget(this.target)
                    .subscribe(function (response) {
                    console.log('Successful added target.');
                    _this.createEditDestinationOpened = false;
                    _this.reload.emit(true);
                }, function (error) {
                    _this.errorMessageOpened = true;
                    var errorMessageKey = '';
                    switch (error.status) {
                        case 409:
                            errorMessageKey = 'DESTINATION.CONFLICT_NAME';
                            break;
                        case 400:
                            errorMessageKey = 'DESTINATION.INVALID_NAME';
                            break;
                        default:
                            errorMessageKey = 'UNKNOWN_ERROR';
                    }
                    _this.translateService
                        .get(errorMessageKey)
                        .subscribe(function (res) {
                        _this.errorMessage = res;
                        _this.messageService.announceMessage(error.status, errorMessageKey, shared_const_1.AlertType.DANGER);
                    });
                });
                break;
            case shared_const_1.ActionType.EDIT:
                this.replicationService
                    .updateTarget(this.target)
                    .subscribe(function (response) {
                    console.log('Successful updated target.');
                    _this.createEditDestinationOpened = false;
                    _this.reload.emit(true);
                }, function (error) {
                    _this.errorMessageOpened = true;
                    _this.errorMessage = 'Failed to update target:' + error;
                    var errorMessageKey = '';
                    switch (error.status) {
                        case 409:
                            errorMessageKey = 'DESTINATION.CONFLICT_NAME';
                            break;
                        case 400:
                            errorMessageKey = 'DESTINATION.INVALID_NAME';
                            break;
                        default:
                            errorMessageKey = 'UNKNOWN_ERROR';
                    }
                    _this.translateService
                        .get(errorMessageKey)
                        .subscribe(function (res) {
                        _this.errorMessage = res;
                        _this.messageService.announceMessage(error.status, errorMessageKey, shared_const_1.AlertType.DANGER);
                    });
                });
                break;
        }
    };
    CreateEditDestinationComponent.prototype.onErrorMessageClose = function () {
        this.errorMessageOpened = false;
        this.errorMessage = '';
    };
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], CreateEditDestinationComponent.prototype, "reload", void 0);
    CreateEditDestinationComponent = __decorate([
        core_1.Component({
            selector: 'create-edit-destination',
            template: __webpack_require__(845)
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof replication_service_1.ReplicationService !== 'undefined' && replication_service_1.ReplicationService) === 'function' && _a) || Object, (typeof (_b = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _b) || Object, (typeof (_c = typeof core_2.TranslateService !== 'undefined' && core_2.TranslateService) === 'function' && _c) || Object])
    ], CreateEditDestinationComponent);
    return CreateEditDestinationComponent;
    var _a, _b, _c;
}());
exports.CreateEditDestinationComponent = CreateEditDestinationComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/create-edit-destination.component.js.map

/***/ }),

/***/ 400:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var target_1 = __webpack_require__(249);
var replication_service_1 = __webpack_require__(79);
var message_service_1 = __webpack_require__(10);
var shared_const_1 = __webpack_require__(2);
var deletion_dialog_service_1 = __webpack_require__(48);
var deletion_message_1 = __webpack_require__(68);
var shared_const_2 = __webpack_require__(2);
var create_edit_destination_component_1 = __webpack_require__(399);
var DestinationComponent = (function () {
    function DestinationComponent(replicationService, messageService, deletionDialogService) {
        var _this = this;
        this.replicationService = replicationService;
        this.messageService = messageService;
        this.deletionDialogService = deletionDialogService;
        this.subscription = this.deletionDialogService.deletionConfirm$.subscribe(function (message) {
            var targetId = message.data;
            _this.replicationService
                .deleteTarget(targetId)
                .subscribe(function (response) {
                console.log('Successful deleted target with ID:' + targetId);
                _this.reload();
            }, function (error) { return _this.messageService
                .announceMessage(error.status, 'Failed to delete target with ID:' + targetId + ', error:' + error, shared_const_1.AlertType.DANGER); });
        });
    }
    DestinationComponent.prototype.ngOnInit = function () {
        this.targetName = '';
        this.retrieve('');
    };
    DestinationComponent.prototype.ngOnDestroy = function () {
        if (this.subscription) {
            this.subscription.unsubscribe();
        }
    };
    DestinationComponent.prototype.retrieve = function (targetName) {
        var _this = this;
        this.replicationService
            .listTargets(targetName)
            .subscribe(function (targets) { return _this.targets = targets; }, function (error) { return _this.messageService.announceMessage(error.status, 'Failed to get targets:' + error, shared_const_1.AlertType.DANGER); });
    };
    DestinationComponent.prototype.doSearchTargets = function (targetName) {
        this.targetName = targetName;
        this.retrieve(targetName);
    };
    DestinationComponent.prototype.refreshTargets = function () {
        this.retrieve('');
    };
    DestinationComponent.prototype.reload = function () {
        this.retrieve(this.targetName);
    };
    DestinationComponent.prototype.openModal = function () {
        this.createEditDestinationComponent.openCreateEditTarget();
        this.target = new target_1.Target();
    };
    DestinationComponent.prototype.editTarget = function (target) {
        if (target) {
            this.createEditDestinationComponent.openCreateEditTarget(target.id);
        }
    };
    DestinationComponent.prototype.deleteTarget = function (target) {
        if (target) {
            var targetId = target.id;
            var deletionMessage = new deletion_message_1.DeletionMessage('REPLICATION.DELETION_TITLE_TARGET', 'REPLICATION.DELETION_SUMMARY_TARGET', target.name, target.id, shared_const_2.DeletionTargets.TARGET);
            this.deletionDialogService.openComfirmDialog(deletionMessage);
        }
    };
    __decorate([
        core_1.ViewChild(create_edit_destination_component_1.CreateEditDestinationComponent), 
        __metadata('design:type', (typeof (_a = typeof create_edit_destination_component_1.CreateEditDestinationComponent !== 'undefined' && create_edit_destination_component_1.CreateEditDestinationComponent) === 'function' && _a) || Object)
    ], DestinationComponent.prototype, "createEditDestinationComponent", void 0);
    DestinationComponent = __decorate([
        core_1.Component({
            selector: 'destination',
            template: __webpack_require__(846)
        }), 
        __metadata('design:paramtypes', [(typeof (_b = typeof replication_service_1.ReplicationService !== 'undefined' && replication_service_1.ReplicationService) === 'function' && _b) || Object, (typeof (_c = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _c) || Object, (typeof (_d = typeof deletion_dialog_service_1.DeletionDialogService !== 'undefined' && deletion_dialog_service_1.DeletionDialogService) === 'function' && _d) || Object])
    ], DestinationComponent);
    return DestinationComponent;
    var _a, _b, _c, _d;
}());
exports.DestinationComponent = DestinationComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/destination.component.js.map

/***/ }),

/***/ 401:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var ReplicationManagementComponent = (function () {
    function ReplicationManagementComponent() {
    }
    ReplicationManagementComponent = __decorate([
        core_1.Component({
            selector: 'replication-management',
            template: __webpack_require__(848),
            styles: [__webpack_require__(812)]
        }), 
        __metadata('design:paramtypes', [])
    ], ReplicationManagementComponent);
    return ReplicationManagementComponent;
}());
exports.ReplicationManagementComponent = ReplicationManagementComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/replication-management.component.js.map

/***/ }),

/***/ 402:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var create_edit_policy_component_1 = __webpack_require__(252);
var message_service_1 = __webpack_require__(10);
var shared_const_1 = __webpack_require__(2);
var session_service_1 = __webpack_require__(14);
var replication_service_1 = __webpack_require__(79);
var ruleStatus = [
    { 'key': '', 'description': 'REPLICATION.ALL_STATUS' },
    { 'key': '1', 'description': 'REPLICATION.ENABLED' },
    { 'key': '0', 'description': 'REPLICATION.DISABLED' }
];
var jobStatus = [
    { 'key': '', 'description': 'REPLICATION.ALL' },
    { 'key': 'pending', 'description': 'REPLICATION.PENDING' },
    { 'key': 'running', 'description': 'REPLICATION.RUNNING' },
    { 'key': 'error', 'description': 'REPLICATION.ERROR' },
    { 'key': 'retrying', 'description': 'REPLICATION.RETRYING' },
    { 'key': 'stopped', 'description': 'REPLICATION.STOPPED' },
    { 'key': 'finished', 'description': 'REPLICATION.FINISHED' },
    { 'key': 'canceled', 'description': 'REPLICATION.CANCELED' }
];
var optionalSearch = { 0: 'REPLICATION.ADVANCED', 1: 'REPLICATION.SIMPLE' };
var SearchOption = (function () {
    function SearchOption() {
        this.policyName = '';
        this.repoName = '';
        this.status = '';
        this.startTime = '';
        this.endTime = '';
        this.page = 1;
        this.pageSize = 5;
    }
    return SearchOption;
}());
var ReplicationComponent = (function () {
    function ReplicationComponent(sessionService, messageService, replicationService, route) {
        this.sessionService = sessionService;
        this.messageService = messageService;
        this.replicationService = replicationService;
        this.route = route;
        this.ruleStatus = ruleStatus;
        this.jobStatus = jobStatus;
        this.toggleJobSearchOption = optionalSearch;
        this.currentUser = this.sessionService.getCurrentUser();
    }
    ReplicationComponent.prototype.ngOnInit = function () {
        this.projectId = +this.route.snapshot.parent.params['id'];
        console.log('Get projectId from route params snapshot:' + this.projectId);
        this.search = new SearchOption();
        this.currentRuleStatus = this.ruleStatus[0];
        this.currentJobStatus = this.jobStatus[0];
        this.currentJobSearchOption = 0;
        this.retrievePolicies();
    };
    ReplicationComponent.prototype.retrievePolicies = function () {
        var _this = this;
        this.replicationService
            .listPolicies(this.search.policyName, this.projectId)
            .subscribe(function (response) {
            _this.changedPolicies = response;
            if (_this.changedPolicies && _this.changedPolicies.length > 0) {
                _this.initSelectedId = _this.changedPolicies[0].id;
            }
            _this.policies = _this.changedPolicies;
            if (_this.changedPolicies && _this.changedPolicies.length > 0) {
                _this.search.policyId = _this.changedPolicies[0].id;
                _this.fetchPolicyJobs();
            }
            else {
                _this.changedJobs = [];
            }
        }, function (error) { return _this.messageService.announceMessage(error.status, 'Failed to get policies with project ID:' + _this.projectId, shared_const_1.AlertType.DANGER); });
    };
    ReplicationComponent.prototype.openModal = function () {
        console.log('Open modal to create policy.');
        this.createEditPolicyComponent.openCreateEditPolicy();
    };
    ReplicationComponent.prototype.openEditPolicy = function (policyId) {
        console.log('Open modal to edit policy ID:' + policyId);
        this.createEditPolicyComponent.openCreateEditPolicy(policyId);
    };
    ReplicationComponent.prototype.fetchPolicyJobs = function (state) {
        var _this = this;
        if (state) {
            this.search.page = state.page.to + 1;
        }
        console.log('Received policy ID ' + this.search.policyId + ' by clicked row.');
        this.replicationService
            .listJobs(this.search.policyId, this.search.status, this.search.repoName, this.search.startTime, this.search.endTime, this.search.page, this.search.pageSize)
            .subscribe(function (response) {
            _this.jobsTotalRecordCount = response.headers.get('x-total-count');
            _this.jobsTotalPage = Math.ceil(_this.jobsTotalRecordCount / _this.search.pageSize);
            _this.changedJobs = response.json();
            _this.jobs = _this.changedJobs;
        }, function (error) { return _this.messageService.announceMessage(error.status, 'Failed to fetch jobs with policy ID:' + _this.search.policyId, shared_const_1.AlertType.DANGER); });
    };
    ReplicationComponent.prototype.selectOne = function (policy) {
        if (policy) {
            this.search.policyId = policy.id;
            this.fetchPolicyJobs();
        }
    };
    ReplicationComponent.prototype.doSearchPolicies = function (policyName) {
        this.search.policyName = policyName;
        this.retrievePolicies();
    };
    ReplicationComponent.prototype.doFilterPolicyStatus = function (status) {
        var _this = this;
        console.log('Do filter policies with status:' + status);
        this.currentRuleStatus = this.ruleStatus.find(function (r) { return r.key === status; });
        if (status.trim() === '') {
            this.changedPolicies = this.policies;
        }
        else {
            this.changedPolicies = this.policies.filter(function (policy) { return policy.enabled === +_this.currentRuleStatus.key; });
        }
    };
    ReplicationComponent.prototype.doFilterJobStatus = function (status) {
        console.log('Do filter jobs with status:' + status);
        this.currentJobStatus = this.jobStatus.find(function (r) { return r.key === status; });
        if (status.trim() === '') {
            this.changedJobs = this.jobs;
        }
        else {
            this.changedJobs = this.jobs.filter(function (job) { return job.status === status; });
        }
    };
    ReplicationComponent.prototype.doSearchJobs = function (repoName) {
        this.search.repoName = repoName;
        this.fetchPolicyJobs();
    };
    ReplicationComponent.prototype.reloadPolicies = function (isReady) {
        if (isReady) {
            this.retrievePolicies();
        }
    };
    ReplicationComponent.prototype.refreshPolicies = function () {
        this.retrievePolicies();
    };
    ReplicationComponent.prototype.refreshJobs = function () {
        this.fetchPolicyJobs();
    };
    ReplicationComponent.prototype.toggleSearchJobOptionalName = function (option) {
        (option === 1) ? this.currentJobSearchOption = 0 : this.currentJobSearchOption = 1;
    };
    ReplicationComponent.prototype.doJobSearchByTimeRange = function (strDate, target) {
        if (!strDate || strDate.trim() === '') {
            strDate = 0 + '';
        }
        var oneDayOffset = 3600 * 24;
        switch (target) {
            case 'begin':
                this.search.startTime = (new Date(strDate).getTime() / 1000) + '';
                break;
            case 'end':
                this.search.endTime = (new Date(strDate).getTime() / 1000 + oneDayOffset) + '';
                break;
        }
        console.log('Search jobs filtered by time range, begin: ' + this.search.startTime + ', end:' + this.search.endTime);
        this.fetchPolicyJobs();
    };
    __decorate([
        core_1.ViewChild(create_edit_policy_component_1.CreateEditPolicyComponent), 
        __metadata('design:type', (typeof (_a = typeof create_edit_policy_component_1.CreateEditPolicyComponent !== 'undefined' && create_edit_policy_component_1.CreateEditPolicyComponent) === 'function' && _a) || Object)
    ], ReplicationComponent.prototype, "createEditPolicyComponent", void 0);
    ReplicationComponent = __decorate([
        core_1.Component({
            selector: 'replicaton',
            template: __webpack_require__(849)
        }), 
        __metadata('design:paramtypes', [(typeof (_b = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _b) || Object, (typeof (_c = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _c) || Object, (typeof (_d = typeof replication_service_1.ReplicationService !== 'undefined' && replication_service_1.ReplicationService) === 'function' && _d) || Object, (typeof (_e = typeof router_1.ActivatedRoute !== 'undefined' && router_1.ActivatedRoute) === 'function' && _e) || Object])
    ], ReplicationComponent);
    return ReplicationComponent;
    var _a, _b, _c, _d, _e;
}());
exports.ReplicationComponent = ReplicationComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/replication.component.js.map

/***/ }),

/***/ 403:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var replication_service_1 = __webpack_require__(79);
var create_edit_policy_component_1 = __webpack_require__(252);
var message_service_1 = __webpack_require__(10);
var shared_const_1 = __webpack_require__(2);
var TotalReplicationComponent = (function () {
    function TotalReplicationComponent(replicationService, messageService) {
        this.replicationService = replicationService;
        this.messageService = messageService;
        this.policyName = '';
    }
    TotalReplicationComponent.prototype.ngOnInit = function () {
        this.retrievePolicies();
    };
    TotalReplicationComponent.prototype.retrievePolicies = function () {
        var _this = this;
        this.replicationService
            .listPolicies(this.policyName)
            .subscribe(function (response) {
            _this.changedPolicies = response;
            _this.policies = _this.changedPolicies;
        }, function (error) { return _this.messageService.announceMessage(error.status, 'Failed to get policies.', shared_const_1.AlertType.DANGER); });
    };
    TotalReplicationComponent.prototype.doSearchPolicies = function (policyName) {
        this.policyName = policyName;
        this.retrievePolicies();
    };
    TotalReplicationComponent.prototype.openEditPolicy = function (policyId) {
        console.log('Open modal to edit policy ID:' + policyId);
        this.createEditPolicyComponent.openCreateEditPolicy(policyId);
    };
    TotalReplicationComponent.prototype.selectPolicy = function (policy) {
        if (policy) {
            this.projectId = policy.project_id;
        }
    };
    TotalReplicationComponent.prototype.refreshPolicies = function () {
        this.retrievePolicies();
    };
    TotalReplicationComponent.prototype.reloadPolicies = function (isReady) {
        if (isReady) {
            this.retrievePolicies();
        }
    };
    __decorate([
        core_1.ViewChild(create_edit_policy_component_1.CreateEditPolicyComponent), 
        __metadata('design:type', (typeof (_a = typeof create_edit_policy_component_1.CreateEditPolicyComponent !== 'undefined' && create_edit_policy_component_1.CreateEditPolicyComponent) === 'function' && _a) || Object)
    ], TotalReplicationComponent.prototype, "createEditPolicyComponent", void 0);
    TotalReplicationComponent = __decorate([
        core_1.Component({
            selector: 'total-replication',
            template: __webpack_require__(850),
            providers: [replication_service_1.ReplicationService]
        }), 
        __metadata('design:paramtypes', [(typeof (_b = typeof replication_service_1.ReplicationService !== 'undefined' && replication_service_1.ReplicationService) === 'function' && _b) || Object, (typeof (_c = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _c) || Object])
    ], TotalReplicationComponent);
    return TotalReplicationComponent;
    var _a, _b, _c;
}());
exports.TotalReplicationComponent = TotalReplicationComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/total-replication.component.js.map

/***/ }),

/***/ 404:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var repository_service_1 = __webpack_require__(250);
var message_service_1 = __webpack_require__(10);
var shared_const_1 = __webpack_require__(2);
var deletion_dialog_service_1 = __webpack_require__(48);
var deletion_message_1 = __webpack_require__(68);
var repositoryTypes = [
    { key: '0', description: 'REPOSITORY.MY_REPOSITORY' },
    { key: '1', description: 'REPOSITORY.PUBLIC_REPOSITORY' }
];
var RepositoryComponent = (function () {
    function RepositoryComponent(route, repositoryService, messageService, deletionDialogService) {
        var _this = this;
        this.route = route;
        this.repositoryService = repositoryService;
        this.messageService = messageService;
        this.deletionDialogService = deletionDialogService;
        this.repositoryTypes = repositoryTypes;
        this.page = 1;
        this.pageSize = 15;
        this.subscription = this.deletionDialogService
            .deletionConfirm$
            .subscribe(function (message) {
            var repoName = message.data;
            _this.repositoryService
                .deleteRepository(repoName)
                .subscribe(function (response) {
                _this.refresh();
                console.log('Successful deleted repo:' + repoName);
            }, function (error) { return _this.messageService.announceMessage(error.status, 'Failed to delete repo:' + repoName, shared_const_1.AlertType.DANGER); });
        });
    }
    RepositoryComponent.prototype.ngOnInit = function () {
        this.projectId = this.route.snapshot.parent.params['id'];
        this.currentRepositoryType = this.repositoryTypes[0];
        this.lastFilteredRepoName = '';
        this.retrieve();
    };
    RepositoryComponent.prototype.ngOnDestroy = function () {
        if (this.subscription) {
            this.subscription.unsubscribe();
        }
    };
    RepositoryComponent.prototype.retrieve = function (state) {
        var _this = this;
        if (state) {
            this.page = state.page.to + 1;
        }
        this.repositoryService
            .listRepositories(this.projectId, this.lastFilteredRepoName, this.page, this.pageSize)
            .subscribe(function (response) {
            _this.totalRecordCount = response.headers.get('x-total-count');
            _this.totalPage = Math.ceil(_this.totalRecordCount / _this.pageSize);
            console.log('TotalRecordCount:' + _this.totalRecordCount + ', totalPage:' + _this.totalPage);
            _this.changedRepositories = response.json();
        }, function (error) { return _this.messageService.announceMessage(error.status, 'Failed to list repositories.', shared_const_1.AlertType.DANGER); });
    };
    RepositoryComponent.prototype.doFilterRepositoryByType = function (type) {
        this.currentRepositoryType = this.repositoryTypes.find(function (r) { return r.key == type; });
    };
    RepositoryComponent.prototype.doSearchRepoNames = function (repoName) {
        this.lastFilteredRepoName = repoName;
        this.retrieve();
    };
    RepositoryComponent.prototype.deleteRepo = function (repoName) {
        var message = new deletion_message_1.DeletionMessage('REPOSITORY.DELETION_TITLE_REPO', 'REPOSITORY.DELETION_SUMMARY_REPO', repoName, repoName, shared_const_1.DeletionTargets.REPOSITORY);
        this.deletionDialogService.openComfirmDialog(message);
    };
    RepositoryComponent.prototype.refresh = function () {
        this.retrieve();
    };
    RepositoryComponent = __decorate([
        core_1.Component({
            selector: 'repository',
            template: __webpack_require__(852)
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof router_1.ActivatedRoute !== 'undefined' && router_1.ActivatedRoute) === 'function' && _a) || Object, (typeof (_b = typeof repository_service_1.RepositoryService !== 'undefined' && repository_service_1.RepositoryService) === 'function' && _b) || Object, (typeof (_c = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _c) || Object, (typeof (_d = typeof deletion_dialog_service_1.DeletionDialogService !== 'undefined' && deletion_dialog_service_1.DeletionDialogService) === 'function' && _d) || Object])
    ], RepositoryComponent);
    return RepositoryComponent;
    var _a, _b, _c, _d;
}());
exports.RepositoryComponent = RepositoryComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/repository.component.js.map

/***/ }),

/***/ 405:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var shared_module_1 = __webpack_require__(52);
var repository_component_1 = __webpack_require__(404);
var list_repository_component_1 = __webpack_require__(619);
var tag_repository_component_1 = __webpack_require__(406);
var top_repo_component_1 = __webpack_require__(622);
var repository_service_1 = __webpack_require__(250);
var RepositoryModule = (function () {
    function RepositoryModule() {
    }
    RepositoryModule = __decorate([
        core_1.NgModule({
            imports: [
                shared_module_1.SharedModule,
                router_1.RouterModule
            ],
            declarations: [
                repository_component_1.RepositoryComponent,
                list_repository_component_1.ListRepositoryComponent,
                tag_repository_component_1.TagRepositoryComponent,
                top_repo_component_1.TopRepoComponent
            ],
            exports: [repository_component_1.RepositoryComponent, list_repository_component_1.ListRepositoryComponent, top_repo_component_1.TopRepoComponent],
            providers: [repository_service_1.RepositoryService]
        }), 
        __metadata('design:paramtypes', [])
    ], RepositoryModule);
    return RepositoryModule;
}());
exports.RepositoryModule = RepositoryModule;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/repository.module.js.map

/***/ }),

/***/ 406:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var repository_service_1 = __webpack_require__(250);
var message_service_1 = __webpack_require__(10);
var shared_const_1 = __webpack_require__(2);
var deletion_dialog_service_1 = __webpack_require__(48);
var deletion_message_1 = __webpack_require__(68);
var tag_view_1 = __webpack_require__(621);
var TagRepositoryComponent = (function () {
    function TagRepositoryComponent(route, messageService, deletionDialogService, repositoryService) {
        var _this = this;
        this.route = route;
        this.messageService = messageService;
        this.deletionDialogService = deletionDialogService;
        this.repositoryService = repositoryService;
        this.subscription = this.deletionDialogService.deletionConfirm$.subscribe(function (message) {
            var tag = message.data;
            if (tag) {
                if (tag.verified) {
                    return;
                }
                else {
                    var tagName_1 = tag.tag;
                    _this.repositoryService
                        .deleteRepoByTag(_this.repoName, tagName_1)
                        .subscribe(function (response) {
                        _this.retrieve();
                        console.log('Deleted repo:' + _this.repoName + ' with tag:' + tagName_1);
                    }, function (error) { return _this.messageService.announceMessage(error.status, 'Failed to delete tag:' + tagName_1 + ' under repo:' + _this.repoName, shared_const_1.AlertType.DANGER); });
                }
            }
        });
    }
    TagRepositoryComponent.prototype.ngOnInit = function () {
        this.projectId = this.route.snapshot.params['id'];
        this.repoName = this.route.snapshot.params['repo'];
        this.tags = [];
        this.retrieve();
    };
    TagRepositoryComponent.prototype.ngOnDestroy = function () {
        if (this.subscription) {
            this.subscription.unsubscribe();
        }
    };
    TagRepositoryComponent.prototype.retrieve = function () {
        var _this = this;
        this.tags = [];
        this.repositoryService
            .listTagsWithVerifiedSignatures(this.repoName)
            .subscribe(function (items) {
            items.forEach(function (t) {
                var tag = new tag_view_1.TagView();
                tag.tag = t.tag;
                var data = JSON.parse(t.manifest.history[0].v1Compatibility);
                tag.architecture = data['architecture'];
                tag.author = data['author'];
                tag.verified = t.verified || false;
                tag.created = data['created'];
                tag.dockerVersion = data['docker_version'];
                tag.pullCommand = 'docker pull ' + t.manifest.name + ':' + t.tag;
                tag.os = data['os'];
                _this.tags.push(tag);
            });
        }, function (error) { return _this.messageService.announceMessage(error.status, 'Failed to list tags with repo:' + _this.repoName, shared_const_1.AlertType.DANGER); });
    };
    TagRepositoryComponent.prototype.deleteTag = function (tag) {
        if (tag) {
            var titleKey = void 0, summaryKey = void 0;
            if (tag.verified) {
                titleKey = 'REPOSITORY.DELETION_TITLE_TAG_DENIED';
                summaryKey = 'REPOSITORY.DELETION_SUMMARY_TAG_DENIED';
            }
            else {
                titleKey = 'REPOSITORY.DELETION_TITLE_TAG';
                summaryKey = 'REPOSITORY.DELETION_SUMMARY_TAG';
            }
            var message = new deletion_message_1.DeletionMessage(titleKey, summaryKey, tag.tag, tag, shared_const_1.DeletionTargets.TAG);
            this.deletionDialogService.openComfirmDialog(message);
        }
    };
    TagRepositoryComponent = __decorate([
        core_1.Component({
            selector: 'tag-repository',
            template: __webpack_require__(853)
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof router_1.ActivatedRoute !== 'undefined' && router_1.ActivatedRoute) === 'function' && _a) || Object, (typeof (_b = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _b) || Object, (typeof (_c = typeof deletion_dialog_service_1.DeletionDialogService !== 'undefined' && deletion_dialog_service_1.DeletionDialogService) === 'function' && _c) || Object, (typeof (_d = typeof repository_service_1.RepositoryService !== 'undefined' && repository_service_1.RepositoryService) === 'function' && _d) || Object])
    ], TagRepositoryComponent);
    return TagRepositoryComponent;
    var _a, _b, _c, _d;
}());
exports.TagRepositoryComponent = TagRepositoryComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/tag-repository.component.js.map

/***/ }),

/***/ 407:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var AboutDialogComponent = (function () {
    function AboutDialogComponent() {
        this.opened = false;
        this.version = "0.4.1";
        this.build = "4276418";
    }
    AboutDialogComponent.prototype.open = function () {
        this.opened = true;
    };
    AboutDialogComponent.prototype.close = function () {
        this.opened = false;
    };
    AboutDialogComponent = __decorate([
        core_1.Component({
            selector: 'about-dialog',
            template: __webpack_require__(855),
            styles: [__webpack_require__(813)]
        }), 
        __metadata('design:paramtypes', [])
    ], AboutDialogComponent);
    return AboutDialogComponent;
}());
exports.AboutDialogComponent = AboutDialogComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/about-dialog.component.js.map

/***/ }),

/***/ 408:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var defaultInterval = 1000;
var defaultLeftTime = 5;
var PageNotFoundComponent = (function () {
    function PageNotFoundComponent(router) {
        this.router = router;
        this.leftSeconds = defaultLeftTime;
        this.timeInterval = null;
    }
    PageNotFoundComponent.prototype.ngOnInit = function () {
        var _this = this;
        if (!this.timeInterval) {
            this.timeInterval = setInterval(function (interval) {
                _this.leftSeconds--;
                if (_this.leftSeconds <= 0) {
                    _this.router.navigate(['harbor']);
                    clearInterval(_this.timeInterval);
                }
            }, defaultInterval);
        }
    };
    PageNotFoundComponent.prototype.ngOnDestroy = function () {
        if (this.timeInterval) {
            clearInterval(this.timeInterval);
        }
    };
    PageNotFoundComponent = __decorate([
        core_1.Component({
            selector: 'page-not-found',
            template: __webpack_require__(863),
            styles: [__webpack_require__(817)]
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof router_1.Router !== 'undefined' && router_1.Router) === 'function' && _a) || Object])
    ], PageNotFoundComponent);
    return PageNotFoundComponent;
    var _a;
}());
exports.PageNotFoundComponent = PageNotFoundComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/not-found.component.js.map

/***/ }),

/***/ 409:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var session_service_1 = __webpack_require__(14);
var shared_const_1 = __webpack_require__(2);
var AuthCheckGuard = (function () {
    function AuthCheckGuard(authService, router) {
        this.authService = authService;
        this.router = router;
    }
    AuthCheckGuard.prototype.canActivate = function (route, state) {
        var _this = this;
        return new Promise(function (resolve, reject) {
            var user = _this.authService.getCurrentUser();
            if (!user) {
                _this.authService.retrieveUser()
                    .then(function () { return resolve(true); })
                    .catch(function (error) {
                    //Session retrieving failed then redirect to sign-in
                    //no matter what status code is.
                    //Please pay attention that route 'harborRootRoute' support anonymous user
                    if (state.url != shared_const_1.harborRootRoute) {
                        var navigatorExtra = {
                            queryParams: { "redirect_url": state.url }
                        };
                        _this.router.navigate([shared_const_1.signInRoute], navigatorExtra);
                        return resolve(false);
                    }
                    else {
                        return resolve(true);
                    }
                });
            }
            else {
                return resolve(true);
            }
        });
    };
    AuthCheckGuard.prototype.canActivateChild = function (route, state) {
        return this.canActivate(route, state);
    };
    AuthCheckGuard = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _a) || Object, (typeof (_b = typeof router_1.Router !== 'undefined' && router_1.Router) === 'function' && _b) || Object])
    ], AuthCheckGuard);
    return AuthCheckGuard;
    var _a, _b;
}());
exports.AuthCheckGuard = AuthCheckGuard;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/auth-user-activate.service.js.map

/***/ }),

/***/ 410:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var session_service_1 = __webpack_require__(14);
var shared_const_1 = __webpack_require__(2);
var SignInGuard = (function () {
    function SignInGuard(authService, router) {
        this.authService = authService;
        this.router = router;
    }
    SignInGuard.prototype.canActivate = function (route, state) {
        var _this = this;
        //If user has logged in, should not login again
        return new Promise(function (resolve, reject) {
            var user = _this.authService.getCurrentUser();
            if (!user) {
                _this.authService.retrieveUser()
                    .then(function () {
                    _this.router.navigate([shared_const_1.harborRootRoute]);
                    return resolve(false);
                })
                    .catch(function (error) {
                    return resolve(true);
                });
            }
            else {
                _this.router.navigate([shared_const_1.harborRootRoute]);
                return resolve(false);
            }
        });
    };
    SignInGuard.prototype.canActivateChild = function (route, state) {
        return this.canActivate(route, state);
    };
    SignInGuard = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _a) || Object, (typeof (_b = typeof router_1.Router !== 'undefined' && router_1.Router) === 'function' && _b) || Object])
    ], SignInGuard);
    return SignInGuard;
    var _a, _b;
}());
exports.SignInGuard = SignInGuard;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/sign-in-guard-activate.service.js.map

/***/ }),

/***/ 411:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var session_service_1 = __webpack_require__(14);
var shared_const_1 = __webpack_require__(2);
var SystemAdminGuard = (function () {
    function SystemAdminGuard(authService, router) {
        this.authService = authService;
        this.router = router;
    }
    SystemAdminGuard.prototype.canActivate = function (route, state) {
        var _this = this;
        return new Promise(function (resolve, reject) {
            var user = _this.authService.getCurrentUser();
            if (!user) {
                _this.authService.retrieveUser()
                    .then(function () {
                    //updated user
                    user = _this.authService.getCurrentUser();
                    if (user.has_admin_role > 0) {
                        return resolve(true);
                    }
                    else {
                        _this.router.navigate([shared_const_1.harborRootRoute]);
                        return resolve(false);
                    }
                })
                    .catch(function (error) {
                    //Session retrieving failed then redirect to sign-in
                    //no matter what status code is.
                    //Please pay attention that route 'harborRootRoute' support anonymous user
                    if (state.url != shared_const_1.harborRootRoute) {
                        var navigatorExtra = {
                            queryParams: { "redirect_url": state.url }
                        };
                        _this.router.navigate([shared_const_1.signInRoute], navigatorExtra);
                        return resolve(false);
                    }
                    else {
                        return resolve(true);
                    }
                });
            }
            else {
                if (user.has_admin_role > 0) {
                    return resolve(true);
                }
                else {
                    _this.router.navigate([shared_const_1.harborRootRoute]);
                    return resolve(false);
                }
            }
        });
    };
    SystemAdminGuard.prototype.canActivateChild = function (route, state) {
        return this.canActivate(route, state);
    };
    SystemAdminGuard = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _a) || Object, (typeof (_b = typeof router_1.Router !== 'undefined' && router_1.Router) === 'function' && _b) || Object])
    ], SystemAdminGuard);
    return SystemAdminGuard;
    var _a, _b;
}());
exports.SystemAdminGuard = SystemAdminGuard;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/system-admin-activate.service.js.map

/***/ }),

/***/ 412:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var new_user_form_component_1 = __webpack_require__(253);
var session_service_1 = __webpack_require__(14);
var user_service_1 = __webpack_require__(175);
var shared_utils_1 = __webpack_require__(33);
var message_service_1 = __webpack_require__(10);
var shared_const_1 = __webpack_require__(2);
var inline_alert_component_1 = __webpack_require__(80);
var NewUserModalComponent = (function () {
    function NewUserModalComponent(session, userService, msgService) {
        this.session = session;
        this.userService = userService;
        this.msgService = msgService;
        this.opened = false;
        this.onGoing = false;
        this.formValueChanged = false;
        this.addNew = new core_1.EventEmitter();
    }
    NewUserModalComponent.prototype.getNewUser = function () {
        return this.newUserForm.getData();
    };
    Object.defineProperty(NewUserModalComponent.prototype, "inProgress", {
        get: function () {
            return this.onGoing;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(NewUserModalComponent.prototype, "isValid", {
        get: function () {
            return this.newUserForm.isValid && this.error == null;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(NewUserModalComponent.prototype, "errorMessage", {
        get: function () {
            return shared_utils_1.errorHandler(this.error);
        },
        enumerable: true,
        configurable: true
    });
    NewUserModalComponent.prototype.formValueChange = function (flag) {
        if (this.error != null) {
            this.error = null; //clear error
        }
        this.formValueChanged = true;
        this.inlineAlert.close();
    };
    NewUserModalComponent.prototype.open = function () {
        this.newUserForm.reset(); //Reset form
        this.formValueChanged = false;
        this.opened = true;
    };
    NewUserModalComponent.prototype.close = function () {
        if (this.formValueChanged) {
            if (this.newUserForm.isEmpty()) {
                this.opened = false;
            }
            else {
                //Need user confirmation
                this.inlineAlert.showInlineConfirmation({
                    message: "ALERT.FORM_CHANGE_CONFIRMATION"
                });
            }
        }
        else {
            this.opened = false;
        }
    };
    NewUserModalComponent.prototype.confirmCancel = function (event) {
        this.opened = false;
    };
    //Create new user
    NewUserModalComponent.prototype.create = function () {
        var _this = this;
        //Double confirm everything is ok
        //Form is valid
        if (!this.isValid) {
            return;
        }
        //We have new user data
        var u = this.getNewUser();
        if (!u) {
            return;
        }
        //Session is ok and role is matched
        var account = this.session.getCurrentUser();
        if (!account || account.has_admin_role === 0) {
            return;
        }
        //Start process
        this.onGoing = true;
        this.userService.addUser(u)
            .then(function () {
            _this.onGoing = false;
            //TODO:
            //As no response data returned, can not add it to list directly
            _this.addNew.emit(u);
            _this.opened = false;
            _this.msgService.announceMessage(200, "USER.SAVE_SUCCESS", shared_const_1.AlertType.SUCCESS);
        })
            .catch(function (error) {
            _this.onGoing = false;
            _this.error = error;
            if (shared_utils_1.accessErrorHandler(error, _this.msgService)) {
                _this.opened = false;
            }
            else {
                _this.inlineAlert.showInlineError(error);
            }
        });
    };
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], NewUserModalComponent.prototype, "addNew", void 0);
    __decorate([
        core_1.ViewChild(new_user_form_component_1.NewUserFormComponent), 
        __metadata('design:type', (typeof (_a = typeof new_user_form_component_1.NewUserFormComponent !== 'undefined' && new_user_form_component_1.NewUserFormComponent) === 'function' && _a) || Object)
    ], NewUserModalComponent.prototype, "newUserForm", void 0);
    __decorate([
        core_1.ViewChild(inline_alert_component_1.InlineAlertComponent), 
        __metadata('design:type', (typeof (_b = typeof inline_alert_component_1.InlineAlertComponent !== 'undefined' && inline_alert_component_1.InlineAlertComponent) === 'function' && _b) || Object)
    ], NewUserModalComponent.prototype, "inlineAlert", void 0);
    NewUserModalComponent = __decorate([
        core_1.Component({
            selector: "new-user-modal",
            template: __webpack_require__(866)
        }), 
        __metadata('design:paramtypes', [(typeof (_c = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _c) || Object, (typeof (_d = typeof user_service_1.UserService !== 'undefined' && user_service_1.UserService) === 'function' && _d) || Object, (typeof (_e = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _e) || Object])
    ], NewUserModalComponent);
    return NewUserModalComponent;
    var _a, _b, _c, _d, _e;
}());
exports.NewUserModalComponent = NewUserModalComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/new-user-modal.component.js.map

/***/ }),

/***/ 413:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
__webpack_require__(58);
var user_service_1 = __webpack_require__(175);
var new_user_modal_component_1 = __webpack_require__(412);
var core_2 = __webpack_require__(34);
var deletion_dialog_service_1 = __webpack_require__(48);
var deletion_message_1 = __webpack_require__(68);
var shared_const_1 = __webpack_require__(2);
var shared_utils_1 = __webpack_require__(33);
var message_service_1 = __webpack_require__(10);
var UserComponent = (function () {
    function UserComponent(userService, translate, deletionDialogService, msgService) {
        var _this = this;
        this.userService = userService;
        this.translate = translate;
        this.deletionDialogService = deletionDialogService;
        this.msgService = msgService;
        this.users = [];
        this.onGoing = false;
        this.adminMenuText = "";
        this.adminColumn = "";
        this.deletionSubscription = deletionDialogService.deletionConfirm$.subscribe(function (confirmed) {
            if (confirmed && confirmed.targetId === shared_const_1.DeletionTargets.USER) {
                _this.delUser(confirmed.data);
            }
        });
    }
    UserComponent.prototype.isMatchFilterTerm = function (terms, testedItem) {
        return testedItem.indexOf(terms) != -1;
    };
    UserComponent.prototype.isSystemAdmin = function (u) {
        var _this = this;
        if (!u) {
            return "{{MISS}}";
        }
        var key = u.has_admin_role ? "USER.IS_ADMIN" : "USER.IS_NOT_ADMIN";
        this.translate.get(key).subscribe(function (res) { return _this.adminColumn = res; });
        return this.adminColumn;
    };
    UserComponent.prototype.adminActions = function (u) {
        var _this = this;
        if (!u) {
            return "{{MISS}}";
        }
        var key = u.has_admin_role ? "USER.DISABLE_ADMIN_ACTION" : "USER.ENABLE_ADMIN_ACTION";
        this.translate.get(key).subscribe(function (res) { return _this.adminMenuText = res; });
        return this.adminMenuText;
    };
    Object.defineProperty(UserComponent.prototype, "inProgress", {
        get: function () {
            return this.onGoing;
        },
        enumerable: true,
        configurable: true
    });
    UserComponent.prototype.ngOnInit = function () {
        this.refreshUser();
    };
    UserComponent.prototype.ngOnDestroy = function () {
        if (this.deletionSubscription) {
            this.deletionSubscription.unsubscribe();
        }
    };
    //Filter items by keywords
    UserComponent.prototype.doFilter = function (terms) {
        var _this = this;
        this.originalUsers.then(function (users) {
            if (terms.trim() === "") {
                _this.users = users;
            }
            else {
                _this.users = users.filter(function (user) {
                    return _this.isMatchFilterTerm(terms, user.username);
                });
            }
        });
    };
    //Disable the admin role for the specified user
    UserComponent.prototype.changeAdminRole = function (user) {
        var _this = this;
        //Double confirm user is existing
        if (!user || user.user_id === 0) {
            return;
        }
        //Value copy
        var updatedUser = {
            user_id: user.user_id
        };
        if (user.has_admin_role === 0) {
            updatedUser.has_admin_role = 1; //Set as admin
        }
        else {
            updatedUser.has_admin_role = 0; //Set as none admin
        }
        this.userService.updateUserRole(updatedUser)
            .then(function () {
            //Change view now
            user.has_admin_role = updatedUser.has_admin_role;
        })
            .catch(function (error) {
            if (!shared_utils_1.accessErrorHandler(error, _this.msgService)) {
                _this.msgService.announceMessage(500, shared_utils_1.errorHandler(error), shared_const_1.AlertType.DANGER);
            }
        });
    };
    //Delete the specified user
    UserComponent.prototype.deleteUser = function (user) {
        if (!user) {
            return;
        }
        //Confirm deletion
        var msg = new deletion_message_1.DeletionMessage("USER.DELETION_TITLE", "USER.DELETION_SUMMARY", user.username, user, shared_const_1.DeletionTargets.USER);
        this.deletionDialogService.openComfirmDialog(msg);
    };
    UserComponent.prototype.delUser = function (user) {
        var _this = this;
        this.userService.deleteUser(user.user_id)
            .then(function () {
            //Remove it from current user list
            //and then view refreshed
            _this.originalUsers.then(function (users) {
                _this.users = users.filter(function (u) { return u.user_id != user.user_id; });
                _this.msgService.announceMessage(500, "USER.DELETE_SUCCESS", shared_const_1.AlertType.SUCCESS);
            });
        })
            .catch(function (error) {
            if (!shared_utils_1.accessErrorHandler(error, _this.msgService)) {
                _this.msgService.announceMessage(500, shared_utils_1.errorHandler(error), shared_const_1.AlertType.DANGER);
            }
        });
    };
    //Refresh the user list
    UserComponent.prototype.refreshUser = function () {
        var _this = this;
        //Start to get
        this.onGoing = true;
        this.originalUsers = this.userService.getUsers()
            .then(function (users) {
            _this.onGoing = false;
            _this.users = users;
            return users;
        })
            .catch(function (error) {
            _this.onGoing = false;
            if (!shared_utils_1.accessErrorHandler(error, _this.msgService)) {
                _this.msgService.announceMessage(500, shared_utils_1.errorHandler(error), shared_const_1.AlertType.DANGER);
            }
        });
    };
    //Add new user
    UserComponent.prototype.addNewUser = function () {
        this.newUserDialog.open();
    };
    //Add user to the user list
    UserComponent.prototype.addUserToList = function (user) {
        //Currently we can only add it by reloading all
        this.refreshUser();
    };
    __decorate([
        core_1.ViewChild(new_user_modal_component_1.NewUserModalComponent), 
        __metadata('design:type', (typeof (_a = typeof new_user_modal_component_1.NewUserModalComponent !== 'undefined' && new_user_modal_component_1.NewUserModalComponent) === 'function' && _a) || Object)
    ], UserComponent.prototype, "newUserDialog", void 0);
    UserComponent = __decorate([
        core_1.Component({
            selector: 'harbor-user',
            template: __webpack_require__(867),
            styles: [__webpack_require__(818)],
            providers: [user_service_1.UserService]
        }), 
        __metadata('design:paramtypes', [(typeof (_b = typeof user_service_1.UserService !== 'undefined' && user_service_1.UserService) === 'function' && _b) || Object, (typeof (_c = typeof core_2.TranslateService !== 'undefined' && core_2.TranslateService) === 'function' && _c) || Object, (typeof (_d = typeof deletion_dialog_service_1.DeletionDialogService !== 'undefined' && deletion_dialog_service_1.DeletionDialogService) === 'function' && _d) || Object, (typeof (_e = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _e) || Object])
    ], UserComponent);
    return UserComponent;
    var _a, _b, _c, _d, _e;
}());
exports.UserComponent = UserComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/user.component.js.map

/***/ }),

/***/ 457:
/***/ (function(module, exports) {

module.exports = ".reset-modal-title-override {\n    font-size: 14px !important;\n}"

/***/ }),

/***/ 458:
/***/ (function(module, exports) {

module.exports = ".statistic-wrapper {\n    padding: 12px;\n    margin: 12px;\n    text-align: center;\n    vertical-align: middle;\n    height: 72px;\n    min-width: 108px;\n    max-width: 216px;\n    display: inline-block;\n}\n\n.statistic-data {\n    font-size: 48px;\n    font-weight: bolder;\n    font-family: \"Metropolis\";\n    line-height: 48px;\n}\n\n.statistic-text {\n    font-size: 24px;\n    font-weight: 400;\n    line-height: 24px;\n    text-transform: uppercase;\n    font-family: \"Metropolis\";\n}\n\n.statistic-column-title {\n    position: relative;\n    top: 40%;\n}"

/***/ }),

/***/ 477:
/***/ (function(module, exports) {

function webpackEmptyContext(req) {
	throw new Error("Cannot find module '" + req + "'.");
}
webpackEmptyContext.keys = function() { return []; };
webpackEmptyContext.resolve = webpackEmptyContext;
module.exports = webpackEmptyContext;
webpackEmptyContext.id = 477;


/***/ }),

/***/ 478:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

__webpack_require__(640);
var platform_browser_dynamic_1 = __webpack_require__(571);
var core_1 = __webpack_require__(0);
var environment_1 = __webpack_require__(639);
var _1 = __webpack_require__(610);
if (environment_1.environment.production) {
    core_1.enableProdMode();
}
platform_browser_dynamic_1.platformBrowserDynamic().bootstrapModule(_1.AppModule);
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/main.js.map

/***/ }),

/***/ 48:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var Subject_1 = __webpack_require__(25);
var DeletionDialogService = (function () {
    function DeletionDialogService() {
        this.deletionAnnoucedSource = new Subject_1.Subject();
        this.deletionConfirmSource = new Subject_1.Subject();
        this.deletionAnnouced$ = this.deletionAnnoucedSource.asObservable();
        this.deletionConfirm$ = this.deletionConfirmSource.asObservable();
    }
    DeletionDialogService.prototype.confirmDeletion = function (message) {
        this.deletionConfirmSource.next(message);
    };
    DeletionDialogService.prototype.openComfirmDialog = function (message) {
        this.deletionAnnoucedSource.next(message);
    };
    DeletionDialogService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [])
    ], DeletionDialogService);
    return DeletionDialogService;
}());
exports.DeletionDialogService = DeletionDialogService;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/deletion-dialog.service.js.map

/***/ }),

/***/ 52:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var core_module_1 = __webpack_require__(246);
var core_2 = __webpack_require__(257);
var session_service_1 = __webpack_require__(14);
var message_component_1 = __webpack_require__(607);
var message_service_1 = __webpack_require__(10);
var max_length_ext_directive_1 = __webpack_require__(629);
var filter_component_1 = __webpack_require__(626);
var harbor_action_overflow_1 = __webpack_require__(627);
var core_3 = __webpack_require__(34);
var router_1 = __webpack_require__(8);
var deletion_dialog_component_1 = __webpack_require__(625);
var deletion_dialog_service_1 = __webpack_require__(48);
var base_routing_resolver_service_1 = __webpack_require__(631);
var system_admin_activate_service_1 = __webpack_require__(411);
var new_user_form_component_1 = __webpack_require__(253);
var inline_alert_component_1 = __webpack_require__(80);
var list_policy_component_1 = __webpack_require__(628);
var create_edit_policy_component_1 = __webpack_require__(252);
var port_directive_1 = __webpack_require__(630);
var not_found_component_1 = __webpack_require__(408);
var about_dialog_component_1 = __webpack_require__(407);
var auth_user_activate_service_1 = __webpack_require__(409);
var statistics_component_1 = __webpack_require__(634);
var statistics_panel_component_1 = __webpack_require__(633);
var sign_in_guard_activate_service_1 = __webpack_require__(410);
var SharedModule = (function () {
    function SharedModule() {
    }
    SharedModule = __decorate([
        core_1.NgModule({
            imports: [
                core_module_1.CoreModule,
                core_3.TranslateModule,
                router_1.RouterModule
            ],
            declarations: [
                message_component_1.MessageComponent,
                max_length_ext_directive_1.MaxLengthExtValidatorDirective,
                filter_component_1.FilterComponent,
                harbor_action_overflow_1.HarborActionOverflow,
                deletion_dialog_component_1.DeletionDialogComponent,
                new_user_form_component_1.NewUserFormComponent,
                inline_alert_component_1.InlineAlertComponent,
                list_policy_component_1.ListPolicyComponent,
                create_edit_policy_component_1.CreateEditPolicyComponent,
                port_directive_1.PortValidatorDirective,
                not_found_component_1.PageNotFoundComponent,
                about_dialog_component_1.AboutDialogComponent,
                statistics_component_1.StatisticsComponent,
                statistics_panel_component_1.StatisticsPanelComponent
            ],
            exports: [
                core_module_1.CoreModule,
                message_component_1.MessageComponent,
                max_length_ext_directive_1.MaxLengthExtValidatorDirective,
                filter_component_1.FilterComponent,
                harbor_action_overflow_1.HarborActionOverflow,
                core_3.TranslateModule,
                deletion_dialog_component_1.DeletionDialogComponent,
                new_user_form_component_1.NewUserFormComponent,
                inline_alert_component_1.InlineAlertComponent,
                list_policy_component_1.ListPolicyComponent,
                create_edit_policy_component_1.CreateEditPolicyComponent,
                port_directive_1.PortValidatorDirective,
                not_found_component_1.PageNotFoundComponent,
                about_dialog_component_1.AboutDialogComponent,
                statistics_component_1.StatisticsComponent,
                statistics_panel_component_1.StatisticsPanelComponent
            ],
            providers: [
                session_service_1.SessionService,
                message_service_1.MessageService,
                core_2.CookieService,
                deletion_dialog_service_1.DeletionDialogService,
                base_routing_resolver_service_1.BaseRoutingResolver,
                system_admin_activate_service_1.SystemAdminGuard,
                auth_user_activate_service_1.AuthCheckGuard,
                sign_in_guard_activate_service_1.SignInGuard]
        }), 
        __metadata('design:paramtypes', [])
    ], SharedModule);
    return SharedModule;
}());
exports.SharedModule = SharedModule;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/shared.module.js.map

/***/ }),

/***/ 600:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var app_component_1 = __webpack_require__(380);
var base_module_1 = __webpack_require__(601);
var harbor_routing_module_1 = __webpack_require__(608);
var shared_module_1 = __webpack_require__(52);
var account_module_1 = __webpack_require__(375);
var config_module_1 = __webpack_require__(606);
var core_2 = __webpack_require__(34);
var missing_trans_handler_1 = __webpack_require__(609);
var http_loader_1 = __webpack_require__(641);
var http_1 = __webpack_require__(20);
var app_config_service_1 = __webpack_require__(172);
function HttpLoaderFactory(http) {
    return new http_loader_1.TranslateHttpLoader(http, 'i18n/lang/', '-lang.json');
}
exports.HttpLoaderFactory = HttpLoaderFactory;
function initConfig(configService) {
    return function () { return configService.load(); };
}
exports.initConfig = initConfig;
var AppModule = (function () {
    function AppModule() {
    }
    AppModule = __decorate([
        core_1.NgModule({
            declarations: [
                app_component_1.AppComponent,
            ],
            imports: [
                shared_module_1.SharedModule,
                base_module_1.BaseModule,
                account_module_1.AccountModule,
                harbor_routing_module_1.HarborRoutingModule,
                config_module_1.ConfigurationModule,
                core_2.TranslateModule.forRoot({
                    loader: {
                        provide: core_2.TranslateLoader,
                        useFactory: (HttpLoaderFactory),
                        deps: [http_1.Http]
                    },
                    missingTranslationHandler: {
                        provide: core_2.MissingTranslationHandler,
                        useClass: missing_trans_handler_1.MyMissingTranslationHandler
                    }
                })
            ],
            providers: [
                app_config_service_1.AppConfigService,
                {
                    provide: core_1.APP_INITIALIZER,
                    useFactory: initConfig,
                    deps: [app_config_service_1.AppConfigService],
                    multi: true
                }],
            bootstrap: [app_component_1.AppComponent]
        }), 
        __metadata('design:paramtypes', [])
    ], AppModule);
    return AppModule;
}());
exports.AppModule = AppModule;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/app.module.js.map

/***/ }),

/***/ 601:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var shared_module_1 = __webpack_require__(52);
var router_1 = __webpack_require__(8);
var project_module_1 = __webpack_require__(614);
var user_module_1 = __webpack_require__(637);
var account_module_1 = __webpack_require__(375);
var repository_module_1 = __webpack_require__(405);
var navigator_component_1 = __webpack_require__(384);
var global_search_component_1 = __webpack_require__(603);
var footer_component_1 = __webpack_require__(602);
var harbor_shell_component_1 = __webpack_require__(382);
var search_result_component_1 = __webpack_require__(381);
var start_component_1 = __webpack_require__(245);
var search_trigger_service_1 = __webpack_require__(95);
var BaseModule = (function () {
    function BaseModule() {
    }
    BaseModule = __decorate([
        core_1.NgModule({
            imports: [
                shared_module_1.SharedModule,
                project_module_1.ProjectModule,
                user_module_1.UserModule,
                account_module_1.AccountModule,
                router_1.RouterModule,
                repository_module_1.RepositoryModule
            ],
            declarations: [
                navigator_component_1.NavigatorComponent,
                global_search_component_1.GlobalSearchComponent,
                footer_component_1.FooterComponent,
                harbor_shell_component_1.HarborShellComponent,
                search_result_component_1.SearchResultComponent,
                start_component_1.StartPageComponent
            ],
            exports: [harbor_shell_component_1.HarborShellComponent],
            providers: [search_trigger_service_1.SearchTriggerService]
        }), 
        __metadata('design:paramtypes', [])
    ], BaseModule);
    return BaseModule;
}());
exports.BaseModule = BaseModule;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/base.module.js.map

/***/ }),

/***/ 602:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var FooterComponent = (function () {
    function FooterComponent() {
    }
    FooterComponent = __decorate([
        core_1.Component({
            selector: 'footer',
            template: __webpack_require__(827)
        }), 
        __metadata('design:paramtypes', [])
    ], FooterComponent);
    return FooterComponent;
}());
exports.FooterComponent = FooterComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/footer.component.js.map

/***/ }),

/***/ 603:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var Subject_1 = __webpack_require__(25);
var search_trigger_service_1 = __webpack_require__(95);
__webpack_require__(461);
__webpack_require__(462);
var deBounceTime = 500; //ms
var GlobalSearchComponent = (function () {
    function GlobalSearchComponent(searchTrigger, router) {
        this.searchTrigger = searchTrigger;
        this.router = router;
        //Keep search term as Subject
        this.searchTerms = new Subject_1.Subject();
        //To indicate if the result panel is opened
        this.isResPanelOpened = false;
    }
    //Implement ngOnIni
    GlobalSearchComponent.prototype.ngOnInit = function () {
        var _this = this;
        this.searchSub = this.searchTerms
            .debounceTime(deBounceTime)
            .distinctUntilChanged()
            .subscribe(function (term) {
            _this.searchTrigger.triggerSearch(term);
        });
    };
    GlobalSearchComponent.prototype.ngOnDestroy = function () {
        if (this.searchSub) {
            this.searchSub.unsubscribe();
        }
    };
    //Handle the term inputting event
    GlobalSearchComponent.prototype.search = function (term) {
        //Send event even term is empty
        this.searchTerms.next(term.trim());
    };
    GlobalSearchComponent = __decorate([
        //ms
        core_1.Component({
            selector: 'global-search',
            template: __webpack_require__(828)
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof search_trigger_service_1.SearchTriggerService !== 'undefined' && search_trigger_service_1.SearchTriggerService) === 'function' && _a) || Object, (typeof (_b = typeof router_1.Router !== 'undefined' && router_1.Router) === 'function' && _b) || Object])
    ], GlobalSearchComponent);
    return GlobalSearchComponent;
    var _a, _b;
}());
exports.GlobalSearchComponent = GlobalSearchComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/global-search.component.js.map

/***/ }),

/***/ 604:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var http_1 = __webpack_require__(20);
__webpack_require__(58);
var searchEndpoint = "/api/search";
/**
 * Declare service to handle the global search
 *
 *
 * @export
 * @class GlobalSearchService
 */
var GlobalSearchService = (function () {
    function GlobalSearchService(http) {
        this.http = http;
        this.headers = new http_1.Headers({
            "Content-Type": 'application/json'
        });
        this.options = new http_1.RequestOptions({
            headers: this.headers
        });
    }
    /**
     * Search related artifacts with the provided keyword
     *
     * @param {string} keyword
     * @returns {Promise<SearchResults>}
     *
     * @memberOf GlobalSearchService
     */
    GlobalSearchService.prototype.doSearch = function (term) {
        var searchUrl = searchEndpoint + "?q=" + term;
        return this.http.get(searchUrl, this.options).toPromise()
            .then(function (response) { return response.json(); })
            .catch(function (error) { return Promise.reject(error); });
    };
    GlobalSearchService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof http_1.Http !== 'undefined' && http_1.Http) === 'function' && _a) || Object])
    ], GlobalSearchService);
    return GlobalSearchService;
    var _a;
}());
exports.GlobalSearchService = GlobalSearchService;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/global-search.service.js.map

/***/ }),

/***/ 605:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var SearchResults = (function () {
    function SearchResults() {
        this.project = [];
        this.repository = [];
    }
    return SearchResults;
}());
exports.SearchResults = SearchResults;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/search-results.js.map

/***/ }),

/***/ 606:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var core_module_1 = __webpack_require__(246);
var shared_module_1 = __webpack_require__(52);
var config_component_1 = __webpack_require__(386);
var config_service_1 = __webpack_require__(387);
var config_auth_component_1 = __webpack_require__(385);
var config_email_component_1 = __webpack_require__(388);
var ConfigurationModule = (function () {
    function ConfigurationModule() {
    }
    ConfigurationModule = __decorate([
        core_1.NgModule({
            imports: [
                core_module_1.CoreModule,
                shared_module_1.SharedModule
            ],
            declarations: [
                config_component_1.ConfigurationComponent,
                config_auth_component_1.ConfigurationAuthComponent,
                config_email_component_1.ConfigurationEmailComponent],
            exports: [config_component_1.ConfigurationComponent],
            providers: [config_service_1.ConfigurationService]
        }), 
        __metadata('design:paramtypes', [])
    ], ConfigurationModule);
    return ConfigurationModule;
}());
exports.ConfigurationModule = ConfigurationModule;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/config.module.js.map

/***/ }),

/***/ 607:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var core_2 = __webpack_require__(34);
var message_1 = __webpack_require__(389);
var message_service_1 = __webpack_require__(10);
var shared_const_1 = __webpack_require__(2);
var MessageComponent = (function () {
    function MessageComponent(messageService, router, translate) {
        this.messageService = messageService;
        this.router = router;
        this.translate = translate;
        this.globalMessage = new message_1.Message();
        this.messageText = "";
    }
    MessageComponent.prototype.ngOnInit = function () {
        var _this = this;
        //Only subscribe application level message
        if (this.isAppLevel) {
            this.messageService.appLevelAnnounced$.subscribe(function (message) {
                _this.globalMessageOpened = true;
                _this.globalMessage = message;
                _this.messageText = message.message;
                _this.translateMessage(message);
            });
        }
        else {
            //Only subscribe general messages
            this.messageService.messageAnnounced$.subscribe(function (message) {
                _this.globalMessageOpened = true;
                _this.globalMessage = message;
                _this.messageText = message.message;
                _this.translateMessage(message);
                // Make the message alert bar dismiss after several intervals.
                //Only for this case
                setInterval(function () { return _this.onClose(); }, shared_const_1.dismissInterval);
            });
        }
    };
    //Translate or refactor the message shown to user
    MessageComponent.prototype.translateMessage = function (msg) {
        var _this = this;
        if (!msg) {
            return;
        }
        var key = "";
        if (!msg.message) {
            key = "UNKNOWN_ERROR";
        }
        else {
            key = typeof msg.message === "string" ? msg.message.trim() : msg.message;
            if (key === "") {
                key = "UNKNOWN_ERROR";
            }
        }
        //Override key for HTTP 401 and 403
        if (this.globalMessage.statusCode === shared_const_1.httpStatusCode.Unauthorized) {
            key = "UNAUTHORIZED_ERROR";
        }
        if (this.globalMessage.statusCode === shared_const_1.httpStatusCode.Forbidden) {
            key = "FORBIDDEN_ERROR";
        }
        this.translate.get(key).subscribe(function (res) { return _this.messageText = res; });
    };
    Object.defineProperty(MessageComponent.prototype, "needAuth", {
        get: function () {
            return this.globalMessage ?
                (this.globalMessage.statusCode === shared_const_1.httpStatusCode.Unauthorized) ||
                    (this.globalMessage.statusCode === shared_const_1.httpStatusCode.Forbidden) : false;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(MessageComponent.prototype, "message", {
        //Show message text
        get: function () {
            return this.messageText;
        },
        enumerable: true,
        configurable: true
    });
    MessageComponent.prototype.signIn = function () {
        this.router.navigate(['sign-in']);
    };
    MessageComponent.prototype.onClose = function () {
        this.globalMessageOpened = false;
    };
    __decorate([
        core_1.Input(), 
        __metadata('design:type', Boolean)
    ], MessageComponent.prototype, "isAppLevel", void 0);
    MessageComponent = __decorate([
        core_1.Component({
            selector: 'global-message',
            template: __webpack_require__(836)
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _a) || Object, (typeof (_b = typeof router_1.Router !== 'undefined' && router_1.Router) === 'function' && _b) || Object, (typeof (_c = typeof core_2.TranslateService !== 'undefined' && core_2.TranslateService) === 'function' && _c) || Object])
    ], MessageComponent);
    return MessageComponent;
    var _a, _b, _c;
}());
exports.MessageComponent = MessageComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/message.component.js.map

/***/ }),

/***/ 608:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var sign_in_component_1 = __webpack_require__(379);
var harbor_shell_component_1 = __webpack_require__(382);
var project_component_1 = __webpack_require__(398);
var user_component_1 = __webpack_require__(413);
var replication_management_component_1 = __webpack_require__(401);
var total_replication_component_1 = __webpack_require__(403);
var destination_component_1 = __webpack_require__(400);
var project_detail_component_1 = __webpack_require__(396);
var repository_component_1 = __webpack_require__(404);
var tag_repository_component_1 = __webpack_require__(406);
var replication_component_1 = __webpack_require__(402);
var member_component_1 = __webpack_require__(395);
var audit_log_component_1 = __webpack_require__(390);
var project_routing_resolver_service_1 = __webpack_require__(397);
var system_admin_activate_service_1 = __webpack_require__(411);
var sign_up_component_1 = __webpack_require__(243);
var reset_password_component_1 = __webpack_require__(378);
var recent_log_component_1 = __webpack_require__(391);
var config_component_1 = __webpack_require__(386);
var not_found_component_1 = __webpack_require__(408);
var start_component_1 = __webpack_require__(245);
var auth_user_activate_service_1 = __webpack_require__(409);
var sign_in_guard_activate_service_1 = __webpack_require__(410);
var harborRoutes = [
    { path: '', redirectTo: '/harbor/dashboard', pathMatch: 'full' },
    { path: 'harbor', redirectTo: '/harbor/dashboard', pathMatch: 'full' },
    { path: 'sign-in', component: sign_in_component_1.SignInComponent, canActivate: [sign_in_guard_activate_service_1.SignInGuard] },
    { path: 'sign-up', component: sign_up_component_1.SignUpComponent },
    { path: 'password-reset', component: reset_password_component_1.ResetPasswordComponent },
    {
        path: 'harbor',
        component: harbor_shell_component_1.HarborShellComponent,
        children: [
            { path: 'sign-in', component: sign_in_component_1.SignInComponent, canActivate: [sign_in_guard_activate_service_1.SignInGuard] },
            { path: 'sign-up', component: sign_up_component_1.SignUpComponent },
            { path: 'dashboard', component: start_component_1.StartPageComponent, canActivate: [auth_user_activate_service_1.AuthCheckGuard] },
            {
                path: 'projects',
                component: project_component_1.ProjectComponent,
                canActivate: [auth_user_activate_service_1.AuthCheckGuard]
            },
            {
                path: 'logs',
                component: recent_log_component_1.RecentLogComponent,
                canActivate: [auth_user_activate_service_1.AuthCheckGuard]
            },
            {
                path: 'users',
                component: user_component_1.UserComponent,
                canActivate: [auth_user_activate_service_1.AuthCheckGuard, system_admin_activate_service_1.SystemAdminGuard]
            },
            {
                path: 'replications',
                component: replication_management_component_1.ReplicationManagementComponent,
                canActivate: [auth_user_activate_service_1.AuthCheckGuard, system_admin_activate_service_1.SystemAdminGuard],
                canActivateChild: [auth_user_activate_service_1.AuthCheckGuard, system_admin_activate_service_1.SystemAdminGuard],
                children: [
                    {
                        path: 'rules',
                        component: total_replication_component_1.TotalReplicationComponent
                    },
                    {
                        path: 'endpoints',
                        component: destination_component_1.DestinationComponent
                    }
                ]
            },
            {
                path: 'tags/:id/:repo',
                component: tag_repository_component_1.TagRepositoryComponent,
                canActivate: [auth_user_activate_service_1.AuthCheckGuard]
            },
            {
                path: 'projects/:id',
                component: project_detail_component_1.ProjectDetailComponent,
                canActivate: [auth_user_activate_service_1.AuthCheckGuard],
                canActivateChild: [auth_user_activate_service_1.AuthCheckGuard],
                resolve: {
                    projectResolver: project_routing_resolver_service_1.ProjectRoutingResolver
                },
                children: [
                    {
                        path: 'repository',
                        component: repository_component_1.RepositoryComponent
                    },
                    {
                        path: 'replication',
                        component: replication_component_1.ReplicationComponent
                    },
                    {
                        path: 'member',
                        component: member_component_1.MemberComponent
                    },
                    {
                        path: 'log',
                        component: audit_log_component_1.AuditLogComponent
                    }
                ]
            },
            {
                path: 'configs',
                component: config_component_1.ConfigurationComponent,
                canActivate: [auth_user_activate_service_1.AuthCheckGuard, system_admin_activate_service_1.SystemAdminGuard],
            }
        ]
    },
    { path: "**", component: not_found_component_1.PageNotFoundComponent }
];
var HarborRoutingModule = (function () {
    function HarborRoutingModule() {
    }
    HarborRoutingModule = __decorate([
        core_1.NgModule({
            imports: [
                router_1.RouterModule.forRoot(harborRoutes)
            ],
            exports: [router_1.RouterModule]
        }), 
        __metadata('design:paramtypes', [])
    ], HarborRoutingModule);
    return HarborRoutingModule;
}());
exports.HarborRoutingModule = HarborRoutingModule;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/harbor-routing.module.js.map

/***/ }),

/***/ 609:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var MyMissingTranslationHandler = (function () {
    function MyMissingTranslationHandler() {
    }
    MyMissingTranslationHandler.prototype.handle = function (params) {
        var missingText = "{Miss Harbor Text}";
        return params.key || missingText;
    };
    return MyMissingTranslationHandler;
}());
exports.MyMissingTranslationHandler = MyMissingTranslationHandler;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/missing-trans.handler.js.map

/***/ }),

/***/ 610:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

function __export(m) {
    for (var p in m) if (!exports.hasOwnProperty(p)) exports[p] = m[p];
}
__export(__webpack_require__(380));
__export(__webpack_require__(600));
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/index.js.map

/***/ }),

/***/ 611:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

/*
 {
    "log_id": 3,
    "user_id": 0,
    "project_id": 0,
    "repo_name": "library/mysql",
    "repo_tag": "5.6",
    "guid": "",
    "operation": "push",
    "op_time": "2017-02-14T09:22:58Z",
    "username": "admin",
    "keywords": "",
    "BeginTime": "0001-01-01T00:00:00Z",
    "begin_timestamp": 0,
    "EndTime": "0001-01-01T00:00:00Z",
    "end_timestamp": 0
  }
*/
var AuditLog = (function () {
    function AuditLog() {
        this.begin_timestamp = 0;
        this.end_timestamp = 0;
    }
    return AuditLog;
}());
exports.AuditLog = AuditLog;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/audit-log.js.map

/***/ }),

/***/ 612:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var audit_log_component_1 = __webpack_require__(390);
var shared_module_1 = __webpack_require__(52);
var audit_log_service_1 = __webpack_require__(247);
var recent_log_component_1 = __webpack_require__(391);
var LogModule = (function () {
    function LogModule() {
    }
    LogModule = __decorate([
        core_1.NgModule({
            imports: [shared_module_1.SharedModule],
            declarations: [
                audit_log_component_1.AuditLogComponent,
                recent_log_component_1.RecentLogComponent],
            providers: [audit_log_service_1.AuditLogService],
            exports: [
                audit_log_component_1.AuditLogComponent,
                recent_log_component_1.RecentLogComponent]
        }), 
        __metadata('design:paramtypes', [])
    ], LogModule);
    return LogModule;
}());
exports.LogModule = LogModule;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/log.module.js.map

/***/ }),

/***/ 613:
/***/ (function(module, exports, __webpack_require__) {

"use strict";
/*
{
  "user_id": 1,
  "username": "admin",
  "email": "",
  "password": "",
  "realname": "",
  "comment": "",
  "deleted": 0,
  "role_name": "projectAdmin",
  "role_id": 1,
  "has_admin_role": 0,
  "reset_uuid": "",
  "creation_time": "0001-01-01T00:00:00Z",
  "update_time": "0001-01-01T00:00:00Z"
}
*/

var Member = (function () {
    function Member() {
    }
    return Member;
}());
exports.Member = Member;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/member.js.map

/***/ }),

/***/ 614:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var shared_module_1 = __webpack_require__(52);
var repository_module_1 = __webpack_require__(405);
var replication_module_1 = __webpack_require__(618);
var log_module_1 = __webpack_require__(612);
var project_component_1 = __webpack_require__(398);
var create_project_component_1 = __webpack_require__(392);
var list_project_component_1 = __webpack_require__(393);
var project_detail_component_1 = __webpack_require__(396);
var member_component_1 = __webpack_require__(395);
var add_member_component_1 = __webpack_require__(394);
var project_service_1 = __webpack_require__(174);
var member_service_1 = __webpack_require__(248);
var project_routing_resolver_service_1 = __webpack_require__(397);
var ProjectModule = (function () {
    function ProjectModule() {
    }
    ProjectModule = __decorate([
        core_1.NgModule({
            imports: [
                shared_module_1.SharedModule,
                repository_module_1.RepositoryModule,
                replication_module_1.ReplicationModule,
                log_module_1.LogModule,
                router_1.RouterModule
            ],
            declarations: [
                project_component_1.ProjectComponent,
                create_project_component_1.CreateProjectComponent,
                list_project_component_1.ListProjectComponent,
                project_detail_component_1.ProjectDetailComponent,
                member_component_1.MemberComponent,
                add_member_component_1.AddMemberComponent
            ],
            exports: [project_component_1.ProjectComponent, list_project_component_1.ListProjectComponent],
            providers: [project_routing_resolver_service_1.ProjectRoutingResolver, project_service_1.ProjectService, member_service_1.MemberService]
        }), 
        __metadata('design:paramtypes', [])
    ], ProjectModule);
    return ProjectModule;
}());
exports.ProjectModule = ProjectModule;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/project.module.js.map

/***/ }),

/***/ 615:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

/*
  [
    {
        "project_id": 1,
        "owner_id": 1,
        "name": "library",
        "creation_time": "2017-02-10T07:57:56Z",
        "creation_time_str": "",
        "deleted": 0,
        "owner_name": "",
        "public": 1,
        "Togglable": true,
        "update_time": "2017-02-10T07:57:56Z",
        "current_user_role_id": 1,
        "repo_count": 0
    }
  ]
*/
var Project = (function () {
    function Project() {
    }
    return Project;
}());
exports.Project = Project;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/project.js.map

/***/ }),

/***/ 616:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var ListJobComponent = (function () {
    function ListJobComponent() {
        this.paginate = new core_1.EventEmitter();
        this.pageOffset = 1;
    }
    ListJobComponent.prototype.refresh = function (state) {
        if (this.jobs) {
            this.paginate.emit(state);
        }
    };
    __decorate([
        core_1.Input(), 
        __metadata('design:type', Array)
    ], ListJobComponent.prototype, "jobs", void 0);
    __decorate([
        core_1.Input(), 
        __metadata('design:type', Number)
    ], ListJobComponent.prototype, "totalRecordCount", void 0);
    __decorate([
        core_1.Input(), 
        __metadata('design:type', Number)
    ], ListJobComponent.prototype, "totalPage", void 0);
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], ListJobComponent.prototype, "paginate", void 0);
    ListJobComponent = __decorate([
        core_1.Component({
            selector: 'list-job',
            template: __webpack_require__(847)
        }), 
        __metadata('design:paramtypes', [])
    ], ListJobComponent);
    return ListJobComponent;
}());
exports.ListJobComponent = ListJobComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/list-job.component.js.map

/***/ }),

/***/ 617:
/***/ (function(module, exports, __webpack_require__) {

"use strict";
/*
  {
    "id": 1,
    "project_id": 1,
    "project_name": "library",
    "target_id": 1,
    "target_name": "target_01",
    "name": "sync_01",
    "enabled": 0,
    "description": "sync_01 desc.",
    "cron_str": "",
    "start_time": "0001-01-01T00:00:00Z",
    "creation_time": "2017-02-24T06:41:52Z",
    "update_time": "2017-02-24T06:41:52Z",
    "error_job_count": 0,
    "deleted": 0
  }
*/

var Policy = (function () {
    function Policy() {
    }
    return Policy;
}());
exports.Policy = Policy;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/policy.js.map

/***/ }),

/***/ 618:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var replication_management_component_1 = __webpack_require__(401);
var replication_component_1 = __webpack_require__(402);
var list_job_component_1 = __webpack_require__(616);
var total_replication_component_1 = __webpack_require__(403);
var destination_component_1 = __webpack_require__(400);
var create_edit_destination_component_1 = __webpack_require__(399);
var shared_module_1 = __webpack_require__(52);
var replication_service_1 = __webpack_require__(79);
var ReplicationModule = (function () {
    function ReplicationModule() {
    }
    ReplicationModule = __decorate([
        core_1.NgModule({
            imports: [
                shared_module_1.SharedModule,
                router_1.RouterModule
            ],
            declarations: [
                replication_component_1.ReplicationComponent,
                replication_management_component_1.ReplicationManagementComponent,
                list_job_component_1.ListJobComponent,
                total_replication_component_1.TotalReplicationComponent,
                destination_component_1.DestinationComponent,
                create_edit_destination_component_1.CreateEditDestinationComponent
            ],
            exports: [replication_component_1.ReplicationComponent],
            providers: [replication_service_1.ReplicationService]
        }), 
        __metadata('design:paramtypes', [])
    ], ReplicationModule);
    return ReplicationModule;
}());
exports.ReplicationModule = ReplicationModule;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/replication.module.js.map

/***/ }),

/***/ 619:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var search_trigger_service_1 = __webpack_require__(95);
var session_service_1 = __webpack_require__(14);
var shared_const_1 = __webpack_require__(2);
var ListRepositoryComponent = (function () {
    function ListRepositoryComponent(router, searchTrigger, session) {
        this.router = router;
        this.searchTrigger = searchTrigger;
        this.session = session;
        this.delete = new core_1.EventEmitter();
        this.paginate = new core_1.EventEmitter();
        this.mode = shared_const_1.ListMode.FULL;
        this.pageOffset = 1;
    }
    ListRepositoryComponent.prototype.deleteRepo = function (repoName) {
        this.delete.emit(repoName);
    };
    ListRepositoryComponent.prototype.refresh = function (state) {
        if (this.repositories) {
            this.paginate.emit(state);
        }
    };
    Object.defineProperty(ListRepositoryComponent.prototype, "listFullMode", {
        get: function () {
            return this.mode === shared_const_1.ListMode.FULL;
        },
        enumerable: true,
        configurable: true
    });
    ListRepositoryComponent.prototype.gotoLink = function (projectId, repoName) {
        this.searchTrigger.closeSearch(false);
        var linkUrl = ['harbor', 'tags', projectId, repoName];
        if (!this.session.getCurrentUser()) {
            var navigatorExtra = {
                queryParams: { "redirect_url": linkUrl.join("/") }
            };
            this.router.navigate([shared_const_1.signInRoute], navigatorExtra);
        }
        else {
            this.router.navigate(linkUrl);
        }
    };
    __decorate([
        core_1.Input(), 
        __metadata('design:type', Number)
    ], ListRepositoryComponent.prototype, "projectId", void 0);
    __decorate([
        core_1.Input(), 
        __metadata('design:type', Array)
    ], ListRepositoryComponent.prototype, "repositories", void 0);
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], ListRepositoryComponent.prototype, "delete", void 0);
    __decorate([
        core_1.Input(), 
        __metadata('design:type', Number)
    ], ListRepositoryComponent.prototype, "totalPage", void 0);
    __decorate([
        core_1.Input(), 
        __metadata('design:type', Number)
    ], ListRepositoryComponent.prototype, "totalRecordCount", void 0);
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], ListRepositoryComponent.prototype, "paginate", void 0);
    __decorate([
        core_1.Input(), 
        __metadata('design:type', String)
    ], ListRepositoryComponent.prototype, "mode", void 0);
    ListRepositoryComponent = __decorate([
        core_1.Component({
            selector: 'list-repository',
            template: __webpack_require__(851)
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof router_1.Router !== 'undefined' && router_1.Router) === 'function' && _a) || Object, (typeof (_b = typeof search_trigger_service_1.SearchTriggerService !== 'undefined' && search_trigger_service_1.SearchTriggerService) === 'function' && _b) || Object, (typeof (_c = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _c) || Object])
    ], ListRepositoryComponent);
    return ListRepositoryComponent;
    var _a, _b, _c;
}());
exports.ListRepositoryComponent = ListRepositoryComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/list-repository.component.js.map

/***/ }),

/***/ 620:
/***/ (function(module, exports, __webpack_require__) {

"use strict";
/*
  {
    "id": "2",
    "name": "library/mysql",
    "owner_id": 1,
    "project_id": 1,
    "description": "",
    "pull_count": 0,
    "star_count": 0,
    "tags_count": 1,
    "creation_time": "2017-02-14T09:22:58Z",
    "update_time": "0001-01-01T00:00:00Z"
  }
*/

var Repository = (function () {
    function Repository(name, tags_count) {
        this.name = name;
        this.tags_count = tags_count;
    }
    return Repository;
}());
exports.Repository = Repository;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/repository.js.map

/***/ }),

/***/ 621:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var TagView = (function () {
    function TagView() {
    }
    return TagView;
}());
exports.TagView = TagView;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/tag-view.js.map

/***/ }),

/***/ 622:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var shared_utils_1 = __webpack_require__(33);
var shared_const_1 = __webpack_require__(2);
var message_service_1 = __webpack_require__(10);
var top_repository_service_1 = __webpack_require__(623);
var repository_1 = __webpack_require__(620);
var TopRepoComponent = (function () {
    function TopRepoComponent(topRepoService, msgService) {
        this.topRepoService = topRepoService;
        this.msgService = msgService;
        this.topRepos = [];
    }
    Object.defineProperty(TopRepoComponent.prototype, "listMode", {
        get: function () {
            return shared_const_1.ListMode.READONLY;
        },
        enumerable: true,
        configurable: true
    });
    //Implement ngOnIni
    TopRepoComponent.prototype.ngOnInit = function () {
        this.getTopRepos();
    };
    //Get top popular repositories
    TopRepoComponent.prototype.getTopRepos = function () {
        var _this = this;
        this.topRepoService.getTopRepos()
            .then(function (repos) { return repos.forEach(function (item) {
            var repo = new repository_1.Repository(item.name, item.count);
            repo.pull_count = 0;
            _this.topRepos.push(repo);
        }); })
            .catch(function (error) {
            _this.msgService.announceMessage(error.status, shared_utils_1.errorHandler(error), shared_const_1.AlertType.WARNING);
        });
    };
    TopRepoComponent = __decorate([
        core_1.Component({
            selector: 'top-repo',
            template: __webpack_require__(854),
            providers: [top_repository_service_1.TopRepoService]
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof top_repository_service_1.TopRepoService !== 'undefined' && top_repository_service_1.TopRepoService) === 'function' && _a) || Object, (typeof (_b = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _b) || Object])
    ], TopRepoComponent);
    return TopRepoComponent;
    var _a, _b;
}());
exports.TopRepoComponent = TopRepoComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/top-repo.component.js.map

/***/ }),

/***/ 623:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var http_1 = __webpack_require__(20);
__webpack_require__(58);
exports.topRepoEndpoint = "/api/repositories/top";
/**
 * Declare service to handle the top repositories
 *
 *
 * @export
 * @class GlobalSearchService
 */
var TopRepoService = (function () {
    function TopRepoService(http) {
        this.http = http;
        this.headers = new http_1.Headers({
            "Content-Type": 'application/json'
        });
        this.options = new http_1.RequestOptions({
            headers: this.headers
        });
    }
    /**
     * Get top popular repositories
     *
     * @param {string} keyword
     * @returns {Promise<TopRepo>}
     *
     * @memberOf GlobalSearchService
     */
    TopRepoService.prototype.getTopRepos = function () {
        return this.http.get(exports.topRepoEndpoint, this.options).toPromise()
            .then(function (response) { return response.json(); })
            .catch(function (error) { return Promise.reject(error); });
    };
    TopRepoService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof http_1.Http !== 'undefined' && http_1.Http) === 'function' && _a) || Object])
    ], TopRepoService);
    return TopRepoService;
    var _a;
}());
exports.TopRepoService = TopRepoService;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/top-repository.service.js.map

/***/ }),

/***/ 624:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var CreateEditPolicy = (function () {
    function CreateEditPolicy() {
    }
    return CreateEditPolicy;
}());
exports.CreateEditPolicy = CreateEditPolicy;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/create-edit-policy.js.map

/***/ }),

/***/ 625:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var core_2 = __webpack_require__(34);
var deletion_dialog_service_1 = __webpack_require__(48);
var DeletionDialogComponent = (function () {
    function DeletionDialogComponent(delService, translate) {
        var _this = this;
        this.delService = delService;
        this.translate = translate;
        this.opened = false;
        this.dialogTitle = "";
        this.dialogContent = "";
        this.annouceSubscription = delService.deletionAnnouced$.subscribe(function (msg) {
            _this.dialogTitle = msg.title;
            _this.dialogContent = msg.message;
            _this.message = msg;
            _this.translate.get(_this.dialogTitle).subscribe(function (res) { return _this.dialogTitle = res; });
            _this.translate.get(_this.dialogContent, { 'param': msg.param }).subscribe(function (res) { return _this.dialogContent = res; });
            //Open dialog
            _this.open();
        });
    }
    DeletionDialogComponent.prototype.ngOnDestroy = function () {
        if (this.annouceSubscription) {
            this.annouceSubscription.unsubscribe();
        }
    };
    DeletionDialogComponent.prototype.open = function () {
        this.opened = true;
    };
    DeletionDialogComponent.prototype.close = function () {
        this.opened = false;
    };
    DeletionDialogComponent.prototype.confirm = function () {
        this.delService.confirmDeletion(this.message);
        this.close();
    };
    DeletionDialogComponent = __decorate([
        core_1.Component({
            selector: 'deletion-dialog',
            template: __webpack_require__(857),
            styles: [__webpack_require__(814)]
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof deletion_dialog_service_1.DeletionDialogService !== 'undefined' && deletion_dialog_service_1.DeletionDialogService) === 'function' && _a) || Object, (typeof (_b = typeof core_2.TranslateService !== 'undefined' && core_2.TranslateService) === 'function' && _b) || Object])
    ], DeletionDialogComponent);
    return DeletionDialogComponent;
    var _a, _b;
}());
exports.DeletionDialogComponent = DeletionDialogComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/deletion-dialog.component.js.map

/***/ }),

/***/ 626:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var Subject_1 = __webpack_require__(25);
__webpack_require__(461);
__webpack_require__(462);
var FilterComponent = (function () {
    function FilterComponent() {
        this.placeHolder = "";
        this.currentValue = "";
        this.leadingSpacesAdded = false;
        this.filterTerms = new Subject_1.Subject();
        this.filterEvt = new core_1.EventEmitter();
    }
    Object.defineProperty(FilterComponent.prototype, "flPlaceholder", {
        set: function (placeHolder) {
            this.placeHolder = placeHolder;
        },
        enumerable: true,
        configurable: true
    });
    FilterComponent.prototype.ngOnInit = function () {
        var _this = this;
        this.filterTerms
            .debounceTime(300)
            .distinctUntilChanged()
            .subscribe(function (terms) {
            _this.filterEvt.emit(terms);
        });
    };
    FilterComponent.prototype.valueChange = function () {
        //Send out filter terms
        this.filterTerms.next(this.currentValue.trim());
    };
    __decorate([
        core_1.Output("filter"), 
        __metadata('design:type', Object)
    ], FilterComponent.prototype, "filterEvt", void 0);
    __decorate([
        core_1.Input("filterPlaceholder"), 
        __metadata('design:type', String), 
        __metadata('design:paramtypes', [String])
    ], FilterComponent.prototype, "flPlaceholder", null);
    FilterComponent = __decorate([
        core_1.Component({
            selector: 'grid-filter',
            template: __webpack_require__(858),
            styles: [__webpack_require__(815)]
        }), 
        __metadata('design:paramtypes', [])
    ], FilterComponent);
    return FilterComponent;
}());
exports.FilterComponent = FilterComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/filter.component.js.map

/***/ }),

/***/ 627:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var HarborActionOverflow = (function () {
    function HarborActionOverflow() {
    }
    HarborActionOverflow = __decorate([
        core_1.Component({
            selector: "harbor-action-overflow",
            template: __webpack_require__(859)
        }), 
        __metadata('design:paramtypes', [])
    ], HarborActionOverflow);
    return HarborActionOverflow;
}());
exports.HarborActionOverflow = HarborActionOverflow;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/harbor-action-overflow.js.map

/***/ }),

/***/ 628:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var replication_service_1 = __webpack_require__(79);
var deletion_dialog_service_1 = __webpack_require__(48);
var deletion_message_1 = __webpack_require__(68);
var shared_const_1 = __webpack_require__(2);
var message_service_1 = __webpack_require__(10);
var shared_const_2 = __webpack_require__(2);
var ListPolicyComponent = (function () {
    function ListPolicyComponent(replicationService, deletionDialogService, messageService) {
        var _this = this;
        this.replicationService = replicationService;
        this.deletionDialogService = deletionDialogService;
        this.messageService = messageService;
        this.reload = new core_1.EventEmitter();
        this.selectOne = new core_1.EventEmitter();
        this.editOne = new core_1.EventEmitter();
        this.subscription = this.subscription = this.deletionDialogService
            .deletionConfirm$
            .subscribe(function (message) {
            if (message && message.targetId === shared_const_1.DeletionTargets.POLICY) {
                _this.replicationService
                    .deletePolicy(message.data)
                    .subscribe(function (response) {
                    console.log('Successful delete policy with ID:' + message.data);
                    _this.reload.emit(true);
                }, function (error) { return _this.messageService.announceMessage(error.status, 'Failed to delete policy with ID:' + message.data, shared_const_2.AlertType.DANGER); });
            }
        });
    }
    ListPolicyComponent.prototype.ngOnDestroy = function () {
        if (this.subscription) {
            this.subscription.unsubscribe();
        }
    };
    ListPolicyComponent.prototype.selectPolicy = function (policy) {
        this.selectedId = policy.id;
        console.log('Select policy ID:' + policy.id);
        this.selectOne.emit(policy);
    };
    ListPolicyComponent.prototype.editPolicy = function (policy) {
        console.log('Open modal to edit policy.');
        this.editOne.emit(policy.id);
    };
    ListPolicyComponent.prototype.enablePolicy = function (policy) {
        policy.enabled = policy.enabled === 0 ? 1 : 0;
        console.log('Enable policy ID:' + policy.id + ' with activation status ' + policy.enabled);
        this.replicationService.enablePolicy(policy.id, policy.enabled);
    };
    ListPolicyComponent.prototype.deletePolicy = function (policy) {
        var deletionMessage = new deletion_message_1.DeletionMessage('REPLICATION.DELETION_TITLE', 'REPLICATION.DELETION_SUMMARY', policy.name, policy.id, shared_const_1.DeletionTargets.POLICY);
        this.deletionDialogService.openComfirmDialog(deletionMessage);
    };
    __decorate([
        core_1.Input(), 
        __metadata('design:type', Array)
    ], ListPolicyComponent.prototype, "policies", void 0);
    __decorate([
        core_1.Input(), 
        __metadata('design:type', Boolean)
    ], ListPolicyComponent.prototype, "projectless", void 0);
    __decorate([
        core_1.Input(), 
        __metadata('design:type', Number)
    ], ListPolicyComponent.prototype, "selectedId", void 0);
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], ListPolicyComponent.prototype, "reload", void 0);
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], ListPolicyComponent.prototype, "selectOne", void 0);
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], ListPolicyComponent.prototype, "editOne", void 0);
    ListPolicyComponent = __decorate([
        core_1.Component({
            selector: 'list-policy',
            template: __webpack_require__(861),
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof replication_service_1.ReplicationService !== 'undefined' && replication_service_1.ReplicationService) === 'function' && _a) || Object, (typeof (_b = typeof deletion_dialog_service_1.DeletionDialogService !== 'undefined' && deletion_dialog_service_1.DeletionDialogService) === 'function' && _b) || Object, (typeof (_c = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _c) || Object])
    ], ListPolicyComponent);
    return ListPolicyComponent;
    var _a, _b, _c;
}());
exports.ListPolicyComponent = ListPolicyComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/list-policy.component.js.map

/***/ }),

/***/ 629:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var forms_1 = __webpack_require__(26);
exports.assiiChars = /[\u4e00-\u9fa5]/;
function maxLengthExtValidator(length) {
    return function (control) {
        var value = control.value;
        if (!value || value.trim() === "") {
            return null;
        }
        var regExp = new RegExp(exports.assiiChars, 'i');
        var count = 0;
        var len = value.length;
        for (var i = 0; i < len; i++) {
            if (regExp.test(value[i])) {
                count += 3;
            }
            else {
                count++;
            }
        }
        return count > length ? { 'maxLengthExt': count } : null;
    };
}
exports.maxLengthExtValidator = maxLengthExtValidator;
var MaxLengthExtValidatorDirective = (function () {
    function MaxLengthExtValidatorDirective() {
        this.valFn = forms_1.Validators.nullValidator;
    }
    MaxLengthExtValidatorDirective.prototype.ngOnChanges = function (changes) {
        var change = changes['maxLengthExt'];
        if (change) {
            var val = change.currentValue;
            this.valFn = maxLengthExtValidator(val);
        }
        else {
            this.valFn = forms_1.Validators.nullValidator;
        }
    };
    MaxLengthExtValidatorDirective.prototype.validate = function (control) {
        return this.valFn(control);
    };
    __decorate([
        core_1.Input(), 
        __metadata('design:type', Number)
    ], MaxLengthExtValidatorDirective.prototype, "maxLengthExt", void 0);
    MaxLengthExtValidatorDirective = __decorate([
        core_1.Directive({
            selector: '[maxLengthExt]',
            providers: [{ provide: forms_1.NG_VALIDATORS, useExisting: MaxLengthExtValidatorDirective, multi: true }]
        }), 
        __metadata('design:paramtypes', [])
    ], MaxLengthExtValidatorDirective);
    return MaxLengthExtValidatorDirective;
}());
exports.MaxLengthExtValidatorDirective = MaxLengthExtValidatorDirective;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/max-length-ext.directive.js.map

/***/ }),

/***/ 630:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var forms_1 = __webpack_require__(26);
exports.portNumbers = /[\d]+/;
function portValidator() {
    return function (control) {
        var value = control.value;
        if (!value) {
            return { 'port': 65535 };
        }
        var regExp = new RegExp(exports.portNumbers, 'i');
        if (!regExp.test(value)) {
            return { 'port': 65535 };
        }
        else {
            var portV = parseInt(value);
            if (portV <= 0 || portV > 65535) {
                return { 'port': 65535 };
            }
        }
        return null;
    };
}
exports.portValidator = portValidator;
var PortValidatorDirective = (function () {
    function PortValidatorDirective() {
        this.valFn = portValidator();
    }
    PortValidatorDirective.prototype.validate = function (control) {
        return this.valFn(control);
    };
    PortValidatorDirective = __decorate([
        core_1.Directive({
            selector: '[port]',
            providers: [{ provide: forms_1.NG_VALIDATORS, useExisting: PortValidatorDirective, multi: true }]
        }), 
        __metadata('design:paramtypes', [])
    ], PortValidatorDirective);
    return PortValidatorDirective;
}());
exports.PortValidatorDirective = PortValidatorDirective;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/port.directive.js.map

/***/ }),

/***/ 631:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var router_1 = __webpack_require__(8);
var session_service_1 = __webpack_require__(14);
var shared_const_1 = __webpack_require__(2);
var BaseRoutingResolver = (function () {
    function BaseRoutingResolver(session, router) {
        this.session = session;
        this.router = router;
    }
    BaseRoutingResolver.prototype.resolve = function (route, state) {
        var _this = this;
        //To refresh seesion
        return this.session.retrieveUser()
            .then(function (sessionUser) {
            return sessionUser;
        })
            .catch(function (error) {
            //Session retrieving failed then redirect to sign-in
            //no matter what status code is.
            //Please pay attention that route 'harborRootRoute' support anonymous user
            if (state.url != shared_const_1.harborRootRoute) {
                var navigatorExtra = {
                    queryParams: { "redirect_url": state.url }
                };
                _this.router.navigate(['sign-in'], navigatorExtra);
            }
        });
    };
    BaseRoutingResolver = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _a) || Object, (typeof (_b = typeof router_1.Router !== 'undefined' && router_1.Router) === 'function' && _b) || Object])
    ], BaseRoutingResolver);
    return BaseRoutingResolver;
    var _a, _b;
}());
exports.BaseRoutingResolver = BaseRoutingResolver;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/base-routing-resolver.service.js.map

/***/ }),

/***/ 632:
/***/ (function(module, exports, __webpack_require__) {

"use strict";
/**
 * Declare class for store the sign in data,
 * two prperties:
 *   principal: The username used to sign in
 *   password: The password used to sign in
 *
 * @export
 * @class SignInCredential
 */

var SignInCredential = (function () {
    function SignInCredential() {
    }
    return SignInCredential;
}());
exports.SignInCredential = SignInCredential;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/sign-in-credential.js.map

/***/ }),

/***/ 633:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var statistics_service_1 = __webpack_require__(635);
var shared_utils_1 = __webpack_require__(33);
var shared_const_1 = __webpack_require__(2);
var message_service_1 = __webpack_require__(10);
var statistics_1 = __webpack_require__(636);
var session_service_1 = __webpack_require__(14);
var StatisticsPanelComponent = (function () {
    function StatisticsPanelComponent(statistics, msgService, session) {
        this.statistics = statistics;
        this.msgService = msgService;
        this.session = session;
        this.originalCopy = new statistics_1.Statistics();
    }
    StatisticsPanelComponent.prototype.ngOnInit = function () {
        if (this.session.getCurrentUser()) {
            this.getStatistics();
        }
    };
    StatisticsPanelComponent.prototype.getStatistics = function () {
        var _this = this;
        this.statistics.getStatistics()
            .then(function (statistics) { return _this.originalCopy = statistics; })
            .catch(function (error) {
            _this.msgService.announceMessage(error.status, shared_utils_1.errorHandler(error), shared_const_1.AlertType.WARNING);
        });
    };
    Object.defineProperty(StatisticsPanelComponent.prototype, "isValidSession", {
        get: function () {
            var user = this.session.getCurrentUser();
            return user && user.has_admin_role > 0;
        },
        enumerable: true,
        configurable: true
    });
    StatisticsPanelComponent = __decorate([
        core_1.Component({
            selector: 'statistics-panel',
            template: __webpack_require__(864),
            styles: [__webpack_require__(458)],
            providers: [statistics_service_1.StatisticsService]
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof statistics_service_1.StatisticsService !== 'undefined' && statistics_service_1.StatisticsService) === 'function' && _a) || Object, (typeof (_b = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _b) || Object, (typeof (_c = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _c) || Object])
    ], StatisticsPanelComponent);
    return StatisticsPanelComponent;
    var _a, _b, _c;
}());
exports.StatisticsPanelComponent = StatisticsPanelComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/statistics-panel.component.js.map

/***/ }),

/***/ 634:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var StatisticsComponent = (function () {
    function StatisticsComponent() {
    }
    __decorate([
        core_1.Input(), 
        __metadata('design:type', Object)
    ], StatisticsComponent.prototype, "data", void 0);
    StatisticsComponent = __decorate([
        core_1.Component({
            selector: 'statistics',
            template: __webpack_require__(865),
            styles: [__webpack_require__(458)]
        }), 
        __metadata('design:paramtypes', [])
    ], StatisticsComponent);
    return StatisticsComponent;
}());
exports.StatisticsComponent = StatisticsComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/statistics.component.js.map

/***/ }),

/***/ 635:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var http_1 = __webpack_require__(20);
__webpack_require__(58);
exports.statisticsEndpoint = "/api/statistics";
/**
 * Declare service to handle the top repositories
 *
 *
 * @export
 * @class GlobalSearchService
 */
var StatisticsService = (function () {
    function StatisticsService(http) {
        this.http = http;
        this.headers = new http_1.Headers({
            "Content-Type": 'application/json'
        });
        this.options = new http_1.RequestOptions({
            headers: this.headers
        });
    }
    StatisticsService.prototype.getStatistics = function () {
        return this.http.get(exports.statisticsEndpoint, this.options).toPromise()
            .then(function (response) { return response.json(); })
            .catch(function (error) { return Promise.reject(error); });
    };
    StatisticsService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof http_1.Http !== 'undefined' && http_1.Http) === 'function' && _a) || Object])
    ], StatisticsService);
    return StatisticsService;
    var _a;
}());
exports.StatisticsService = StatisticsService;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/statistics.service.js.map

/***/ }),

/***/ 636:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var Statistics = (function () {
    function Statistics() {
    }
    return Statistics;
}());
exports.Statistics = Statistics;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/statistics.js.map

/***/ }),

/***/ 637:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var shared_module_1 = __webpack_require__(52);
var user_component_1 = __webpack_require__(413);
var new_user_modal_component_1 = __webpack_require__(412);
var user_service_1 = __webpack_require__(175);
var UserModule = (function () {
    function UserModule() {
    }
    UserModule = __decorate([
        core_1.NgModule({
            imports: [
                shared_module_1.SharedModule
            ],
            declarations: [
                user_component_1.UserComponent,
                new_user_modal_component_1.NewUserModalComponent
            ],
            exports: [
                user_component_1.UserComponent
            ],
            providers: [user_service_1.UserService]
        }), 
        __metadata('design:paramtypes', [])
    ], UserModule);
    return UserModule;
}());
exports.UserModule = UserModule;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/user.module.js.map

/***/ }),

/***/ 638:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

/**
 * For user management
 *
 * @export
 * @class User
 */
var User = (function () {
    function User() {
    }
    return User;
}());
exports.User = User;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/user.js.map

/***/ }),

/***/ 639:
/***/ (function(module, exports, __webpack_require__) {

"use strict";
// The file contents for the current environment will overwrite these during build.
// The build system defaults to the dev environment which uses `environment.ts`, but if you do
// `ng build --env=prod` then `environment.prod.ts` will be used instead.
// The list of which env maps to which file can be found in `angular-cli.json`.

exports.environment = {
    production: false
};
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/environment.js.map

/***/ }),

/***/ 640:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

// This file includes polyfills needed by Angular 2 and is loaded before
// the app. You can add your own extra polyfills to this file.
__webpack_require__(657);
__webpack_require__(650);
__webpack_require__(646);
__webpack_require__(652);
__webpack_require__(651);
__webpack_require__(649);
__webpack_require__(648);
__webpack_require__(656);
__webpack_require__(645);
__webpack_require__(644);
__webpack_require__(654);
__webpack_require__(647);
__webpack_require__(655);
__webpack_require__(653);
__webpack_require__(658);
__webpack_require__(902);
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/polyfills.js.map

/***/ }),

/***/ 68:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var shared_const_1 = __webpack_require__(2);
var DeletionMessage = (function () {
    function DeletionMessage(title, message, param, data, targetId) {
        this.targetId = shared_const_1.DeletionTargets.EMPTY;
        this.title = title;
        this.message = message;
        this.data = data;
        this.targetId = targetId;
        this.param = param;
    }
    return DeletionMessage;
}());
exports.DeletionMessage = DeletionMessage;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/deletion-message.js.map

/***/ }),

/***/ 79:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var http_1 = __webpack_require__(20);
var base_service_1 = __webpack_require__(251);
var Observable_1 = __webpack_require__(3);
__webpack_require__(133);
__webpack_require__(85);
__webpack_require__(132);
__webpack_require__(463);
var ReplicationService = (function (_super) {
    __extends(ReplicationService, _super);
    function ReplicationService(http) {
        _super.call(this);
        this.http = http;
    }
    ReplicationService.prototype.listPolicies = function (policyName, projectId) {
        if (!projectId) {
            projectId = '';
        }
        console.log('Get policies with project ID:' + projectId + ', policy name:' + policyName);
        return this.http
            .get("/api/policies/replication?project_id=" + projectId + "&name=" + policyName)
            .map(function (response) { return response.json(); })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    ReplicationService.prototype.getPolicy = function (policyId) {
        console.log('Get policy with ID:' + policyId);
        return this.http
            .get("/api/policies/replication/" + policyId)
            .map(function (response) { return response.json(); })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    ReplicationService.prototype.createPolicy = function (policy) {
        console.log('Create policy with project ID:' + policy.project_id + ', policy:' + JSON.stringify(policy));
        return this.http
            .post("/api/policies/replication", JSON.stringify(policy))
            .map(function (response) { return response.status; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    ReplicationService.prototype.updatePolicy = function (policy) {
        if (policy && policy.id) {
            return this.http
                .put("/api/policies/replication/" + policy.id, JSON.stringify(policy))
                .map(function (response) { return response.status; })
                .catch(function (error) { return Observable_1.Observable.throw(error); });
        }
        return Observable_1.Observable.throw(new Error("Policy is nil or has no ID set."));
    };
    ReplicationService.prototype.createOrUpdatePolicyWithNewTarget = function (policy, target) {
        var _this = this;
        return this.http
            .post("/api/targets", JSON.stringify(target))
            .map(function (response) {
            return response.status;
        })
            .catch(function (error) { return Observable_1.Observable.throw(error); })
            .flatMap(function (status) {
            if (status === 201) {
                return _this.http
                    .get("/api/targets?name=" + target.name)
                    .map(function (res) { return res; })
                    .catch(function (error) { return Observable_1.Observable.throw(error); });
            }
        })
            .flatMap(function (res) {
            if (res.status === 200) {
                var lastAddedTarget = res.json()[0];
                if (lastAddedTarget && lastAddedTarget.id) {
                    policy.target_id = lastAddedTarget.id;
                    if (policy.id) {
                        return _this.http
                            .put("/api/policies/replication/" + policy.id, JSON.stringify(policy))
                            .map(function (response) { return response.status; })
                            .catch(function (error) { return Observable_1.Observable.throw(error); });
                    }
                    else {
                        return _this.http
                            .post("/api/policies/replication", JSON.stringify(policy))
                            .map(function (response) { return response.status; })
                            .catch(function (error) { return Observable_1.Observable.throw(error); });
                    }
                }
            }
        })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    ReplicationService.prototype.enablePolicy = function (policyId, enabled) {
        console.log('Enable or disable policy ID:' + policyId + ' with activation status:' + enabled);
        return this.http
            .put("/api/policies/replication/" + policyId + "/enablement", { enabled: enabled })
            .map(function (response) { return response.status; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    ReplicationService.prototype.deletePolicy = function (policyId) {
        console.log('Delete policy ID:' + policyId);
        return this.http
            .delete("/api/policies/replication/" + policyId)
            .map(function (response) { return response.status; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    // /api/jobs/replication/?page=1&page_size=20&end_time=&policy_id=1&start_time=&status=&repository=
    ReplicationService.prototype.listJobs = function (policyId, status, repoName, startTime, endTime, page, pageSize) {
        if (status === void 0) { status = ''; }
        if (repoName === void 0) { repoName = ''; }
        if (startTime === void 0) { startTime = ''; }
        if (endTime === void 0) { endTime = ''; }
        console.log('Get jobs under policy ID:' + policyId);
        return this.http
            .get("/api/jobs/replication?policy_id=" + policyId + "&status=" + status + "&repository=" + repoName + "&start_time=" + startTime + "&end_time=" + endTime + "&page=" + page + "&page_size=" + pageSize)
            .map(function (response) { return response; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    ReplicationService.prototype.listTargets = function (targetName) {
        console.log('Get targets.');
        return this.http
            .get("/api/targets?name=" + targetName)
            .map(function (response) { return response.json(); })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    ReplicationService.prototype.getTarget = function (targetId) {
        console.log('Get target by ID:' + targetId);
        return this.http
            .get("/api/targets/" + targetId)
            .map(function (response) { return response.json(); })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    ReplicationService.prototype.createTarget = function (target) {
        console.log('Create target:' + JSON.stringify(target));
        return this.http
            .post("/api/targets", JSON.stringify(target))
            .map(function (response) { return response.status; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    ReplicationService.prototype.pingTarget = function (target) {
        console.log('Ping target.');
        var body = new http_1.URLSearchParams();
        body.set('endpoint', target.endpoint);
        body.set('username', target.username);
        body.set('password', target.password);
        return this.http
            .post("/api/targets/ping", body)
            .map(function (response) { return response.status; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    ReplicationService.prototype.updateTarget = function (target) {
        console.log('Update target with target ID' + target.id);
        return this.http
            .put("/api/targets/" + target.id, JSON.stringify(target))
            .map(function (response) { return response.status; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    ReplicationService.prototype.deleteTarget = function (targetId) {
        console.log('Deleting  target with ID:' + targetId);
        return this.http
            .delete("/api/targets/" + targetId)
            .map(function (response) { return response.status; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    ReplicationService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof http_1.Http !== 'undefined' && http_1.Http) === 'function' && _a) || Object])
    ], ReplicationService);
    return ReplicationService;
    var _a;
}(base_service_1.BaseService));
exports.ReplicationService = ReplicationService;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/replication.service.js.map

/***/ }),

/***/ 80:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var core_2 = __webpack_require__(34);
var shared_utils_1 = __webpack_require__(33);
var InlineAlertComponent = (function () {
    function InlineAlertComponent(translate) {
        this.translate = translate;
        this.inlineAlertType = 'alert-danger';
        this.inlineAlertClosable = true;
        this.alertClose = true;
        this.displayedText = "";
        this.showCancelAction = false;
        this.useAppLevelStyle = false;
        this.confirmEvt = new core_1.EventEmitter();
    }
    Object.defineProperty(InlineAlertComponent.prototype, "errorMessage", {
        get: function () {
            return this.displayedText;
        },
        enumerable: true,
        configurable: true
    });
    //Show error message inline
    InlineAlertComponent.prototype.showInlineError = function (error) {
        this.displayedText = shared_utils_1.errorHandler(error);
        this.inlineAlertType = 'alert-danger';
        this.showCancelAction = false;
        this.inlineAlertClosable = true;
        this.alertClose = false;
        this.useAppLevelStyle = false;
    };
    //Show confirmation info with action button
    InlineAlertComponent.prototype.showInlineConfirmation = function (warning) {
        var _this = this;
        this.displayedText = "";
        if (warning && warning.message) {
            this.translate.get(warning.message).subscribe(function (res) { return _this.displayedText = res; });
        }
        this.inlineAlertType = 'alert-warning';
        this.showCancelAction = true;
        this.inlineAlertClosable = true;
        this.alertClose = false;
        this.useAppLevelStyle = true;
    };
    //Show inline sccess info
    InlineAlertComponent.prototype.showInlineSuccess = function (info) {
        var _this = this;
        this.displayedText = "";
        if (info && info.message) {
            this.translate.get(info.message).subscribe(function (res) { return _this.displayedText = res; });
        }
        this.inlineAlertType = 'alert-success';
        this.showCancelAction = false;
        this.inlineAlertClosable = true;
        this.alertClose = false;
        this.useAppLevelStyle = false;
    };
    //Close alert
    InlineAlertComponent.prototype.close = function () {
        this.alertClose = true;
    };
    InlineAlertComponent.prototype.confirmCancel = function () {
        this.confirmEvt.emit(true);
    };
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], InlineAlertComponent.prototype, "confirmEvt", void 0);
    InlineAlertComponent = __decorate([
        core_1.Component({
            selector: 'inline-alert',
            template: __webpack_require__(860)
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof core_2.TranslateService !== 'undefined' && core_2.TranslateService) === 'function' && _a) || Object])
    ], InlineAlertComponent);
    return InlineAlertComponent;
    var _a;
}());
exports.InlineAlertComponent = InlineAlertComponent;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/inline-alert.component.js.map

/***/ }),

/***/ 802:
/***/ (function(module, exports) {

module.exports = ".progress-size-small {\n    height: 0.5em !important;\n}\n\n.visibility-hidden {\n    visibility: hidden;\n}\n\n.forgot-password-link {\n    position: relative;\n    line-height: 36px;\n    font-size: 14px;\n    float: right;\n    top: -5px;\n}"

/***/ }),

/***/ 803:
/***/ (function(module, exports) {

module.exports = ".search-overlay {\n    display: block;\n    position: absolute;\n    height: 100%;\n    width: 98%;\n    /*shoud be lesser than 1000 to aoivd override the popup menu*/\n    z-index: 999;\n    box-sizing: border-box;\n    background: #fafafa;\n    top: 0px;\n    padding-left: 24px;\n}\n\n.search-header {\n    display: inline-block;\n    width: 100%;\n    position: relative;\n}\n\n.search-title {\n    font-size: 28px;\n    letter-spacing: normal;\n    color: #000;\n}\n\n.search-close {\n    position: absolute;\n    right: 24px;\n    cursor: pointer;\n}\n\n.search-parent-override {\n    position: relative !important;\n}\n\n.search-spinner {\n    top: 50%;\n    left: 50%;\n    position: absolute;\n}\n\n.grid-header-wrapper {\n    text-align: right;\n}\n\n.grid-filter {\n    position: relative;\n    top: 8px;\n    margin: 0px auto 0px auto;\n}"

/***/ }),

/***/ 804:
/***/ (function(module, exports) {

module.exports = ".side-nav-override {\n    box-shadow: 6px 0px 0px 0px #ccc;\n}\n\n.container-override {\n    position: relative !important;\n}\n\n.start-content-padding {\n    padding-top: 0px !important;\n    padding-bottom: 0px !important;\n    padding-left: 0px !important;\n}"

/***/ }),

/***/ 805:
/***/ (function(module, exports) {

module.exports = ".sign-in-override {\n    padding-left: 0px !important;\n    padding-right: 5px !important;\n}\n\n.sign-up-override {\n    padding-left: 5px !important;\n}\n\n.custom-divider {\n    display: inline-block;\n    border-right: 2px inset snow;\n    padding: 2px 0px 2px 0px;\n    vertical-align: middle;\n    height: 24px;\n}\n\n.lang-selected {\n    font-weight: bold;\n}\n\n.nav-divider {\n    display: inline-block;\n    width: 1px;\n    height: 40px;\n    background-color: #fafafa;\n    position: relative;\n    top: 10px;\n}"

/***/ }),

/***/ 806:
/***/ (function(module, exports) {

module.exports = ".start-card {\n    border-right: 1px solid #cccccc;\n    padding: 24px;\n    background-color: white;\n    height: 100%;\n}\n\n.row-fill-height {\n    height: 100%;\n}\n\n.row-margin {\n    margin-left: 24px;\n}\n\n.column-fill-height {\n    height: 100%;\n}\n\n.my-card-img {\n    background-image: url('../../../images/harbor-logo.png');\n    background-repeat: no-repeat;\n    background-size: contain;\n    height: 160px;\n}\n\n.my-card-footer {\n    float: right;\n    margin-top: 100px;\n}"

/***/ }),

/***/ 807:
/***/ (function(module, exports) {

module.exports = ".advance-option {\n  font-size: 12px;\n}\n"

/***/ }),

/***/ 808:
/***/ (function(module, exports) {

module.exports = ".h2-log-override {\n    margin-top: 0px !important;\n}\n\n.filter-log {\n    float: right;\n    margin-right: 24px;\n    position: relative;\n    top: 8px;\n}\n\n.action-head-pos {\n    position: relative;\n    top: 20px;\n}\n\n.refresh-btn {\n    position: absolute;\n    right: -4px;\n    top: 8px;\n    cursor: pointer;\n}\n\n.custom-lines-button {\n    padding: 0px !important;\n    min-width: 25px !important;\n}\n\n.lines-button-toggole {\n    font-size: 16px;\n    text-decoration: underline;\n}"

/***/ }),

/***/ 809:
/***/ (function(module, exports) {

module.exports = ""

/***/ }),

/***/ 810:
/***/ (function(module, exports) {

module.exports = ".display-in-line {\n    display: inline-block;\n}\n\n.project-title {\n    margin-left: 10px; \n}\n\n.pull-right {\n    float: right !important;\n}"

/***/ }),

/***/ 811:
/***/ (function(module, exports) {

module.exports = ""

/***/ }),

/***/ 812:
/***/ (function(module, exports) {

module.exports = ".custom-h2 {\n    margin-top: 0px !important;\n}\n\n.custom-add-button {\n    font-size: medium;\n    margin-left: -12px;\n}\n\n.filter-icon {\n    position: relative;\n    right: -12px;\n}\n\n.filter-pos {\n    float: right;\n    margin-right: 24px;\n    position: relative;\n    top: 8px;\n}\n\n.action-panel-pos {\n    position: relative;\n    top: 20px;\n}\n\n.refresh-btn {\n    position: absolute;\n    right: -4px;\n    top: 8px;\n    cursor: pointer;\n}"

/***/ }),

/***/ 813:
/***/ (function(module, exports) {

module.exports = ".margin-left-override {\n    margin-left: 24px !important;\n}\n\n.about-text-link {\n    font-family: \"Proxima Nova Light\";\n    font-size: 14px;\n    color: #007CBB;\n    line-height: 24px;\n}\n\n.about-copyright-text {\n    font-family: \"Proxima Nova Light\";\n    font-size: 13px;\n    color: #565656;\n    line-height: 16px;\n}\n\n.about-product-title {\n    font-family: \"Metropolis Light\";\n    font-size: 28px;\n    color: #000000;\n    line-height: 36px;\n}\n\n.about-version {\n    font-family: \"Metropolis\";\n    font-size: 14px;\n    color: #565656;\n    font-weight: 500;\n}\n\n.about-build {\n    font-family: \"Metropolis\";\n    font-size: 14px;\n    color: #565656;\n}"

/***/ }),

/***/ 814:
/***/ (function(module, exports) {

module.exports = ".deletion-icon-inline {\n    display: inline-block;\n}\n\n.deletion-title {\n    line-height: 24px;\n    color: #000000;\n    font-size: 22px;\n}\n\n.deletion-content {\n    font-size: 14px;\n    color: #565656;\n    line-height: 24px;\n    display: inline-block;\n    vertical-align: middle;\n    width: 80%;\n}"

/***/ }),

/***/ 815:
/***/ (function(module, exports) {

module.exports = ".filter-icon {\n    position: relative;\n    right: -12px;\n}"

/***/ }),

/***/ 816:
/***/ (function(module, exports) {

module.exports = ".label-info {\n    margin: 0px !important;\n    padding: 0px !important;\n    margin-top: -5px !important;\n}"

/***/ }),

/***/ 817:
/***/ (function(module, exports) {

module.exports = ".wrapper-back {\n    position: absolute;\n    top: 50%;\n    height: 240px;\n    margin-top: -120px;\n    text-align: center;\n    left: 50%;\n    margin-left: -300px;\n}\n\n.status-code {\n    font-weight: bolder;\n    font-size: 4em;\n    color: #A32100;\n    vertical-align: middle;\n}\n\n.status-text {\n    font-weight: bold;\n    font-size: 3em;\n    margin-left: 10px;\n    vertical-align: middle;\n}\n\n.status-subtitle {\n    font-size: 18px;\n}\n\n.second-number {\n    font-weight: bold;\n    font-size: 2em;\n    color: #EB8D00;\n}"

/***/ }),

/***/ 818:
/***/ (function(module, exports) {

module.exports = ".custom-h2 {\n    margin-top: 0px !important;\n}\n\n.custom-add-button {\n    font-size: medium;\n    margin-left: -12px;\n}\n\n.filter-icon {\n    position: relative;\n    right: -12px;\n}\n\n.filter-pos {\n    float: right;\n    margin-right: 24px;\n    position: relative;\n    top: 8px;\n}\n\n.action-panel-pos {\n    position: relative;\n    top: 20px;\n}\n\n.refresh-btn {\n    position: absolute;\n    right: -4px;\n    top: 8px;\n    cursor: pointer;\n}"

/***/ }),

/***/ 820:
/***/ (function(module, exports) {

module.exports = "<clr-modal [(clrModalOpen)]=\"opened\" [clrModalStaticBackdrop]=\"staticBackdrop\">\n    <h3 class=\"modal-title\">{{'PROFILE.TITLE' | translate}}</h3>\n    <div class=\"modal-body\" style=\"overflow-y: hidden;\">\n        <form #accountSettingsFrom=\"ngForm\" class=\"form\">\n            <section class=\"form-block\">\n                <div class=\"form-group\">\n                    <label for=\"account_settings_username\" class=\"col-md-4\">{{'PROFILE.USER_NAME' | translate}}</label>\n                    <input type=\"text\" name=\"account_settings_username\" [(ngModel)]=\"account.username\" disabled id=\"account_settings_username\" size=\"31\">\n                </div>\n                <div class=\"form-group\">\n                    <label for=\"account_settings_email\" class=\"col-md-4 required\">{{'PROFILE.EMAIL' | translate}}</label>\n                    <label for=\"account_settings_email\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-left\" [class.invalid]=\"eamilInput.invalid && (eamilInput.dirty || eamilInput.touched)\">\n                      <input name=\"account_settings_email\" type=\"text\" #eamilInput=\"ngModel\" [(ngModel)]=\"account.email\" \n                      required \n                      pattern='^[a-zA-Z0-9.!#$%&’*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\\.[a-zA-Z0-9-]+)*$' id=\"account_settings_email\" size=\"28\">\n                      <span class=\"tooltip-content\">\n                          {{'TOOLTIP.EMAIL' | translate}}\n                      </span>\n                    </label>\n                </div>\n                <div class=\"form-group\">\n                    <label for=\"account_settings_full_name\" class=\"col-md-4 required\">{{'PROFILE.FULL_NAME' | translate}}</label>\n                    <label for=\"account_settings_full_name\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-left\" [class.invalid]=\"fullNameInput.invalid && (fullNameInput.dirty || fullNameInput.touched)\">\n                      <input type=\"text\" name=\"account_settings_full_name\" #fullNameInput=\"ngModel\" [(ngModel)]=\"account.realname\" required maxLengthExt=\"20\" id=\"account_settings_full_name\" size=\"28\">\n                      <span class=\"tooltip-content\">\n                          {{'TOOLTIP.FULL_NAME' | translate}}\n                      </span>\n                    </label>\n                </div>\n                <div class=\"form-group\">\n                    <label for=\"account_settings_comments\" class=\"col-md-4\">{{'PROFILE.COMMENT' | translate}}</label>\n                    <label for=\"account_settings_comments\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-left\" [class.invalid]=\"commentInput.invalid && (commentInput.dirty || commentInput.touched)\">\n                    <input type=\"text\" #commentInput=\"ngModel\" maxLengthExt=\"20\" name=\"account_settings_comments\" [(ngModel)]=\"account.comment\" id=\"account_settings_comments\" size=\"28\">\n                    <span class=\"tooltip-content\">\n                        {{'TOOLTIP.COMMENT' | translate}}\n                    </span>\n                    </label>\n                </div>\n            </section>\n        </form>\n        <inline-alert (confirmEvt)=\"confirmCancel($event)\"></inline-alert>\n    </div>\n    <div class=\"modal-footer\">\n        <span class=\"spinner spinner-inline\" style=\"top:8px;\" [hidden]=\"showProgress === false\"></span>\n        <button type=\"button\" class=\"btn btn-outline\" (click)=\"close()\">{{'BUTTON.CANCEL' | translate}}</button>\n        <button type=\"button\" class=\"btn btn-primary\" [disabled]=\"!isValid || showProgress\" (click)=\"submit()\">{{'BUTTON.OK' | translate}}</button>\n    </div>\n</clr-modal>"

/***/ }),

/***/ 821:
/***/ (function(module, exports) {

module.exports = "<clr-modal [(clrModalOpen)]=\"opened\" [clrModalStaticBackdrop]=\"true\">\n    <h3 class=\"modal-title\">{{'RESET_PWD.TITLE' | translate}}</h3>\n    <label class=\"modal-title reset-modal-title-override\">{{'RESET_PWD.CAPTION' | translate}}</label>\n    <div class=\"modal-body\" style=\"overflow-y: hidden;\">\n        <form #forgotPasswordFrom=\"ngForm\" class=\"form\">\n            <section class=\"form-block\">\n                <div class=\"form-group\">\n                    <label for=\"reset_pwd_email\" class=\"col-md-4 required\">{{'RESET_PWD.EMAIL' | translate}}</label>\n                    <label for=\"reset_pwd_email\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-bottom-left\" [class.invalid]=\"validationState === false\">\n                      <input name=\"reset_pwd_email\" type=\"text\" #eamilInput=\"ngModel\" [(ngModel)]=\"email\" \n                      required \n                      pattern='^[a-zA-Z0-9.!#$%&’*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\\.[a-zA-Z0-9-]+)*$' \n                      id=\"reset_pwd_email\" \n                      size=\"36\"\n                      (input)=\"handleValidation(true)\" \n                      (focusout)=\"handleValidation(false)\">\n                      <span class=\"tooltip-content\">\n                          {{'TOOLTIP.EMAIL' | translate}}\n                      </span>\n                    </label>\n                </div>\n            </section>\n        </form>\n        <inline-alert></inline-alert>\n        <div style=\"height: 30px;\"></div>\n    </div>\n    <div class=\"modal-footer\">\n        <span class=\"spinner spinner-inline\" style=\"top:8px;\" [hidden]=\"showProgress === false\"></span>\n        <button type=\"button\" class=\"btn btn-outline\" (click)=\"close()\">{{'BUTTON.CANCEL' | translate}}</button>\n        <button type=\"button\" class=\"btn btn-primary\" [disabled]=\"!isValid || showProgress\" (click)=\"send()\">{{'BUTTON.SEND' | translate}}</button>\n    </div>\n</clr-modal>"

/***/ }),

/***/ 822:
/***/ (function(module, exports) {

module.exports = "<clr-modal [(clrModalOpen)]=\"opened\" [clrModalStaticBackdrop]=\"true\">\n    <h3 class=\"modal-title\">{{'CHANGE_PWD.TITLE' | translate}}</h3>\n    <div class=\"modal-body\" style=\"min-height: 250px; overflow-y: hidden;\">\n        <form #changepwdForm=\"ngForm\" class=\"form\">\n            <section class=\"form-block\">\n                <div class=\"form-group\">\n                    <label for=\"oldPassword\">{{'CHANGE_PWD.CURRENT_PWD' | translate}}</label>\n                    <label for=\"oldPassword\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-bottom-left\" [class.invalid]=\"oldPassInput.invalid && (oldPassInput.dirty || oldPassInput.touched)\">\n                <input type=\"password\" id=\"oldPassword\" placeholder='{{\"PLACEHOLDER.CURRENT_PWD\" | translate}}'\n                    required\n                    name=\"oldPassword\"\n                    [(ngModel)]=\"oldPwd\"\n                    #oldPassInput=\"ngModel\" size=\"25\">\n                <span class=\"tooltip-content\">\n                    {{'TOOLTIP.CURRENT_PWD' | translate}}\n                </span>\n            </label>\n                </div>\n                <div class=\"form-group\">\n                    <label for=\"newPassword\">{{'CHANGE_PWD.NEW_PWD' | translate}}</label>\n                    <label for=\"newPassword\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-left\" [class.invalid]=\"newPassInput.invalid && (newPassInput.dirty || newPassInput.touched)\">\n                <input type=\"password\" id=\"newPassword\" placeholder='{{\"PLACEHOLDER.NEW_PWD\" | translate}}'\n                    required\n                    pattern=\"^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)[a-zA-Z\\d]{7,}$\"\n                    name=\"newPassword\"\n                    [(ngModel)]=\"newPwd\"\n                    #newPassInput=\"ngModel\" size=\"25\">\n                <span class=\"tooltip-content\">\n                    {{'TOOLTIP.PASSWORD' | translate}}\n                </span>\n            </label>\n                </div>\n                <div class=\"form-group\">\n                    <label for=\"reNewPassword\">{{'CHANGE_PWD.CONFIRM_PWD' | translate}}</label>\n                    <label for=\"reNewPassword\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-left\" [class.invalid]=\"(reNewPassInput.invalid && (reNewPassInput.dirty || reNewPassInput.touched)) || (!newPassInput.invalid && reNewPassInput.value != newPassInput.value)\">\n                <input type=\"password\" id=\"reNewPassword\" placeholder='{{\"PLACEHOLDER.CONFIRM_PWD\" | translate}}'\n                    required\n                    pattern=\"^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)[a-zA-Z\\d]{7,}$\"\n                    name=\"reNewPassword\"\n                    [(ngModel)]=\"reNewPwd\"\n                    #reNewPassInput=\"ngModel\" size=\"25\">\n                <span class=\"tooltip-content\">\n                    {{'TOOLTIP.CONFIRM_PWD' | translate}}\n                </span>\n            </label>\n                </div>\n            </section>\n            <inline-alert (confirmEvt)=\"confirmCancel($event)\"></inline-alert>\n        </form>\n    </div>\n    <div class=\"modal-footer\">\n        <span class=\"spinner spinner-inline\" style=\"top:8px;\" [hidden]=\"showProgress === false\"></span>\n        <button type=\"button\" class=\"btn btn-outline\" (click)=\"close()\">{{'BUTTON.CANCEL' | translate}}</button>\n        <button type=\"button\" class=\"btn btn-primary\" [disabled]=\"!isValid || showProgress\" (click)=\"doOk()\">{{'BUTTON.OK' | translate}}</button>\n    </div>\n</clr-modal>"

/***/ }),

/***/ 823:
/***/ (function(module, exports) {

module.exports = "<clr-modal [(clrModalOpen)]=\"opened\" [clrModalStaticBackdrop]=\"true\">\n    <h3 class=\"modal-title\">{{'RESET_PWD.TITLE' | translate}}</h3>\n    <label class=\"modal-title reset-modal-title-override\">{{'RESET_PWD.CAPTION2' | translate}}</label>\n    <div class=\"modal-body\" style=\"overflow-y: hidden;\">\n        <form #resetPwdForm=\"ngForm\" class=\"form\">\n            <section class=\"form-block\">\n                <div class=\"form-group\">\n                    <label for=\"newPassword\">{{'CHANGE_PWD.NEW_PWD' | translate}}</label>\n                    <label for=\"newPassword\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-bottom-left\" [class.invalid]='getValidationState(\"newPassword\") === false'>\n                <input type=\"password\" id=\"newPassword\" placeholder='{{\"PLACEHOLDER.NEW_PWD\" | translate}}'\n                    required\n                    pattern=\"^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)[a-zA-Z\\d]{7,}$\"\n                    name=\"newPassword\"\n                    [(ngModel)]=\"password\"\n                    #newPassInput=\"ngModel\" \n                    size=\"25\"\n                    (input)='handleValidation(\"newPassword\", true)'\n                    (focusout)='handleValidation(\"newPassword\", false)'>\n                <span class=\"tooltip-content\">\n                    {{'TOOLTIP.PASSWORD' | translate}}\n                </span>\n            </label>\n                </div>\n                <div class=\"form-group\">\n                    <label for=\"reNewPassword\">{{'CHANGE_PWD.CONFIRM_PWD' | translate}}</label>\n                    <label for=\"reNewPassword\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-bottom-left\" [class.invalid]='getValidationState(\"reNewPassword\") === false'>\n                <input type=\"password\" id=\"reNewPassword\" placeholder='{{\"PLACEHOLDER.CONFIRM_PWD\" | translate}}'\n                    required\n                    pattern=\"^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)[a-zA-Z\\d]{7,}$\"\n                    name=\"reNewPassword\"\n                    [(ngModel)]=\"ngModel\"\n                    #reNewPassInput \n                    size=\"25\"\n                    (input)='handleValidation(\"reNewPassword\", true)' \n                    (focusout)='handleValidation(\"reNewPassword\", false)'>\n                <span class=\"tooltip-content\">\n                    {{'TOOLTIP.CONFIRM_PWD' | translate}}\n                </span>\n            </label>\n                </div>\n            </section>\n            <inline-alert></inline-alert>\n            <div style=\"height: 30px;\"></div>\n        </form>\n    </div>\n    <div class=\"modal-footer\">\n        <span class=\"spinner spinner-inline\" style=\"top:8px;\" [hidden]=\"showProgress === false\"></span>\n        <button type=\"button\" class=\"btn btn-outline\" (click)=\"close()\">{{'BUTTON.CANCEL' | translate}}</button>\n        <button type=\"button\" class=\"btn btn-primary\" [disabled]=\"!isValid || showProgress\" (click)=\"send()\">{{'BUTTON.OK' | translate}}</button>\n    </div>\n</clr-modal>"

/***/ }),

/***/ 824:
/***/ (function(module, exports) {

module.exports = "<div class=\"login-wrapper\">\n    <form #signInForm=\"ngForm\" class=\"login\">\n        <label class=\"title\">\n        VMware Harbor<span class=\"trademark\">&#8482;</span>\n    </label>\n        <div class=\"login-group\">\n            <label for=\"username\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-left\" [class.invalid]=\"userNameInput.invalid && (userNameInput.dirty || userNameInput.touched)\">\n                <input class=\"username\" type=\"text\" required\n                [(ngModel)]=\"signInCredential.principal\" \n                name=\"login_username\" id=\"login_username\" placeholder='{{\"PLACEHOLDER.SIGN_IN_NAME\" | translate}}'\n                #userNameInput='ngModel'>\n                <span class=\"tooltip-content\">\n                    {{ 'TOOLTIP.SIGN_IN_USERNAME' | translate }}\n                </span>\n            </label>\n            <label for=\"username\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-left\" [class.invalid]=\"passwordInput.invalid && (passwordInput.dirty || passwordInput.touched)\">\n                <input class=\"password\" type=\"password\" required \n                [(ngModel)]=\"signInCredential.password\" \n                name=\"login_password\" id=\"login_password\" placeholder='{{\"PLACEHOLDER.SIGN_IN_PWD\" | translate}}'\n                #passwordInput=\"ngModel\">\n                <span class=\"tooltip-content\">\n                    {{ 'TOOLTIP.SIGN_IN_PWD' | translate }}\n                </span>\n            </label>\n            <div class=\"checkbox\">\n                <input type=\"checkbox\" id=\"rememberme\">\n                <label for=\"rememberme\">{{ 'SIGN_IN.REMEMBER' | translate }}</label>\n                <a href=\"javascript:void(0)\" class=\"forgot-password-link\" (click)=\"forgotPassword()\">{{'SIGN_IN.FORGOT_PWD' | translate}}</a>\n            </div>\n            <div [class.visibility-hidden]=\"!isError\" class=\"error active\">\n                {{ 'SIGN_IN.INVALID_MSG' | translate }}\n            </div>\n            <button [disabled]=\"isOnGoing || !isValid\" type=\"submit\" class=\"btn btn-primary\" (click)=\"signIn()\">{{ 'BUTTON.LOG_IN' | translate }}</button>\n            <a href=\"javascript:void(0)\" class=\"signup\" (click)=\"signUp()\" *ngIf=\"selfSignUp\">{{ 'BUTTON.SIGN_UP_LINK' | translate }}</a>\n        </div>\n    </form>\n</div>\n<sign-up #signupDialog></sign-up>\n<forgot-password #forgotPwdDialog></forgot-password>"

/***/ }),

/***/ 825:
/***/ (function(module, exports) {

module.exports = "<clr-modal [(clrModalOpen)]=\"opened\" [clrModalStaticBackdrop]=\"staticBackdrop\" [clrModalClosable]=\"true\">\n    <h3 class=\"modal-title\">{{'SIGN_UP.TITLE' | translate}}</h3>\n    <div class=\"modal-body\" style=\"overflow-y: hidden;\">\n        <new-user-form isSelfRegistration=\"true\" (valueChange)=\"formValueChange($event)\"></new-user-form>\n        <inline-alert (confirmEvt)=\"confirmCancel($event)\"></inline-alert>\n    </div>\n    <div class=\"modal-footer\">\n        <span class=\"spinner spinner-inline\" style=\"top:8px;\" [hidden]=\"inProgress === false\"> </span>\n        <button type=\"button\" class=\"btn btn-outline\" (click)=\"close()\">{{'BUTTON.CANCEL' | translate}}</button>\n        <button type=\"button\" class=\"btn btn-primary\" [disabled]=\"!isValid || inProgress\" (click)=\"create()\">{{ 'BUTTON.SIGN_UP' | translate }}</button>\n    </div>\n</clr-modal>"

/***/ }),

/***/ 826:
/***/ (function(module, exports) {

module.exports = "<router-outlet></router-outlet>"

/***/ }),

/***/ 827:
/***/ (function(module, exports) {

module.exports = ""

/***/ }),

/***/ 828:
/***/ (function(module, exports) {

module.exports = "<form class=\"search\">\n    <label for=\"search_input\">\n      <input #globalSearchBox id=\"search_input\" type=\"text\" (keyup)=\"search(globalSearchBox.value)\" placeholder='{{\"GLOBAL_SEARCH.PLACEHOLDER\" | translate}}'>\n    </label>\n</form>"

/***/ }),

/***/ 829:
/***/ (function(module, exports) {

module.exports = "<div class=\"search-overlay\" *ngIf=\"state\">\n    <div id=\"placeholder1\" style=\"height: 24px;\"></div>\n    <div class=\"search-header\">\n        <span class=\"search-title\">Search results for '{{currentTerm}}'</span>\n        <span class=\"search-close\" (mouseover)=\"mouseAction(true)\" (mouseout)=\"mouseAction(false)\">\n            <clr-icon shape=\"close\" [class.is-highlight]=\"hover\" size=\"36\" (click)=\"close()\"></clr-icon>\n        </span>\n    </div>\n    <!-- spinner -->\n    <div class=\"spinner spinner-lg search-spinner\" [hidden]=\"done\">Search...</div>\n    <div id=\"results\">\n        <h2>Projects</h2>\n        <div class=\"grid-header-wrapper\">\n            <grid-filter class=\"grid-filter\" filterPlaceholder='{{\"PROJECT.FILTER_PLACEHOLDER\" | translate}}' (filter)=\"doFilterProjects($event)\"></grid-filter>\n        </div>\n        <list-project [projects]=\"searchResults.project\" [mode]=\"listMode\"></list-project>\n        <h2>Repositories</h2>\n        <list-repository [repositories]=\"searchResults.repository\" [mode]=\"listMode\"></list-repository>\n    </div>\n</div>"

/***/ }),

/***/ 830:
/***/ (function(module, exports) {

module.exports = "<clr-main-container>\n    <global-message [isAppLevel]=\"true\"></global-message>\n    <navigator (showAccountSettingsModal)=\"openModal($event)\" (showPwdChangeModal)=\"openModal($event)\"></navigator>\n    <div class=\"content-container\">\n        <div class=\"content-area\" [class.container-override]=\"showSearch\" [class.start-content-padding]=\"isStartPage\">\n            <global-message [isAppLevel]=\"false\"></global-message>\n            <!-- Only appear when searching -->\n            <search-result></search-result>\n            <router-outlet></router-outlet>\n        </div>\n        <nav class=\"sidenav\" *ngIf=\"isUserExisting\" [class.side-nav-override]=\"showSearch\" (click)='watchClickEvt()'>\n            <section class=\"sidenav-content\">\n                <a routerLink=\"/harbor/dashboard\" routerLinkActive=\"active\" class=\"nav-link\">{{'SIDE_NAV.DASHBOARD' | translate}}</a>\n                <a routerLink=\"/harbor/projects\" routerLinkActive=\"active\" class=\"nav-link\">{{'SIDE_NAV.PROJECTS' | translate}}</a>\n                <a routerLink=\"/harbor/logs\" routerLinkActive=\"active\" class=\"nav-link\">{{'SIDE_NAV.LOGS' | translate}}</a>\n                <section class=\"nav-group collapsible\" *ngIf=\"isSystemAdmin\">\n                    <input id=\"tabsystem\" type=\"checkbox\">\n                    <label for=\"tabsystem\">{{'SIDE_NAV.SYSTEM_MGMT.NAME' | translate}}</label>\n                    <ul class=\"nav-list\">\n                        <li><a class=\"nav-link\" routerLink=\"/harbor/users\" routerLinkActive=\"active\">{{'SIDE_NAV.SYSTEM_MGMT.USER' | translate}}</a></li>\n                        <li><a class=\"nav-link\" routerLink=\"/harbor/replications/endpoints\" routerLinkActive=\"active\">{{'SIDE_NAV.SYSTEM_MGMT.REPLICATION' | translate}}</a></li>\n                        <li><a class=\"nav-link\" routerLink=\"/harbor/configs\" routerLinkActive=\"active\">{{'SIDE_NAV.SYSTEM_MGMT.CONFIG' | translate}}</a></li>\n                    </ul>\n                </section>\n            </section>\n        </nav>\n    </div>\n</clr-main-container>\n<account-settings-modal></account-settings-modal>\n<password-setting></password-setting>\n<deletion-dialog></deletion-dialog>\n<about-dialog></about-dialog>"

/***/ }),

/***/ 831:
/***/ (function(module, exports) {

module.exports = "<clr-header class=\"header-5 header\">\n    <div class=\"branding\">\n        <a href=\"javascript:void(0)\" class=\"nav-link\" (click)=\"homeAction()\">\n            <clr-icon shape=\"vm-bug\"></clr-icon>\n            <span class=\"title\">Harbor</span>\n        </a>\n    </div>\n    <div class=\"header-nav\">\n        <a href=\"{{admiralLink}}\" class=\"nav-link\" *ngIf=\"isIntegrationMode\"><span class=\"nav-text\">Management</span></a>\n        <a href=\"javascript:void(0)\" routerLink=\"/harbor/dashboard\" class=\"active nav-link\" *ngIf=\"isIntegrationMode\"><span class=\"nav-text\">Registry</span></a>\n    </div>\n    <global-search></global-search>\n    <div class=\"header-actions\">\n        <a href=\"javascript:void(0)\" class=\"nav-link nav-text\" routerLink=\"/sign-in\" routerLinkActive=\"active\" *ngIf=\"isSessionValid === false\">{{'SIGN_IN.HEADER_LINK' | translate}}</a>\n        <div class=\"nav-divider\" *ngIf=\"!isSessionValid\"></div>\n        <a href=\"javascript:void(0)\" class=\"nav-link nav-text\" (click)=\"openSignUp()\" *ngIf=\"isSessionValid === false\">{{'SIGN_UP.TITLE' | translate}}</a>\n        <clr-dropdown class=\"dropdown bottom-left\">\n            <button class=\"nav-icon\" clrDropdownToggle style=\"width: 98px;\">\n                <clr-icon shape=\"world\" style=\"left:-8px;\"></clr-icon>\n                <span style=\"padding-right: 8px;\">{{currentLang}}</span>\n                <clr-icon shape=\"caret down\"></clr-icon>\n            </button>\n            <div class=\"dropdown-menu\">\n                <a href=\"javascript:void(0)\" clrDropdownItem (click)='switchLanguage(\"en\")' [class.lang-selected]='matchLang(\"en\")'>English</a>\n                <a href=\"javascript:void(0)\" clrDropdownItem (click)='switchLanguage(\"zh\")' [class.lang-selected]='matchLang(\"zh\")'>中文简体</a>\n            </div>\n        </clr-dropdown>\n        <clr-dropdown [clrMenuPosition]=\"'bottom-right'\" class=\"dropdown\" *ngIf=\"isSessionValid\">\n            <button class=\"nav-text\" clrDropdownToggle>\n              <clr-icon shape=\"user\" class=\"is-inverse\" size=\"24\" style=\"left: -2px;\"></clr-icon>\n              <span>{{accountName}}</span>\n              <clr-icon shape=\"caret down\"></clr-icon>\n            </button>\n            <div class=\"dropdown-menu\">\n                <a href=\"javascript:void(0)\" clrDropdownItem (click)=\"openAccountSettingsModal()\">{{'ACCOUNT_SETTINGS.PROFILE' | translate}}</a>\n                <a href=\"javascript:void(0)\" clrDropdownItem (click)=\"openChangePwdModal()\">{{'ACCOUNT_SETTINGS.CHANGE_PWD' | translate}}</a>\n                <a href=\"javascript:void(0)\" clrDropdownItem (click)=\"openAboutDialog()\">{{'ACCOUNT_SETTINGS.ABOUT' | translate}}</a>\n                <div class=\"dropdown-divider\"></div>\n                <a href=\"javascript:void(0)\" clrDropdownItem (click)=\"logOut()\">{{'ACCOUNT_SETTINGS.LOGOUT' | translate}}</a>\n            </div>\n        </clr-dropdown>\n        <a href=\"javascript:void(0)\" class=\"nav-link nav-text\" (click)=\"openAboutDialog()\" *ngIf=\"isSessionValid === false\">{{'ACCOUNT_SETTINGS.ABOUT' | translate}}</a>\n    </div>\n</clr-header>"

/***/ }),

/***/ 832:
/***/ (function(module, exports) {

module.exports = "<!-- Authenticated-->\n<div class=\"row row-fill-height row-margin\" *ngIf=\"isSessionValid\">\n    <div class=\"col-xs-12 col-sm-12 col-md-12 col-lg-12 col-xl-12\">\n        <statistics-panel></statistics-panel>\n        <top-repo></top-repo>\n    </div>\n</div>\n\n<!-- Guest -->\n<div class=\"row row-fill-height\" *ngIf=\"!isSessionValid\">\n    <div class=\"col-xs-12 col-sm-12 col-md-5 col-lg-5 col-xl-5 column-fill-height\">\n        <div class=\"start-card\">\n            <div class=\"card-img my-card-img\">\n            </div>\n            <div class=\"card-block\">\n                <h3 class=\"card-title\">Getting Start</h3>\n                <p class=\"card-text\">\n                    {{'START_PAGE.GETTING_START' | translate}}\n                </p>\n            </div>\n            <div class=\"card-footer my-card-footer\">\n                <a href=\"http://vmware.github.io/harbor/\" target=\"_blank\" class=\"btn btn-sm btn-link\">Learn More</a>\n            </div>\n        </div>\n    </div>\n    <div class=\"col-xs-12 col-sm-12 col-md-7 col-lg-7 col-xl-7\">\n        <top-repo></top-repo>\n    </div>\n</div>"

/***/ }),

/***/ 833:
/***/ (function(module, exports) {

module.exports = "<form #authConfigFrom=\"ngForm\" class=\"form\">\n    <section class=\"form-block\">\n        <div class=\"form-group\">\n            <label for=\"authMode\">{{'CONFIG.AUTH_MODE' | translate }}</label>\n            <div class=\"select\">\n                <select id=\"authMode\" name=\"authMode\" [disabled]=\"disabled(currentConfig.auth_mode)\" [(ngModel)]=\"currentConfig.auth_mode.value\">\n                    <option value=\"db_auth\">{{'CONFIG.AUTH_MODE_DB' | translate }}</option>\n                    <option value=\"ldap_auth\">{{'CONFIG.AUTH_MODE_LDAP' | translate }}</option>\n                </select>\n            </div>\n            <a href=\"javascript:void(0)\" role=\"tooltip\" aria-haspopup=\"true\" class=\"tooltip tooltip-top-right\">\n                <clr-icon shape=\"info-circle\" class=\"is-info\" size=\"24\"></clr-icon>\n                <span class=\"tooltip-content\">{{'CONFIG.TOOLTIP.AUTH_MODE' | translate}}</span>\n            </a>\n        </div>\n    </section>\n    <section class=\"form-block\" *ngIf=\"showLdap\">\n        <div class=\"form-group\">\n            <label for=\"ldapUrl\" class=\"required\">{{'CONFIG.LDAP.URL' | translate}}</label>\n            <label for=\"ldapUrl\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-right\" [class.invalid]=\"ldapUrlInput.invalid && (ldapUrlInput.dirty || ldapUrlInput.touched)\">\n                      <input name=\"ldapUrl\" type=\"text\" #ldapUrlInput=\"ngModel\" [(ngModel)]=\"currentConfig.ldap_url.value\" \n                      required \n                      id=\"ldapUrl\" \n                      size=\"40\" \n                      [disabled]=\"disabled(currentConfig.ldap_url)\">\n                      <span class=\"tooltip-content\">\n                          {{'TOOLTIP.ITEM_REQUIRED' | translate}}\n                      </span>\n                    </label>\n        </div>\n        <div class=\"form-group\">\n            <label for=\"ldapSearchDN\" class=\"required\">{{'CONFIG.LDAP.SEARCH_DN' | translate}}</label>\n            <label for=\"ldapSearchDN\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-right\" [class.invalid]=\"ldapSearchDNInput.invalid && (ldapSearchDNInput.dirty || ldapSearchDNInput.touched)\">\n                      <input name=\"ldapSearchDN\" type=\"text\" #ldapSearchDNInput=\"ngModel\" [(ngModel)]=\"currentConfig.ldap_search_dn.value\" \n                      required \n                      id=\"ldapSearchDN\" \n                      size=\"40\" [disabled]=\"disabled(currentConfig.ldap_search_dn)\">\n                      <span class=\"tooltip-content\">\n                          {{'TOOLTIP.ITEM_REQUIRED' | translate}}\n                      </span>\n                    </label>\n            <a href=\"javascript:void(0)\" role=\"tooltip\" aria-haspopup=\"true\" class=\"tooltip tooltip-top-right\">\n                <clr-icon shape=\"info-circle\" class=\"is-info\" size=\"24\"></clr-icon>\n                <span class=\"tooltip-content\">{{'CONFIG.TOOLTIP.LDAP_SEARCH_DN' | translate}}</span>\n            </a>\n        </div>\n        <div class=\"form-group\">\n            <label for=\"ldapSearchPwd\" class=\"required\">{{'CONFIG.LDAP.SEARCH_PWD' | translate}}</label>\n            <label for=\"ldapSearchPwd\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-right\" [class.invalid]=\"ldapSearchPwdInput.invalid && (ldapSearchPwdInput.dirty || ldapSearchPwdInput.touched)\">\n                      <input name=\"ldapSearchPwd\" type=\"password\" #ldapSearchPwdInput=\"ngModel\" [(ngModel)]=\"currentConfig.ldap_search_password.value\" \n                      required \n                      id=\"ldapSearchPwd\" \n                      size=\"40\" [disabled]=\"disabled(currentConfig.ldap_search_password)\">\n                      <span class=\"tooltip-content\">\n                          {{'TOOLTIP.ITEM_REQUIRED' | translate}}\n                      </span>\n                    </label>\n        </div>\n        <div class=\"form-group\">\n            <label for=\"ldapBaseDN\" class=\"required\">{{'CONFIG.LDAP.BASE_DN' | translate}}</label>\n            <label for=\"ldapBaseDN\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-right\" [class.invalid]=\"ldapBaseDNInput.invalid && (ldapBaseDNInput.dirty || ldapBaseDNInput.touched)\">\n                      <input name=\"ldapBaseDN\" type=\"text\" #ldapBaseDNInput=\"ngModel\" [(ngModel)]=\"currentConfig.ldap_base_dn.value\" \n                      required \n                      id=\"ldapBaseDN\" \n                      size=\"40\" [disabled]=\"disabled(currentConfig.ldap_base_dn)\">\n                      <span class=\"tooltip-content\">\n                          {{'TOOLTIP.ITEM_REQUIRED' | translate}}\n                      </span>\n                    </label>\n            <a href=\"javascript:void(0)\" role=\"tooltip\" aria-haspopup=\"true\" class=\"tooltip tooltip-top-right\">\n                <clr-icon shape=\"info-circle\" class=\"is-info\" size=\"24\"></clr-icon>\n                <span class=\"tooltip-content\">{{'CONFIG.TOOLTIP.LDAP_BASE_DN' | translate}}</span>\n            </a>\n        </div>\n        <div class=\"form-group\">\n            <label for=\"ldapFilter\">{{'CONFIG.LDAP.FILTER' | translate}}</label>\n            <label for=\"ldapFilter\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-right\">\n                      <input name=\"ldapFilter\" type=\"text\" #ldapFilterInput=\"ngModel\" [(ngModel)]=\"currentConfig.ldap_filter.value\" \n                      id=\"ldapFilter\" \n                      size=\"40\" [disabled]=\"disabled(currentConfig.ldap_filter)\">\n                      <span class=\"tooltip-content\">\n                          {{'TOOLTIP.ITEM_REQUIRED' | translate}}\n                      </span>\n                    </label>\n        </div>\n        <div class=\"form-group\">\n            <label for=\"ldapUid\" class=\"required\">{{'CONFIG.LDAP.UID' | translate}}</label>\n            <label for=\"ldapUid\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-right\" [class.invalid]=\"ldapUidInput.invalid && (ldapUidInput.dirty || ldapUidInput.touched)\">\n                      <input name=\"ldapUid\" type=\"text\" #ldapUidInput=\"ngModel\" [(ngModel)]=\"currentConfig.ldap_uid.value\" \n                      required \n                      id=\"ldapUid\" \n                      size=\"40\" [disabled]=\"disabled(currentConfig.ldap_uid)\">\n                      <span class=\"tooltip-content\">\n                          {{'TOOLTIP.ITEM_REQUIRED' | translate}}\n                      </span>\n                    </label>\n            <a href=\"javascript:void(0)\" role=\"tooltip\" aria-haspopup=\"true\" class=\"tooltip tooltip-top-right\">\n                <clr-icon shape=\"info-circle\" class=\"is-info\" size=\"24\"></clr-icon>\n                <span class=\"tooltip-content\">{{'CONFIG.TOOLTIP.LDAP_UID' | translate}}</span>\n            </a>\n        </div>\n        <div class=\"form-group\">\n            <label for=\"ldapScope\">{{'CONFIG.LDAP.SCOPE' | translate}}</label>\n            <div class=\"select\">\n                <select id=\"ldapScope\" name=\"ldapScope\" [(ngModel)]=\"currentConfig.ldap_scope.value\" [disabled]=\"disabled(currentConfig.ldap_scope)\">\n                    <option value=\"1\">{{'CONFIG.SCOPE_BASE' | translate }}</option>\n                    <option value=\"2\">{{'CONFIG.SCOPE_ONE_LEVEL' | translate }}</option>\n                    <option value=\"3\">{{'CONFIG.SCOPE_SUBTREE' | translate }}</option>\n                </select>\n            </div>\n            <a href=\"javascript:void(0)\" role=\"tooltip\" aria-haspopup=\"true\" class=\"tooltip tooltip-top-right\">\n                <clr-icon shape=\"info-circle\" class=\"is-info\" size=\"24\"></clr-icon>\n                <span class=\"tooltip-content\">{{'CONFIG.TOOLTIP.LDAP_SCOPE' | translate}}</span>\n            </a>\n        </div>\n    </section>\n    <section class=\"form-block\">\n        <div class=\"form-group\">\n            <label for=\"proCreation\">{{'CONFIG.PRO_CREATION_RESTRICTION' | translate}}</label>\n            <div class=\"select\">\n                <select id=\"proCreation\" name=\"proCreation\" [(ngModel)]=\"currentConfig.project_creation_restriction.value\" [disabled]=\"disabled(currentConfig.project_creation_restriction)\">\n                    <option value=\"everyone\">{{'CONFIG.PRO_CREATION_EVERYONE' | translate }}</option>\n                    <option value=\"adminonly\">{{'CONFIG.PRO_CREATION_ADMIN' | translate }}</option>\n                </select>\n            </div>\n            <a href=\"javascript:void(0)\" role=\"tooltip\" aria-haspopup=\"true\" class=\"tooltip tooltip-top-right\">\n                <clr-icon shape=\"info-circle\" class=\"is-info\" size=\"24\"></clr-icon>\n                <span class=\"tooltip-content\">{{'CONFIG.TOOLTIP.AUTH_MODE' | translate}}</span>\n            </a>\n        </div>\n        <div class=\"form-group\">\n            <label for=\"selfReg\">{{'CONFIG.SELF_REGISTRATION' | translate}}</label>\n            <clr-checkbox name=\"selfReg\" id=\"selfReg\" [(ngModel)]=\"currentConfig.self_registration.value\" [disabled]=\"disabled(currentConfig.self_registration)\">\n                <a href=\"javascript:void(0)\" role=\"tooltip\" aria-haspopup=\"true\" class=\"tooltip tooltip-top-right\" style=\"top:-8px;\">\n                    <clr-icon shape=\"info-circle\" class=\"is-info\" size=\"24\"></clr-icon>\n                    <span class=\"tooltip-content\">{{'CONFIG.TOOLTIP.SELF_REGISTRATION' | translate}}</span>\n                </a>\n            </clr-checkbox>\n        </div>\n    </section>\n</form>"

/***/ }),

/***/ 834:
/***/ (function(module, exports) {

module.exports = "<h1 style=\"display: inline-block;\">{{'CONFIG.TITLE' | translate }}</h1>\n<span class=\"spinner spinner-inline\" [hidden]=\"inProgress === false\"></span>\n<clr-tabs (clrTabsCurrentTabLinkChanged)=\"tabLinkChanged($event)\">\n    <clr-tab-link [clrTabLinkId]=\"'config-auth'\" [clrTabLinkActive]=\"true\">{{'CONFIG.AUTH' | translate }}</clr-tab-link>\n    <clr-tab-link [clrTabLinkId]=\"'config-replication'\">{{'CONFIG.REPLICATION' | translate }}</clr-tab-link>\n    <clr-tab-link [clrTabLinkId]=\"'config-email'\">{{'CONFIG.EMAIL' | translate }}</clr-tab-link>\n    <clr-tab-link [clrTabLinkId]=\"'config-system'\">{{'CONFIG.SYSTEM' | translate }}</clr-tab-link>\n\n    <clr-tab-content [clrTabContentId]=\"'authentication'\" [clrTabContentActive]=\"true\">\n        <config-auth [ldapConfig]=\"allConfig\"></config-auth>\n    </clr-tab-content>\n    <clr-tab-content [clrTabContentId]=\"'replication'\">\n        <form #repoConfigFrom=\"ngForm\" class=\"form\">\n            <section class=\"form-block\">\n                <div class=\"form-group\">\n                    <label for=\"verifyRemoteCert\">{{'CONFIG.VERIFY_REMOTE_CERT' | translate }}</label>\n                    <clr-checkbox name=\"verifyRemoteCert\" id=\"verifyRemoteCert\" [(ngModel)]=\"allConfig.verify_remote_cert.value\" [disabled]=\"disabled(allConfig.verify_remote_cert)\">\n                        <a href=\"javascript:void(0)\" role=\"tooltip\" aria-haspopup=\"true\" class=\"tooltip tooltip-lg tooltip-top-right\" style=\"top:-8px;\">\n                            <clr-icon shape=\"info-circle\" class=\"is-info\" size=\"24\"></clr-icon>\n                            <span class=\"tooltip-content\">{{'CONFIG.TOOLTIP.VERIFY_REMOTE_CERT' | translate }}</span>\n                        </a>\n                    </clr-checkbox>\n                </div>\n            </section>\n        </form>\n    </clr-tab-content>\n    <clr-tab-content [clrTabContentId]=\"'email'\">\n        <config-email [mailConfig]=\"allConfig\"></config-email>\n    </clr-tab-content>\n    <clr-tab-content [clrTabContentId]=\"'system_settings'\">\n        <form #systemConfigFrom=\"ngForm\" class=\"form\">\n            <section class=\"form-block\">\n                <div class=\"form-group\">\n                    <label for=\"tokenExpiration\" class=\"required\">{{'CONFIG.TOKEN_EXPIRATION' | translate}}</label>\n                    <label for=\"tokenExpiration\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-right\" [class.invalid]=\"tokenExpirationInput.invalid && (tokenExpirationInput.dirty || tokenExpirationInput.touched)\">\n                      <input name=\"tokenExpiration\" type=\"text\" #tokenExpirationInput=\"ngModel\" [(ngModel)]=\"allConfig.token_expiration.value\" \n                      required \n                      pattern=\"^[1-9]{1}[\\d]*$\"\n                      id=\"tokenExpiration\" \n                      size=\"40\" [disabled]=\"disabled(allConfig.token_expiration)\">\n                      <span class=\"tooltip-content\">\n                          {{'TOOLTIP.NUMBER_REQUIRED' | translate}}\n                      </span>\n                    </label>\n                    <a href=\"javascript:void(0)\" role=\"tooltip\" aria-haspopup=\"true\" class=\"tooltip tooltip-top-right\">\n                        <clr-icon shape=\"info-circle\" class=\"is-info\" size=\"24\"></clr-icon>\n                        <span class=\"tooltip-content\">{{'CONFIG.TOOLTIP.TOKEN_EXPIRATION' | translate}}</span>\n                    </a>\n                </div>\n            </section>\n        </form>\n    </clr-tab-content>\n</clr-tabs>\n<div>\n    <button type=\"button\" class=\"btn btn-primary\" (click)=\"save()\" [disabled]=\"!isValid() || !hasChanges()\">{{'BUTTON.SAVE' | translate}}</button>\n    <button type=\"button\" class=\"btn btn-outline\" (click)=\"cancel()\" [disabled]=\"!isValid() || !hasChanges()\">{{'BUTTON.CANCEL' | translate}}</button>\n    <button type=\"button\" class=\"btn btn-outline\" (click)=\"testMailServer()\" *ngIf=\"showTestServerBtn\" [disabled]=\"!isMailConfigValid()\">{{'BUTTON.TEST_MAIL' | translate}}</button>\n    <button type=\"button\" class=\"btn btn-outline\" (click)=\"testLDAPServer()\" *ngIf=\"showLdapServerBtn\" [disabled]=\"!isLDAPConfigValid()\">{{'BUTTON.TEST_LDAP' | translate}}</button>\n    <span class=\"spinner spinner-inline\" [hidden]=\"!testingInProgress\"></span>\n</div>"

/***/ }),

/***/ 835:
/***/ (function(module, exports) {

module.exports = "<form #mailConfigFrom=\"ngForm\" class=\"form\">\n    <section class=\"form-block\">\n        <div class=\"form-group\">\n            <label for=\"mailServer\" class=\"required\">{{'CONFIG.MAIL_SERVER' | translate}}</label>\n            <label for=\"mailServer\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-right\" [class.invalid]=\"mailServerInput.invalid && (mailServerInput.dirty || mailServerInput.touched)\">\n                      <input name=\"mailServer\" type=\"text\" #mailServerInput=\"ngModel\" [(ngModel)]=\"currentConfig.email_host.value\" \n                      required \n                      id=\"mailServer\" \n                      size=\"40\" [disabled]=\"disabled(currentConfig.email_host)\">\n                      <span class=\"tooltip-content\">\n                          {{'TOOLTIP.ITEM_REQUIRED' | translate}}\n                      </span>\n                    </label>\n        </div>\n        <div class=\"form-group\">\n            <label for=\"emailPort\" class=\"required\">{{'CONFIG.MAIL_SERVER_PORT' | translate}}</label>\n            <label for=\"emailPort\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-right\" [class.invalid]=\"emailPortInput.invalid && (emailPortInput.dirty || emailPortInput.touched)\">\n                      <input name=\"emailPort\" type=\"text\" #emailPortInput=\"ngModel\" [(ngModel)]=\"currentConfig.email_port.value\" \n                      required \n                      port\n                      id=\"emailPort\" \n                      size=\"40\" [disabled]=\"disabled(currentConfig.email_port)\">\n                      <span class=\"tooltip-content\">\n                          {{'TOOLTIP.PORT_REQUIRED' | translate}}\n                      </span>\n                    </label>\n        </div>\n        <div class=\"form-group\">\n            <label for=\"emailUsername\">{{'CONFIG.MAIL_USERNAME' | translate}}</label>\n            <label for=\"emailUsername\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-right\" [class.invalid]=\"false\">\n                      <input name=\"emailUsername\" type=\"text\" #emailUsernameInput=\"ngModel\" [(ngModel)]=\"currentConfig.email_username.value\" \n                      id=\"emailUsername\" \n                      size=\"40\" [disabled]=\"disabled(currentConfig.email_username)\">\n                      <span class=\"tooltip-content\">\n                          {{'TOOLTIP.ITEM_REQUIRED' | translate}}\n                      </span>\n                    </label>\n        </div>\n        <div class=\"form-group\">\n            <label for=\"emailPassword\">{{'CONFIG.MAIL_PASSWORD' | translate}}</label>\n            <label for=\"emailPassword\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-right\" [class.invalid]=\"false\">\n                      <input name=\"emailPassword\" type=\"password\" #emailPasswordInput=\"ngModel\" [(ngModel)]=\"currentConfig.email_password.value\" \n                      id=\"emailPassword\" \n                      size=\"40\" [disabled]=\"disabled(currentConfig.email_password)\">\n                      <span class=\"tooltip-content\">\n                          {{'TOOLTIP.ITEM_REQUIRED' | translate}}\n                      </span>\n                    </label>\n        </div>\n        <div class=\"form-group\">\n            <label for=\"emailFrom\" class=\"required\">{{'CONFIG.MAIL_FROM' | translate}}</label>\n            <label for=\"emailFrom\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-right\" [class.invalid]=\"emailFromInput.invalid && (emailFromInput.dirty || emailFromInput.touched)\">\n                      <input name=\"emailFrom\" type=\"text\" #emailFromInput=\"ngModel\" [(ngModel)]=\"currentConfig.email_from.value\" \n                      required \n                      id=\"emailFrom\" \n                      size=\"40\" [disabled]=\"disabled(currentConfig.email_from)\">\n                      <span class=\"tooltip-content\">\n                          {{'TOOLTIP.ITEM_REQUIRED' | translate}}\n                      </span>\n                    </label>\n        </div>\n        <div class=\"form-group\">\n            <label for=\"selfReg\">{{'CONFIG.MAIL_SSL' | translate}}</label>\n            <clr-checkbox name=\"emailSSL\" id=\"emailSSL\" [(ngModel)]=\"currentConfig.email_ssl.value\" [disabled]=\"disabled(currentConfig.email_ssl)\">\n                <a href=\"javascript:void(0)\" role=\"tooltip\" aria-haspopup=\"true\" class=\"tooltip tooltip-top-right\" style=\"top:-8px;\">\n                    <clr-icon shape=\"info-circle\" class=\"is-info\" size=\"24\"></clr-icon>\n                    <span class=\"tooltip-content\">{{'CONFIG.SSL_TOOLTIP' | translate}}</span>\n                </a>\n            </clr-checkbox>\n        </div>\n    </section>\n</form>"

/***/ }),

/***/ 836:
/***/ (function(module, exports) {

module.exports = "<clr-alert [clrAlertType]=\"globalMessage.type\" [clrAlertAppLevel]=\"isAppLevel\" [(clrAlertClosed)]=\"!globalMessageOpened\" (clrAlertClosedChange)=\"onClose()\">\n    <div class=\"alert-item\">\n        <span class=\"alert-text\">\n          {{message}}\n        </span>\n        <div class=\"alert-actions\" *ngIf=\"needAuth\">\n            <button class=\"btn alert-action\" (click)=\"signIn()\">{{ 'BUTTON.LOG_IN' | translate }}</button>\n        </div>\n    </div>\n</clr-alert>"

/***/ }),

/***/ 837:
/***/ (function(module, exports) {

module.exports = "<div class=\"row\">\n  <div class=\"col-lg-12 col-md-12 col-sm-12 col-xs-12\">  \n    <div class=\"row flex-items-xs-right\">\n      <div class=\"flex-xs-middle\">\n        <button class=\"btn btn-link\" (click)=\"toggleOptionalName(currentOption)\">{{toggleName[currentOption] | translate}}</button>\n      </div>\n      <div class=\"flex-xs-middle\">\n        <grid-filter filterPlaceholder='{{\"AUDIT_LOG.FILTER_PLACEHOLDER\" | translate}}' (filter)=\"doSearchAuditLogs($event)\"></grid-filter>\n        <a href=\"javascript:void(0)\" (click)=\"refresh()\"><clr-icon shape=\"refresh\"></clr-icon></a>\n      </div>\n    </div>\n    <div class=\"row flex-items-xs-right\" [hidden]=\"currentOption === 0\">\n      <clr-dropdown [clrMenuPosition]=\"'bottom-left'\" >\n        <button class=\"btn btn-link\" clrDropdownToggle>\n          {{'AUDIT_LOG.ALL_OPERATIONS' | translate}}\n          <clr-icon shape=\"caret down\"></clr-icon>\n        </button>\n        <div class=\"dropdown-menu\">\n          <a href=\"javascript:void(0)\" clrDropdownItem *ngFor=\"let f of filterOptions\" (click)=\"toggleFilterOption(f.key)\"><clr-icon shape=\"check\" [hidden]=\"!f.checked\"></clr-icon> {{f.description | translate}}</a>\n        </div>\n      </clr-dropdown>\n      <div class=\"flex-xs-middle\">\n      <clr-icon shape=\"date\"></clr-icon><input type=\"date\" #fromTime  (change)=\"doSearchByTimeRange(fromTime.value, 'begin')\">\n      <clr-icon shape=\"date\"></clr-icon><input type=\"date\" #toTime  (change)=\"doSearchByTimeRange(toTime.value, 'end')\">\n      </div>\n    </div>\n    <clr-datagrid (clrDgRefresh)=\"retrieve($event)\">\n      <clr-dg-column>{{'AUDIT_LOG.USERNAME' | translate}}</clr-dg-column>\n      <clr-dg-column>{{'AUDIT_LOG.REPOSITORY_NAME' | translate}}</clr-dg-column>\n      <clr-dg-column>{{'AUDIT_LOG.TAGS' | translate}}</clr-dg-column>\n      <clr-dg-column>{{'AUDIT_LOG.OPERATION' | translate}}</clr-dg-column>\n      <clr-dg-column>{{'AUDIT_LOG.TIMESTAMP' | translate}}</clr-dg-column>\n      <clr-dg-row *ngFor=\"let l of auditLogs\">\n        <clr-dg-cell>{{l.username}}</clr-dg-cell>\n        <clr-dg-cell>{{l.repo_name}}</clr-dg-cell>\n        <clr-dg-cell>{{l.repo_tag}}</clr-dg-cell>\n        <clr-dg-cell>{{l.operation}}</clr-dg-cell>\n        <clr-dg-cell>{{l.op_time}}</clr-dg-cell>\n      </clr-dg-row>\n      <clr-dg-footer>\n        {{totalRecordCount}} {{'AUDIT_LOG.ITEMS' | translate}}\n        <clr-dg-pagination [clrDgPageSize]=\"pageOffset\" [clrDgTotalItems]=\"totalPage\"></clr-dg-pagination>\n      </clr-dg-footer>\n    </clr-datagrid>\n  </div>\n</div>"

/***/ }),

/***/ 838:
/***/ (function(module, exports) {

module.exports = "<div>\n    <h2 class=\"h2-log-override\">{{'SIDE_NAV.LOGS' | translate}}</h2>\n    <div class=\"action-head-pos\">\n        <span>\n            <label>{{'RECENT_LOG.SUB_TITLE' | translate}} </label>\n            <button type=\"submit\" class=\"btn btn-link custom-lines-button\" [class.lines-button-toggole]=\"lines === 10\" (click)=\"setLines(10)\">10</button>\n            <label> | </label>\n            <button type=\"submit\" class=\"btn btn-link custom-lines-button\" [class.lines-button-toggole]=\"lines === 25\" (click)=\"setLines(25)\">25</button>\n            <label> | </label>\n            <button type=\"submit\" class=\"btn btn-link custom-lines-button\" [class.lines-button-toggole]=\"lines === 50\" (click)=\"setLines(50)\">50</button>\n            <label>{{'RECENT_LOG.SUB_TITLE_SUFIX' | translate}}</label>\n        </span>\n        <grid-filter class=\"filter-log\" filterPlaceholder='{{\"AUDIT_LOG.FILTER_PLACEHOLDER\" | translate}}' (filter)=\"doFilter($event)\"></grid-filter>\n        <span class=\"refresh-btn\" (click)=\"refresh()\">\n            <clr-icon shape=\"refresh\" [hidden]=\"inProgress\" ng-disabled=\"inProgress\"></clr-icon>\n            <span class=\"spinner spinner-inline\" [hidden]=\"inProgress === false\"></span>\n        </span>\n    </div>\n    <div>\n        <clr-datagrid>\n            <clr-dg-column>{{'AUDIT_LOG.USERNAME' | translate}}</clr-dg-column>\n            <clr-dg-column>{{'AUDIT_LOG.REPOSITORY_NAME' | translate}}</clr-dg-column>\n            <clr-dg-column>{{'AUDIT_LOG.TAGS' | translate}}</clr-dg-column>\n            <clr-dg-column>{{'AUDIT_LOG.OPERATION' | translate}}</clr-dg-column>\n            <clr-dg-column>{{'AUDIT_LOG.TIMESTAMP' | translate}}</clr-dg-column>\n            <clr-dg-row *ngFor=\"let l of recentLogs\">\n                <clr-dg-cell>{{l.username}}</clr-dg-cell>\n                <clr-dg-cell>{{l.repo_name}}</clr-dg-cell>\n                <clr-dg-cell>{{l.repo_tag}}</clr-dg-cell>\n                <clr-dg-cell>{{l.operation}}</clr-dg-cell>\n                <clr-dg-cell>{{formatDateTime(l.op_time)}}</clr-dg-cell>\n            </clr-dg-row>\n            <clr-dg-footer>{{ (recentLogs ? recentLogs.length : 0) }} {{'AUDIT_LOG.ITEMS' | translate}}</clr-dg-footer>\n        </clr-datagrid>\n    </div>\n</div>"

/***/ }),

/***/ 839:
/***/ (function(module, exports) {

module.exports = "<clr-modal [(clrModalOpen)]=\"createProjectOpened\">\n  <h3 class=\"modal-title\">{{'PROJECT.NEW_PROJECT' | translate}}</h3>\n  <div class=\"modal-body\">\n    <form #projectForm=\"ngForm\">\n      <section class=\"form-block\">\n        <clr-alert [clrAlertType]=\"'alert-danger'\" [(clrAlertClosed)]=\"!errorMessageOpened\" (clrAlertClosedChange)=\"onErrorMessageClose()\">\n          <div class=\"alert-item\">\n            <span class=\"alert-text\">\n              {{errorMessage}}\n            </span>\n          </div>\n        </clr-alert>\n        <div class=\"form-group\">\n          <label for=\"create_project_name\" class=\"col-md-4\">{{'PROJECT.NAME' | translate}}</label>\n          <label for=\"create_project_name\" aria-haspopup=\"true\" role=\"tooltip\" [class.invalid]=\"projectName.invalid && (projectName.dirty || projectName.touched)\" [class.valid]=\"projectName.valid\" class=\"tooltip tooltip-validation tooltip-sm tooltip-bottom-right\">\n            <input type=\"text\" id=\"create_project_name\"  [(ngModel)]=\"project.name\" name=\"name\" size=\"20\" required minlength=\"2\" #projectName=\"ngModel\">\n            <span class=\"tooltip-content\" *ngIf=\"projectName.errors && projectName.errors.required && (projectName.dirty || projectName.touched)\">\n              {{'PROJECT.NAME_IS_REQUIRED' | translate}}\n            </span>\n            <span class=\"tooltip-content\" *ngIf=\"projectName.errors && projectName.errors.minlength && (projectName.dirty || projectName.touched)\">\n              {{'PROJECT.NAME_MINIMUM_LENGTH' | translate}}\n            </span>\n          </label>\n        </div>\n        <div class=\"form-group\">\n          <label class=\"col-md-4\">{{'PROJECT.PUBLIC_OR_PRIVATE' | translate}}</label>\n          <div class=\"checkbox-inline\">\n            <input type=\"checkbox\" id=\"create_project_public\" [(ngModel)]=\"project.public\" name=\"public\">\n            <label for=\"create_project_public\"></label>\n          </div>\n        </div>\n      </section>\n    </form>\n  </div>\n  <div class=\"modal-footer\">\n    <button type=\"button\" class=\"btn btn-outline\" (click)=\"createProjectOpened = false\">{{'BUTTON.CANCEL' | translate}}</button>\n    <button type=\"button\" class=\"btn btn-primary\" [disabled]=\"!projectForm.form.valid\" (click)=\"onSubmit()\">{{'BUTTON.OK' | translate}}</button>\n  </div>\n</clr-modal>\n"

/***/ }),

/***/ 840:
/***/ (function(module, exports) {

module.exports = "<clr-datagrid (clrDgRefresh)=\"refresh($event)\">\n    <clr-dg-column>{{'PROJECT.NAME' | translate}}</clr-dg-column>\n    <clr-dg-column>{{'PROJECT.PUBLIC_OR_PRIVATE' | translate}}</clr-dg-column>\n    <clr-dg-column>{{'PROJECT.REPO_COUNT'| translate}}</clr-dg-column>\n    <clr-dg-column>{{'PROJECT.CREATION_TIME' | translate}}</clr-dg-column>\n    <clr-dg-column>{{'PROJECT.DESCRIPTION' | translate}}</clr-dg-column>\n    <clr-dg-row *ngFor=\"let p of projects\">\n        <clr-dg-cell><a href=\"javascript:void(0)\" (click)=\"goToLink(p.project_id)\">{{p.name}}</a></clr-dg-cell>\n        <clr-dg-cell>{{ (p.public === 1 ? 'PROJECT.PUBLIC' : 'PROJECT.PRIVATE') | translate}}</clr-dg-cell>\n        <clr-dg-cell>{{p.repo_count}}</clr-dg-cell>\n        <clr-dg-cell>{{p.creation_time}}</clr-dg-cell>\n        <clr-dg-cell>\n            {{p.description}}\n            <harbor-action-overflow *ngIf=\"listFullMode\">\n                <a href=\"javascript:void(0)\" class=\"dropdown-item\">{{'PROJECT.NEW_POLICY' | translate}}</a>\n                <a href=\"javascript:void(0)\" class=\"dropdown-item\" (click)=\"toggleProject(p)\">{{'PROJECT.MAKE' | translate}} {{(p.public === 0 ? 'PROJECT.PUBLIC' : 'PROJECT.PRIVATE') | translate}} </a>\n                <div class=\"dropdown-divider\"></div>\n                <a href=\"javascript:void(0)\" class=\"dropdown-item\" (click)=\"deleteProject(p)\">{{'PROJECT.DELETE' | translate}}</a>\n            </harbor-action-overflow>\n        </clr-dg-cell>\n    </clr-dg-row>\n    <clr-dg-footer>\n        {{totalRecordCount || (projects ? projects.length : 0)}} {{'PROJECT.ITEMS' | translate}}\n        <clr-dg-pagination [clrDgPageSize]=\"pageOffset\" [clrDgTotalItems]=\"totalPage\"></clr-dg-pagination>\n    </clr-dg-footer>\n</clr-datagrid>"

/***/ }),

/***/ 841:
/***/ (function(module, exports) {

module.exports = "<clr-modal [(clrModalOpen)]=\"addMemberOpened\">\n  <h3 class=\"modal-title\">{{'MEMBER.NEW_MEMBER' | translate}}</h3>\n  <div class=\"modal-body\">\n    <form #memberForm=\"ngForm\">\n      <section class=\"form-block\">\n        <clr-alert [clrAlertType]=\"'alert-danger'\" [(clrAlertClosed)]=\"!errorMessageOpened\" (clrAlertClosedChange)=\"onErrorMessageClose()\">\n          <div class=\"alert-item\">\n            <span class=\"alert-text\">\n              {{errorMessage}}\n            </span>\n          </div>\n        </clr-alert>\n        <div class=\"form-group\">\n          <label for=\"member_name\" class=\"col-md-4\">{{'MEMBER.NAME' | translate}}</label>\n          <label for=\"member_name\" aria-haspopup=\"true\" role=\"tooltip\" [class.invalid]=\"memberName.invalid && (memberName.dirty || memberName.touched)\" [class.valid]=\"memberName.valid\" class=\"tooltip tooltip-validation tooltip-sm tooltip-bottom-right\">\n            <input type=\"text\" id=\"member_name\"  [(ngModel)]=\"member.username\" name=\"name\" size=\"20\" #memberName=\"ngModel\" required>\n            <span class=\"tooltip-content\" *ngIf=\"memberName.errors && memberName.errors.required && (memberName.dirty || memberName.touched)\">\n              Username is required.\n            </span>\n          </label>\n        </div>\n        <div class=\"form-group\">\n          <label class=\"col-md-4\">{{'MEMBER.ROLE' | translate}}</label>\n          <div class=\"radio\">\n            <input type=\"radio\" name=\"roleRadios\" id=\"checkrads_project_admin\" (click)=\"member.role_id = 1\" [checked]=\"member.role_id === 1\">\n            <label for=\"checkrads_project_admin\">{{'MEMBER.PROJECT_ADMIN' | translate}}</label>\n          </div>\n          <div class=\"radio\">\n            <input type=\"radio\" name=\"roleRadios\" id=\"checkrads_developer\"  (click)=\"member.role_id = 2\" [checked]=\"member.role_id === 2\">\n            <label for=\"checkrads_developer\">{{'MEMBER.DEVELOPER' | translate}}</label>\n          </div>\n          <div class=\"radio\">\n            <input type=\"radio\" name=\"roleRadios\" id=\"checkrads_guest\" (click)=\"member.role_id = 3\" [checked]=\"member.role_id === 3\">\n            <label for=\"checkrads_guest\">{{'MEMBER.GUEST' | translate}}</label>\n          </div>\n        </div>\n      </section>\n    </form>\n  </div>\n  <div class=\"modal-footer\">\n    <button type=\"button\" class=\"btn btn-outline\" (click)=\"addMemberOpened = false\">{{'BUTTON.CANCEL' | translate}}</button>\n    <button type=\"button\" class=\"btn btn-primary\" (click)=\"onSubmit()\">{{'BUTTON.OK' | translate}}</button>\n  </div>\n</clr-modal>\n"

/***/ }),

/***/ 842:
/***/ (function(module, exports) {

module.exports = "<div class=\"row\">\n    <div class=\"col-lg-12 col-md-12 col-sm-12 col-xs-12\">\n        <div class=\"row flex-items-xs-between\">\n            <div class=\"flex-xs-middle\">\n                <button class=\"btn btn-link\" (click)=\"openAddMemberModal()\"><clr-icon shape=\"add\"></clr-icon> {{'MEMBER.NEW_MEMBER' | translate }}</button>\n                <add-member [projectId]=\"projectId\" (added)=\"addedMember($event)\"></add-member>\n            </div>\n            <div class=\"flex-xs-middle\">\n                <grid-filter filterPlaceholder='{{\"MEMBER.FILTER_PLACEHOLDER\" | translate}}' (filter)=\"doSearch($event)\"></grid-filter>\n                <a href=\"javascript:void(0)\" (click)=\"refresh()\">\n                    <clr-icon shape=\"refresh\"></clr-icon>\n                </a>\n            </div>\n        </div>\n        <clr-datagrid>\n            <clr-dg-column>{{'MEMBER.NAME' | translate}}</clr-dg-column>\n            <clr-dg-column>{{'MEMBER.ROLE' | translate}}</clr-dg-column>\n            <clr-dg-row *ngFor=\"let u of members\">\n                <clr-dg-cell>{{u.username}}</clr-dg-cell>\n                <clr-dg-cell>\n                    {{roleInfo[u.role_id] | translate}}\n                    <harbor-action-overflow [hidden]=\"u.user_id === currentUser.user_id\">\n                        <a href=\"javascript:void(0)\" class=\"dropdown-item\" (click)=\"changeRole(u.user_id, 1)\">{{'MEMBER.PROJECT_ADMIN' | translate}}</a>\n                        <a href=\"javascript:void(0)\" class=\"dropdown-item\" (click)=\"changeRole(u.user_id, 2)\">{{'MEMBER.DEVELOPER' | translate}}</a>\n                        <a href=\"javascript:void(0)\" class=\"dropdown-item\" (click)=\"changeRole(u.user_id, 3)\">{{'MEMBER.GUEST' | translate}}</a>\n                        <div class=\"dropdown-divider\"></div>\n                        <a href=\"javascript:void(0)\" class=\"dropdown-item\" (click)=\"deleteMember(u.user_id)\">{{'MEMBER.DELETE' | translate}}</a>\n                    </harbor-action-overflow>\n                </clr-dg-cell>\n            </clr-dg-row>\n            <clr-dg-footer>{{ (members ? members.length : 0) }} {{'MEMBER.ITEMS' | translate}}</clr-dg-footer>\n        </clr-datagrid>\n    </div>\n</div>"

/***/ }),

/***/ 843:
/***/ (function(module, exports) {

module.exports = "<a style=\"display: block;\" [routerLink]=\"['/harbor', 'projects']\">&lt; {{'PROJECT_DETAIL.PROJECTS' | translate}}</a>\n<h1 class=\"display-in-line\">{{currentProject.name}}</h1>\n<nav class=\"subnav\">\n  <ul class=\"nav\">\n    <li class=\"nav-item\">\n      <a class=\"nav-link\" routerLink=\"repository\" routerLinkActive=\"active\">{{'PROJECT_DETAIL.REPOSITORIES' | translate}}</a>\n    </li>\n    <li class=\"nav-item\" *ngIf=\"isSystemAdmin\">\n      <a class=\"nav-link\" routerLink=\"replication\" routerLinkActive=\"active\">{{'PROJECT_DETAIL.REPLICATION' | translate}}</a>\n    </li>\n    <li class=\"nav-item\">\n      <a class=\"nav-link\" routerLink=\"member\" routerLinkActive=\"active\">{{'PROJECT_DETAIL.USERS' | translate}}</a>\n    </li>\n    <li class=\"nav-item\">\n      <a class=\"nav-link\" routerLink=\"log\" routerLinkActive=\"active\">{{'PROJECT_DETAIL.LOGS' | translate}}</a>\n    </li>\n  </ul>\n</nav>\n<router-outlet></router-outlet>\n"

/***/ }),

/***/ 844:
/***/ (function(module, exports) {

module.exports = "<h1>{{'PROJECT.PROJECTS' | translate}}</h1>\n<div class=\"row flex-items-xs-between\">\n  <div class=\"flex-items-xs-middle\">\n    <button class=\"btn btn-link\" (click)=\"openModal()\"><clr-icon shape=\"add\"></clr-icon> {{'PROJECT.NEW_PROJECT' | translate}}</button>\n    <create-project (create)=\"createProject($event)\"></create-project>\n  </div>\n  <div class=\"flex-items-xs-middle\">\n    <clr-dropdown [clrMenuPosition]=\"'bottom-left'\">\n      <button class=\"btn btn-link\" clrDropdownToggle>\n        {{projectTypes[currentFilteredType] | translate}}\n        <clr-icon shape=\"caret down\"></clr-icon>\n      </button>\n      <div class=\"dropdown-menu\">\n        <a href=\"javascript:void(0)\" clrDropdownItem (click)=\"doFilterProjects(0)\">{{projectTypes[0] | translate}}</a>\n        <a href=\"javascript:void(0)\" clrDropdownItem (click)=\"doFilterProjects(1)\">{{projectTypes[1] | translate}}</a>\n      </div>\n    </clr-dropdown>\n    <grid-filter filterPlaceholder='{{\"PROJECT.FILTER_PLACEHOLDER\" | translate}}' (filter)=\"doSearchProjects($event)\"></grid-filter>\n    <a href=\"javascript:void(0)\" (click)=\"refresh()\"><clr-icon shape=\"refresh\"></clr-icon></a>\n  </div>\n  <div class=\"col-lg-12 col-md-12 col-sm-12 col-xs-12\">\n    <list-project [projects]=\"changedProjects\" (toggle)=\"toggleProject($event)\" (delete)=\"deleteProject($event)\" (paginate)=\"retrieve($event)\" [totalPage]=\"totalPage\" [totalRecordCount]=\"totalRecordCount\"></list-project>\n  </div>\n</div>"

/***/ }),

/***/ 845:
/***/ (function(module, exports) {

module.exports = "<clr-modal [(clrModalOpen)]=\"createEditDestinationOpened\">\n  <h3 class=\"modal-title\">{{modalTitle}}</h3>\n  <div class=\"modal-body\">\n    <form #targetForm=\"ngForm\">\n      <section class=\"form-block\">\n        <clr-alert [clrAlertType]=\"'alert-danger'\" [(clrAlertClosed)]=\"!errorMessageOpened\" (clrAlertClosedChange)=\"onErrorMessageClose()\">\n          <div class=\"alert-item\">\n            <span class=\"alert-text\">\n              {{errorMessage}}\n            </span>\n          </div>\n        </clr-alert>\n        <div class=\"form-group\">\n          <label for=\"destination_name\" class=\"col-md-4\">{{ 'DESTINATION.NAME' | translate }}<span style=\"color: red\">*</span></label>\n          <label class=\"col-md-8\" for=\"destination_name\" aria-haspopup=\"true\" role=\"tooltip\" [class.invalid]=\"targetName.errors && (targetName.dirty || targetName.touched)\" [class.valid]=\"targetName.valid\" class=\"tooltip tooltip-validation tooltip-sm tooltip-bottom-right\">\n            <input type=\"text\" id=\"destination_name\" [disabled]=\"testOngoing\" [(ngModel)]=\"target.name\" name=\"targetName\" size=\"20\" #targetName=\"ngModel\" value=\"\" required> \n            <span class=\"tooltip-content\" *ngIf=\"targetName.errors && targetName.errors.required && (targetName.dirty || targetName.touched)\">\n              {{ 'DESTINATION.NAME_IS_REQUIRED' | translate }}\n            </span>\n          </label>\n        </div>\n        <div class=\"form-group\">\n          <label for=\"destination_url\" class=\"col-md-4\">{{ 'DESTINATION.URL' | translate }}<span style=\"color: red\">*</span></label>\n          <label class=\"col-md-8\" for=\"destination_url\" aria-haspopup=\"true\" role=\"tooltip\" [class.invalid]=\"targetEndpoint.errors && (targetEndpoint.dirty || targetEndpoint.touched)\" [class.valid]=\"targetEndpoint.valid\" class=\"tooltip tooltip-validation tooltip-sm tooltip-bottom-right\">\n            <input type=\"text\" id=\"destination_url\" [disabled]=\"testOngoing\" [(ngModel)]=\"target.endpoint\" size=\"20\" name=\"endpointUrl\" #targetEndpoint=\"ngModel\" required>\n            <span class=\"tooltip-content\" *ngIf=\"targetEndpoint.errors && targetEndpoint.errors.required && (targetEndpoint.dirty || targetEndpoint.touched)\">\n              {{ 'DESTINATION.URL_IS_REQUIRED' | translate }}\n            </span>\n          </label>\n        </div>\n        <div class=\"form-group\">\n          <label for=\"destination_username\" class=\"col-md-4\">{{ 'DESTINATION.USERNAME' | translate }}</label>\n          <input type=\"text\" class=\"col-md-8\" id=\"destination_username\" [disabled]=\"testOngoing\" [(ngModel)]=\"target.username\" size=\"20\" name=\"username\" #username=\"ngModel\">\n        </div>\n        <div class=\"form-group\">\n          <label for=\"destination_password\" class=\"col-md-4\">{{ 'DESTINATION.PASSWORD' | translate }}</label>\n          <input type=\"password\" class=\"col-md-8\" id=\"destination_password\" [disabled]=\"testOngoing\" [(ngModel)]=\"target.password\" size=\"20\" name=\"password\" #password=\"ngModel\">\n        </div>\n        <div class=\"form-group\">\n          <label for=\"spin\" class=\"col-md-4\"></label>\n          <span class=\"col-md-8 spinner spinner-inline\" [hidden]=\"!testOngoing\"></span>\n          <span [style.color]=\"!pingStatus ? 'red': ''\">{{ pingTestMessage }}</span>\n        </div>\n      </section>\n    </form>\n  </div>\n  <div class=\"modal-footer\">\n      <button type=\"button\" class=\"btn btn-outline\" (click)=\"testConnection()\" [disabled]=\"testOngoing || targetEndpoint.errors\">{{ 'DESTINATION.TEST_CONNECTION' | translate }}</button>\n      <button type=\"button\" class=\"btn btn-outline\" (click)=\"createEditDestinationOpened = false\"  [disabled]=\"testOngoing\">{{ 'BUTTON.CANCEL' | translate }}</button>\n      <button type=\"submit\" class=\"btn btn-primary\" [disabled]=\"!targetForm.form.valid\" (click)=\"onSubmit()\"  [disabled]=\"testOngoing\">{{ 'BUTTON.OK' | translate }}</button>\n  </div>\n</clr-modal>"

/***/ }),

/***/ 846:
/***/ (function(module, exports) {

module.exports = "<div class=\"row\">\n  <div class=\"col-lg-12 col-md-12 col-sm-12 col-xs-12\">\n    <div class=\"row flex-items-xs-between\">\n      <div class=\"flex-items-xs-middle\">\n        <button class=\"btn btn-link\" (click)=\"openModal()\"><clr-icon shape=\"add\"></clr-icon> {{'DESTINATION.NEW_ENDPOINT' | translate}}</button>\n        <create-edit-destination (reload)=\"reload($event)\"></create-edit-destination>\n      </div>\n      <div class=\"flex-items-xs-middle\">\n        <grid-filter filterPlaceholder='{{\"REPLICATION.FILTER_TARGETS_PLACEHOLDER\" | translate}}' (filter)=\"doSearchTargets($event)\"></grid-filter>\n        <a href=\"javascript:void(0)\" (click)=\"refreshTargets()\"><clr-icon shape=\"refresh\"></clr-icon></a>\n      </div>\n    </div>\n    <clr-datagrid>\n      <clr-dg-column>{{'DESTINATION.NAME' | translate}}</clr-dg-column>\n      <clr-dg-column>{{'DESTINATION.URL' | translate}}</clr-dg-column>\n      <clr-dg-column>{{'DESTINATION.CREATION_TIME' | translate}}</clr-dg-column>\n      <clr-dg-row *ngFor=\"let t of targets\">\n        <clr-dg-cell>{{t.name}}</clr-dg-cell>\n        <clr-dg-cell>{{t.endpoint}}</clr-dg-cell>\n        <clr-dg-cell>{{t.creation_time}}\n          <harbor-action-overflow>\n            <a href=\"javascript:void(0)\" class=\"dropdown-item\" (click)=\"editTarget(t)\">{{'DESTINATION.TITLE_EDIT' | translate}}</a>\n            <a href=\"javascript:void(0)\" class=\"dropdown-item\" (click)=\"deleteTarget(t)\">{{'DESTINATION.DELETE' | translate}}</a>\n          </harbor-action-overflow>\n        </clr-dg-cell>\n      </clr-dg-row>\n      <clr-dg-footer>{{ (targets ? targets.length : 0) }} {{'DESTINATION.ITEMS' | translate}}</clr-dg-footer>\n    </clr-datagrid>\n  </div>\n</div>"

/***/ }),

/***/ 847:
/***/ (function(module, exports) {

module.exports = "<clr-datagrid (clrDgRefresh)=\"refresh($event)\"> \n  <clr-dg-column>{{'REPLICATION.NAME' | translate}}</clr-dg-column>\n  <clr-dg-column>{{'REPLICATION.STATUS' | translate}}</clr-dg-column>\n  <clr-dg-column>{{'REPLICATION.OPERATION' | translate}}</clr-dg-column>      \n  <clr-dg-column>{{'REPLICATION.CREATION_TIME' | translate}}</clr-dg-column>\n  <clr-dg-column>{{'REPLICATION.END_TIME' | translate}}</clr-dg-column>\n  <clr-dg-column>{{'REPLICATION.LOGS' | translate}}</clr-dg-column>\n  <clr-dg-row *ngFor=\"let j of jobs\">\n    <clr-dg-cell>{{j.repository}}</clr-dg-cell>\n    <clr-dg-cell>{{j.status}}</clr-dg-cell>\n    <clr-dg-cell>{{j.operation}}</clr-dg-cell>\n    <clr-dg-cell>{{j.creation_time}}</clr-dg-cell>\n    <clr-dg-cell>{{j.update_time}}</clr-dg-cell>\n    <clr-dg-cell><a href=\"/api/jobs/replication/{{j.id}}/log\" target=\"_BLANK\"><clr-icon shape=\"clipboard\"></clr-icon></a></clr-dg-cell>\n  </clr-dg-row>\n  <clr-dg-footer>\n    {{ totalRecordCount }} {{'REPLICATION.ITEMS' | translate}} \n    <clr-dg-pagination [clrDgPageSize]=\"pageOffset\" [clrDgTotalItems]=\"totalPage\"></clr-dg-pagination>\n  </clr-dg-footer>\n</clr-datagrid>"

/***/ }),

/***/ 848:
/***/ (function(module, exports) {

module.exports = "<h2>{{'SIDE_NAV.SYSTEM_MGMT.REPLICATION' | translate}}</h2>\n<nav class=\"subnav\">\n  <ul class=\"nav\">\n    <li class=\"nav-item\">\n      <a class=\"nav-link\" routerLink=\"endpoints\" routerLinkActive=\"active\">{{'REPLICATION.ENDPOINTS' | translate}}</a>\n    </li>\n    <li class=\"nav-item\">\n      <a class=\"nav-link\" routerLink=\"rules\" routerLinkActive=\"active\">{{'REPLICATION.REPLICATION_RULE' | translate}}</a>\n    </li>\n  </ul>\n</nav>\n<router-outlet></router-outlet>\n"

/***/ }),

/***/ 849:
/***/ (function(module, exports) {

module.exports = "<div class=\"row\">\n  <div class=\"col-lg-12 col-md-12 col-sm-12 col-xs-12\">\n    <div class=\"row flex-items-xs-between\">\n      <div class=\"flex-xs-middle\">\n        <button class=\"btn btn-link\" (click)=\"openModal()\"><clr-icon shape=\"add\"></clr-icon> {{'REPLICATION.NEW_REPLICATION_RULE' | translate}}</button>\n        <create-edit-policy [projectId]=\"projectId\" (reload)=\"reloadPolicies($event)\"></create-edit-policy>\n      </div>\n      <div class=\"flex-xs-middle\">\n        <clr-dropdown [clrMenuPosition]=\"'bottom-left'\">\n          <button class=\"btn btn-link\" clrDropdownToggle>\n              {{currentRuleStatus.description | translate}}\n            <clr-icon shape=\"caret down\"></clr-icon>\n          </button>\n          <div class=\"dropdown-menu\">\n            <a href=\"javascript:void(0)\" clrDropdownItem *ngFor=\"let r of ruleStatus\" (click)=\"doFilterPolicyStatus(r.key)\"> {{r.description | translate}}</a>\n          </div>\n        </clr-dropdown>\n        <grid-filter filterPlaceholder='{{\"REPLICATION.FILTER_POLICIES_PLACEHOLDER\" | translate}}' (filter)=\"doSearchPolicies($event)\"></grid-filter>\n        <a href=\"javascript:void(0)\" (click)=\"refreshPolicies()\"><clr-icon shape=\"refresh\"></clr-icon></a>\n      </div>\n    </div>\n    <list-policy [policies]=\"changedPolicies\" [projectless]=\"false\" [selectedId]=\"initSelectedId\" (selectOne)=\"selectOne($event)\" (editOne)=\"openEditPolicy($event)\" (reload)=\"reloadPolicies($event)\"></list-policy>\n    <div class=\"row flex-items-xs-between\">\n      <h5 class=\"flex-items-xs-bottom\" style=\"margin-left: 14px;\">{{'REPLICATION.REPLICATION_JOBS' | translate}}</h5>\n      <div class=\"flex-items-xs-bottom\">\n        <button class=\"btn btn-link\" (click)=\"toggleSearchJobOptionalName(currentJobSearchOption)\">{{toggleJobSearchOption[currentJobSearchOption] | translate}}</button>\n        <grid-filter filterPlaceholder='{{\"REPLICATION.FILTER_POLICIES_PLACEHOLDER\" | translate}}' (filter)=\"doSearchJobs($event)\"></grid-filter>\n        <a href=\"javascript:void(0)\" (click)=\"refreshJobs()\"><clr-icon shape=\"refresh\"></clr-icon></a>\n      </div>\n    </div>\n    <div class=\"row flex-items-xs-right\" [hidden]=\"currentJobSearchOption === 0\">\n      <clr-dropdown [clrMenuPosition]=\"'bottom-left'\">\n        <button class=\"btn btn-link\" clrDropdownToggle>\n          {{currentJobStatus.description | translate}}\n          <clr-icon shape=\"caret down\"></clr-icon>\n        </button>\n        <div class=\"dropdown-menu\">\n          <a href=\"javascript:void(0)\" clrDropdownItem *ngFor=\"let j of jobStatus\" (click)=\"doFilterJobStatus(j.key)\"> {{j.description | translate}}</a>\n        </div>\n      </clr-dropdown>\n      <div class=\"flex-items-xs-middle\">\n        <clr-icon shape=\"date\"></clr-icon><input type=\"date\" #fromTime  (change)=\"doJobSearchByTimeRange(fromTime.value, 'begin')\">\n        <clr-icon shape=\"date\"></clr-icon><input type=\"date\" #toTime  (change)=\"doJobSearchByTimeRange(toTime.value, 'end')\">\n      </div>\n    </div>\n    <list-job [jobs]=\"changedJobs\" [totalPage]=\"jobsTotalPage\" [totalRecordCount]=\"jobsTotalRecordCount\" (paginate)=\"fetchPolicyJobs($event)\"></list-job>     \n  </div>\n</div>"

/***/ }),

/***/ 850:
/***/ (function(module, exports) {

module.exports = "<div class=\"row\">\n  <div class=\"col-lg-12 col-md-12 col-sm-12 col-xs-12\">\n    <div class=\"row flex-items-xs-right\">\n      <div class=\"flex-items-xs-middle\">\n        <grid-filter filterPlaceholder='{{\"REPLICATION.FILTER_POLICIES_PLACEHOLDER\" | translate}}' (filter)=\"doSearchPolicies($event)\"></grid-filter>\n        <a href=\"javascript:void(0)\" (click)=\"refreshPolicies()\"><clr-icon shape=\"refresh\"></clr-icon></a>\n      </div>\n    </div>\n    <create-edit-policy [projectId]=\"projectId\" (reload)=\"reloadPolicies($event)\"></create-edit-policy>\n    <list-policy [policies]=\"changedPolicies\" [projectless]=\"true\" (editOne)=\"openEditPolicy($event)\" (selectOne)=\"selectPolicy($event)\" (reload)=\"reloadPolicies($event)\"></list-policy>\n  </div>\n</div>"

/***/ }),

/***/ 851:
/***/ (function(module, exports) {

module.exports = "<clr-datagrid (clrDgRefresh)=\"refresh($event)\">\n    <clr-dg-column>{{'REPOSITORY.NAME' | translate}}</clr-dg-column>\n    <clr-dg-column>{{'REPOSITORY.TAGS_COUNT' | translate}}</clr-dg-column>\n    <clr-dg-column>{{'REPOSITORY.PULL_COUNT' | translate}}</clr-dg-column>\n    <clr-dg-row *ngFor=\"let r of repositories\">\n        <clr-dg-cell><a href=\"javascript:void(0)\" (click)=\"gotoLink(projectId || r.project_id, r.name || r.repository_name)\">{{r.name || r.repository_name}}</a></clr-dg-cell>\n        <clr-dg-cell>{{r.tags_count}}</clr-dg-cell>\n        <clr-dg-cell>{{r.pull_count}}\n            <harbor-action-overflow *ngIf=\"listFullMode\">\n                <a href=\"javascript:void(0)\" class=\"dropdown-item\">{{'REPOSITORY.COPY_ID' | translate}}</a>\n                <a href=\"javascript:void(0)\" class=\"dropdown-item\">{{'REPOSITORY.COPY_PARENT_ID' | translate}}</a>\n                <div class=\"dropdown-divider\"></div>\n                <a href=\"javascript:void(0)\" class=\"dropdown-item\" (click)=\"deleteRepo(r.name)\">{{'REPOSITORY.DELETE' | translate}}</a>\n            </harbor-action-overflow>\n        </clr-dg-cell>\n    </clr-dg-row>\n    <clr-dg-footer>\n        {{totalRecordCount || (repositories ? repositories.length : 0)}} {{'REPOSITORY.ITEMS' | translate}}\n        <clr-dg-pagination [clrDgPageSize]=\"pageOffset\" [clrDgTotalItems]=\"totalPage\"></clr-dg-pagination>\n    </clr-dg-footer>\n</clr-datagrid>"

/***/ }),

/***/ 852:
/***/ (function(module, exports) {

module.exports = "<div class=\"row\">\n  <div class=\"col-lg-12 col-md-12 col-sm-12 col-xs-12\">  \n    <div class=\"row flex-items-xs-right\">\n      <div class=\"flex-xs-middle\">\n        <grid-filter filterPlaceholder=\"{{'REPOSITORY.FILTER_FOR_REPOSITORIES' | translate}}\" (filter)=\"doSearchRepoNames($event)\"></grid-filter>  \n        <a href=\"javascript:void(0)\" (click)=\"refresh()\"><clr-icon shape=\"refresh\"></clr-icon></a>\n      </div>\n    </div>\n    <list-repository [projectId]=\"projectId\" [repositories]=\"changedRepositories\" (delete)=\"deleteRepo($event)\" [totalPage]=\"totalPage\" [totalRecordCount]=\"totalRecordCount\" (paginate)=\"retrieve($event)\"></list-repository>\n  </div>\n</div>"

/***/ }),

/***/ 853:
/***/ (function(module, exports) {

module.exports = "<a [routerLink]=\"['/harbor', 'projects', projectId, 'repository']\">&lt; {{'REPOSITORY.REPOSITORIES' | translate}}</a>\n<h2>{{repoName}} <span class=\"badge\">{{tags ? tags.length : 0}}</span></h2>\n<clr-datagrid>\n   <clr-dg-column>{{'REPOSITORY.TAG' | translate}}</clr-dg-column>\n   <clr-dg-column>{{'REPOSITORY.PULL_COMMAND' | translate}}</clr-dg-column>\n   <clr-dg-column>{{'REPOSITORY.VERIFIED' | translate}}</clr-dg-column>\n   <clr-dg-column>{{'REPOSITORY.AUTHOR' | translate}}</clr-dg-column>\n   <clr-dg-column>{{'REPOSITORY.CREATED' | translate}}</clr-dg-column>\n   <clr-dg-column>{{'REPOSITORY.DOCKER_VERSION' | translate}}</clr-dg-column>\n   <clr-dg-column>{{'REPOSITORY.ARCHITECTURE' | translate}}</clr-dg-column>\n   <clr-dg-column>{{'REPOSITORY.OS' | translate}}</clr-dg-column>\n   <clr-dg-row *ngFor=\"let t of tags\">\n     <clr-dg-cell>{{t.tag}}</clr-dg-cell>\n     <clr-dg-cell>{{t.pullCommand}}</clr-dg-cell>\n     <clr-dg-cell>\n       <clr-icon shape=\"check\" *ngIf=\"t.verified\" style=\"color: #1D5100;\"></clr-icon>\n       <clr-icon shape=\"close\" *ngIf=\"!t.verified\" style=\"color: #C92100;\"></clr-icon>\n      </clr-dg-cell>\n     <clr-dg-cell>{{t.author}}</clr-dg-cell>\n     <clr-dg-cell>{{t.created | date: 'yyyy/MM/dd'}}</clr-dg-cell>\n     <clr-dg-cell>{{t.dockerVersion}}</clr-dg-cell>\n     <clr-dg-cell>{{t.architecture}}</clr-dg-cell>\n     <clr-dg-cell>{{t.os}}\n       <harbor-action-overflow>\n         <a href=\"javascript:void(0)\" class=\"dropdown-item\" (click)=\"deleteTag(t)\">{{'REPOSITORY.DELETE' | translate}}</a>\n       </harbor-action-overflow>\n      </clr-dg-cell>\n    </clr-dg-row>\n  <clr-dg-footer>{{tags ? tags.length : 0}} {{'REPOSITORY.ITEMS' | translate}}</clr-dg-footer>\n</clr-datagrid>"

/***/ }),

/***/ 854:
/***/ (function(module, exports) {

module.exports = "<div class=\"card card-block\">\n    <h3 class=\"card-title\">Popular Repositories</h3>\n    <list-repository [repositories]=\"topRepos\" [mode]=\"listMode\"></list-repository>\n</div>"

/***/ }),

/***/ 855:
/***/ (function(module, exports) {

module.exports = "<clr-modal [(clrModalOpen)]=\"opened\" [clrModalClosable]=\"true\" [clrModalStaticBackdrop]=\"false\">\n    <h3 class=\"modal-title margin-left-override\">vmware</h3>\n    <div class=\"modal-body margin-left-override\">\n        <div class=\"about-product-title\">Harbor</div>\n        <div style=\"height: 12px;\"></div>\n        <div>\n            <span class=\"about-version\">{{'ABOUT.VERSION' | translate}} {{version}}</span>\n            <span>|</span>\n            <span class=\"about-build\">{{'ABOUT.BUILD' | translate}} {{build}}</span>\n        </div>\n        <div style=\"height: 12px;\"></div>\n        <div>\n            <p class=\"about-copyright-text\">{{'ABOUT.COPYRIGHT' | translate}} <a href=\"http://www.vmware.com/go/patents\" target=\"_blank\" class=\"about-text-link\">http://www.vmware.com/go/patents</a></p>\n            <p class=\"about-copyright-text\">{{'ABOUT.TRADEMARK' | translate}}</p>\n            <p>\n                <a href=\"#\" target=\"_blank\" class=\"about-text-link\">{{'ABOUT.END_USER_LICENSE' | translate}}</a><br>\n                <a href=\"#\" target=\"_blank\" class=\"about-text-link\">{{'ABOUT.OPEN_SOURCE_LICENSE' | translate}}</a>\n            </p>\n            <div style=\"height: 24px;\"></div>\n        </div>\n    </div>\n    <div class=\"modal-footer margin-left-override\">\n        <button type=\"button\" class=\"btn btn-primary\" (click)=\"close()\">{{'BUTTON.CLOSE' | translate}}</button>\n    </div>\n</clr-modal>"

/***/ }),

/***/ 856:
/***/ (function(module, exports) {

module.exports = "<clr-modal [(clrModalOpen)]=\"createEditPolicyOpened\">\n  <h3 class=\"modal-title\">{{modalTitle}}</h3>\n  <div class=\"modal-body\">\n    <form #policyForm=\"ngForm\">\n      <section class=\"form-block\">\n        <clr-alert [clrAlertType]=\"'alert-danger'\" [(clrAlertClosed)]=\"!errorMessageOpened\" (clrAlertClosedChange)=\"onErrorMessageClose()\">\n          <div class=\"alert-item\">\n            <span class=\"alert-text\">\n              {{errorMessage}}\n            </span>\n          </div>\n        </clr-alert>\n        <div class=\"form-group\">\n          <label for=\"policy_name\" class=\"col-md-4\">{{'REPLICATION.NAME' | translate}}<span style=\"color: red\">*</span></label>\n          <label for=\"policy_name\" class=\"col-md-8\"  aria-haspopup=\"true\" role=\"tooltip\" [class.invalid]=\"name.errors && (name.dirty || name.touched)\" [class.valid]=\"name.valid\" class=\"tooltip tooltip-validation tooltip-sm tooltip-bottom-right\">\n            <input type=\"text\" id=\"policy_name\" [(ngModel)]=\"createEditPolicy.name\" name=\"name\" #name=\"ngModel\" required>\n            <span class=\"tooltip-content\" *ngIf=\"name.errors && name.errors.required && (name.dirty || name.touched)\">\n              {{'REPLICATION.NAME_IS_REQUIRED'}}\n            </span>\n          </label>\n        </div>\n        <div class=\"form-group\">\n          <label for=\"policy_description\" class=\"col-md-4\">{{'REPLICATION.DESCRIPTION' | translate}}</label>\n          <input type=\"text\" class=\"col-md-8\" id=\"policy_description\" [(ngModel)]=\"createEditPolicy.description\" name=\"description\" size=\"20\" #description=\"ngModel\"> \n        </div>\n        <div class=\"form-group\">\n          <label class=\"col-md-4\">{{'REPLICATION.ENABLE' | translate}}</label>\n          <div class=\"checkbox-inline\">\n            <input type=\"checkbox\" id=\"policy_enable\" [(ngModel)]=\"createEditPolicy.enable\" name=\"enable\" #enable=\"ngModel\">\n            <label for=\"policy_enable\"></label>\n          </div>\n        </div>\n        <div class=\"form-group\">\n          <label for=\"destination_name\" class=\"col-md-4\">{{'REPLICATION.DESTINATION_NAME' | translate}}<span style=\"color: red\">*</span></label>\n          <div class=\"select\" *ngIf=\"!isCreateDestination\">\n            <select id=\"destination_name\" [(ngModel)]=\"createEditPolicy.targetId\" name=\"targetId\" (change)=\"selectTarget()\" [disabled]=\"testOngoing\">\n              <option *ngFor=\"let t of targets\" [value]=\"t.id\" [selected]=\"t.id == createEditPolicy.targetId\">{{t.name}}</option>\n            </select>\n          </div>\n          <label class=\"col-md-8\" *ngIf=\"isCreateDestination\" for=\"destination_name\" aria-haspopup=\"true\" role=\"tooltip\" [class.invalid]=\"targetName.errors && (targetName.dirty || targetName.touched)\" [class.valid]=\"targetName.valid\" class=\"tooltip tooltip-validation tooltip-sm tooltip-bottom-right\">\n            <input type=\"text\" id=\"destination_name\" [(ngModel)]=\"createEditPolicy.targetName\" name=\"targetName\" size=\"20\" #targetName=\"ngModel\" value=\"\" required> \n            <span class=\"tooltip-content\" *ngIf=\"targetName.errors && targetName.errors.required && (targetName.dirty || targetName.touched)\">\n              {{'REPLICATION.DESTINATION_NAME_IS_REQUIRED' | translate}}\n            </span>\n          </label>\n          <div class=\"checkbox-inline\">\n            <input type=\"checkbox\" id=\"check_new\" (click)=\"newDestination(checkedAddNew.checked)\" #checkedAddNew [checked]=\"isCreateDestination\" [disabled]=\"testOngoing\">\n            <label for=\"check_new\">{{'REPLICATION.NEW_DESTINATION' | translate}}</label>\n          </div>\n        </div>\n        <div class=\"form-group\">\n          <label for=\"destination_url\" class=\"col-md-4\">{{'REPLICATION.DESTINATION_URL' | translate}}<span style=\"color: red\">*</span></label>\n          <label for=\"destination_url\" class=\"col-md-8\" aria-haspopup=\"true\" role=\"tooltip\" [class.invalid]=\"endpointUrl.errors && (endpointUrl.dirty || endpointUrl.touched)\" [class.valid]=\"endpointUrl.valid\" class=\"tooltip tooltip-validation tooltip-sm tooltip-bottom-right\">\n            <input type=\"text\" id=\"destination_url\" [disabled]=\"testOngoing\" [(ngModel)]=\"createEditPolicy.endpointUrl\" size=\"20\" name=\"endpointUrl\" required #endpointUrl=\"ngModel\">\n            <span class=\"tooltip-content\" *ngIf=\"endpointUrl.errors && endpointUrl.errors.required && (endpointUrl.dirty || endpointUrl.touched)\">\n              {{'REPLICATION.DESTINATION_URL_IS_REQUIRED' | translate}}\n            </span>\n          </label>\n        </div>\n        <div class=\"form-group\">\n          <label for=\"destination_username\" class=\"col-md-4\">{{'REPLICATION.DESTINATION_USERNAME' | translate}}</label>\n          <input type=\"text\" class=\"col-md-8\" id=\"destination_username\" [disabled]=\"testOngoing\" [(ngModel)]=\"createEditPolicy.username\" size=\"20\" name=\"username\" #username=\"ngModel\">\n        </div>\n        <div class=\"form-group\">\n          <label for=\"destination_password\" class=\"col-md-4\">{{'REPLICATION.DESTINATION_PASSWORD' | translate}}</label>\n          <input type=\"password\" class=\"col-md-8\" id=\"destination_password\" [disabled]=\"testOngoing\" [(ngModel)]=\"createEditPolicy.password\" size=\"20\" name=\"password\" #password=\"ngModel\">\n        </div>\n        <div class=\"form-group\">\n          <label for=\"spin\" class=\"col-md-4\"></label>\n          <span class=\"col-md-8 spinner spinner-inline\" [hidden]=\"!testOngoing\"></span>\n          <span [style.color]=\"!pingStatus ? 'red': ''\">{{ pingTestMessage }}</span>\n        </div>\n      </section>\n    </form>\n  </div>\n  <div class=\"modal-footer\">\n      <button type=\"button\" class=\"btn btn-outline\" (click)=\"testConnection()\" [disabled]=\"testOngoing\">{{'REPLICATION.TEST_CONNECTION' | translate}}</button>\n      <button type=\"button\" class=\"btn btn-outline\" (click)=\"createEditPolicyOpened = false\">{{'BUTTON.CANCEL' | translate }}</button>\n      <button type=\"submit\" class=\"btn btn-primary\" [disabled]=\"!policyForm.form.valid\" (click)=\"onSubmit()\">{{'BUTTON.OK' | translate}}</button>\n  </div>\n</clr-modal>"

/***/ }),

/***/ 857:
/***/ (function(module, exports) {

module.exports = "<clr-modal [(clrModalOpen)]=\"opened\" [clrModalClosable]=\"false\" [clrModalStaticBackdrop]=\"true\">\n    <h3 class=\"modal-title\" class=\"deletion-title\" style=\"margin-top: 0px;\">{{dialogTitle}}</h3>\n    <div class=\"modal-body\">\n        <div class=\"deletion-icon-inline\">\n            <clr-icon shape=\"warning\" class=\"is-warning\" size=\"64\"></clr-icon>\n        </div>\n        <div class=\"deletion-content\">{{dialogContent}}</div>\n    </div>\n    <div class=\"modal-footer\">\n        <button type=\"button\" class=\"btn btn-outline\" (click)=\"close()\">{{'BUTTON.CANCEL' | translate}}</button>\n        <button type=\"button\" class=\"btn btn-primary\" (click)=\"confirm()\">{{'BUTTON.CONFIRM' | translate}}</button>\n    </div>\n</clr-modal>"

/***/ }),

/***/ 858:
/***/ (function(module, exports) {

module.exports = "<span>\n    <clr-icon shape=\"filter\" size=\"12\" class=\"is-solid filter-icon\"></clr-icon>\n    <input type=\"text\" style=\"padding-left: 15px;\" (keyup)=\"valueChange()\" placeholder=\"{{placeHolder}}\" [(ngModel)]=\"currentValue\"/>\n</span>"

/***/ }),

/***/ 859:
/***/ (function(module, exports) {

module.exports = "<span style=\"float: right; margin-right: 24px;\">\n<clr-dropdown [clrMenuPosition]=\"'bottom-right'\" [clrCloseMenuOnItemClick]=\"true\" style=\"position: absolute;\">\n    <button clrDropdownToggle>\n    <clr-icon shape=\"ellipses-vertical\"></clr-icon>\n  </button>\n    <div class=\"dropdown-menu\">\n        <ng-content></ng-content>\n    </div>\n</clr-dropdown>\n</span>"

/***/ }),

/***/ 860:
/***/ (function(module, exports) {

module.exports = "<clr-alert [clrAlertType]=\"inlineAlertType\" [clrAlertClosable]=\"inlineAlertClosable\" [(clrAlertClosed)]=\"alertClose\" [clrAlertAppLevel]=\"useAppLevelStyle\">\n    <div class=\"alert-item\">\n        <span class=\"alert-text\">\n            {{errorMessage}}\n        </span>\n        <div class=\"alert-actions\" *ngIf=\"showCancelAction\">\n            <button class=\"btn alert-action\" (click)=\"confirmCancel()\">{{'BUTTON.CONFIRM' | translate}}</button>\n        </div>\n    </div>\n</clr-alert>"

/***/ }),

/***/ 861:
/***/ (function(module, exports) {

module.exports = "<clr-datagrid>\n  <clr-dg-column>{{'REPLICATION.NAME' | translate}}</clr-dg-column>\n  <clr-dg-column *ngIf=\"projectless\">{{'REPLICATION.PROJECT' | translate}}</clr-dg-column>   \n  <clr-dg-column>{{'REPLICATION.DESCRIPTION' | translate}}</clr-dg-column>\n  <clr-dg-column>{{'REPLICATION.DESTINATION_NAME' | translate}}</clr-dg-column>\n  <clr-dg-column>{{'REPLICATION.LAST_START_TIME' | translate}}</clr-dg-column>   \n  <clr-dg-column>{{'REPLICATION.ACTIVATION' | translate}}</clr-dg-column>\n  <clr-dg-row *ngFor=\"let p of policies;let i = index;\" (click)=\"selectPolicy(p)\" [style.backgroundColor]=\"(!projectless && selectedId === p.id) ? '#eee' : ''\">\n    <clr-dg-cell>{{p.name}}</clr-dg-cell>\n    <clr-dg-cell *ngIf=\"projectless\">{{p.project_name}}</clr-dg-cell>\n    <clr-dg-cell>{{p.description}}</clr-dg-cell>\n    <clr-dg-cell>{{p.target_name}}</clr-dg-cell>\n    <clr-dg-cell>{{p.start_time}}</clr-dg-cell>\n    <clr-dg-cell>\n      {{ (p.enabled === 1 ? 'REPLICATION.ENABLED' : 'REPLICATION.DISABLED') | translate}}\n      <harbor-action-overflow>\n        <a href=\"javascript:void(0)\" class=\"dropdown-item\" (click)=\"editPolicy(p)\">{{'REPLICATION.EDIT_POLICY' | translate}}</a>\n        <a href=\"javascript:void(0)\" class=\"dropdown-item\" (click)=\"enablePolicy(p)\">{{ (p.enabled === 0 ? 'REPLICATION.ENABLE' : 'REPLICATION.DISABLE') | translate}}</a>\n        <a href=\"javascript:void(0)\" class=\"dropdown-item\" (click)=\"deletePolicy(p)\">{{'REPLICATION.DELETE_POLICY' | translate}}</a>\n      </harbor-action-overflow>\n    </clr-dg-cell>\n  </clr-dg-row>\n  <clr-dg-footer>{{ (policies ? policies.length : 0) }} {{'REPLICATION.ITEMS' | translate}}</clr-dg-footer>\n</clr-datagrid>"

/***/ }),

/***/ 862:
/***/ (function(module, exports) {

module.exports = "<div>\n    <form #newUserFrom=\"ngForm\" class=\"form\">\n        <section class=\"form-block\">\n            <div class=\"form-group\">\n                <label for=\"username\" class=\"col-md-4 required\">{{'PROFILE.USER_NAME' | translate}}</label>\n                <label for=\"username\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-bottom-left\" [class.invalid]=\"usernameInput.invalid && (usernameInput.dirty || usernameInput.touched)\">\n                <input type=\"text\" placeholder='{{\"PLACEHOLDER.USER_NAME\" | translate}}' required pattern='[^\"~#$%]+' maxLengthExt=\"20\" #usernameInput=\"ngModel\" name=\"username\" [(ngModel)]=\"newUser.username\" id=\"username\" size=\"28\">\n                <span class=\"tooltip-content\">\n                    {{'TOOLTIP.USER_NAME' | translate}}\n                </span>\n                </label>\n            </div>\n            <div class=\"form-group\">\n                <label for=\"email\" class=\"col-md-4 required\">{{'PROFILE.EMAIL' | translate}}</label>\n                <label for=\"email\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-bottom-left\" [class.invalid]=\"eamilInput.invalid && (eamilInput.dirty || eamilInput.touched)\">\n                      <input name=\"email\" type=\"text\" #eamilInput=\"ngModel\" [(ngModel)]=\"newUser.email\" \n                      placeholder='{{\"PLACEHOLDER.MAIL\" | translate}}'\n                      required \n                      pattern='^[a-zA-Z0-9.!#$%&’*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\\.[a-zA-Z0-9-]+)*$' id=\"email\" size=\"28\">\n                      <span class=\"tooltip-content\">\n                          {{'TOOLTIP.EMAIL' | translate}}\n                      </span>\n                </label>\n                <label *ngIf=\"isSelfRegistration\" role=\"tooltip\" aria-haspopup=\"true\" class=\"tooltip tooltip-bottom-left\">\n                    <clr-icon shape=\"info\" class=\"is-info\" size=\"24\"></clr-icon>\n                    <span class=\"tooltip-content\">\n                        {{'TOOLTIP.SIGN_UP_MAIL' | translate}}\n                    </span>\n                </label>\n            </div>\n            <div class=\"form-group\">\n                <label for=\"realname\" class=\"col-md-4 required\">{{'PROFILE.FULL_NAME' | translate}}</label>\n                <label for=\"realname\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-bottom-left\" [class.invalid]=\"fullNameInput.invalid && (fullNameInput.dirty || fullNameInput.touched)\">\n                      <input type=\"text\" placeholder='{{\"PLACEHOLDER.FULL_NAME\" | translate}}' name=\"realname\" #fullNameInput=\"ngModel\" [(ngModel)]=\"newUser.realname\" required maxLengthExt=\"20\" id=\"realname\" size=\"28\">\n                      <span class=\"tooltip-content\">\n                          {{'TOOLTIP.FULL_NAME' | translate}}\n                      </span>\n                </label>\n                <label *ngIf=\"isSelfRegistration\" role=\"tooltip\" aria-haspopup=\"true\" class=\"tooltip tooltip-bottom-left\">\n                    <clr-icon shape=\"info\" class=\"is-info\" size=\"24\"></clr-icon>\n                    <span class=\"tooltip-content\">\n                        {{'TOOLTIP.SIGN_UP_REAL_NAME' | translate}}\n                    </span>\n                </label>\n            </div>\n            <div class=\"form-group\">\n                <label for=\"newPassword\" class=\"required\">{{'PROFILE.PASSWORD' | translate}}</label>\n                <label for=\"newPassword\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-bottom-left\" [class.invalid]=\"newPassInput.invalid && (newPassInput.dirty || newPassInput.touched)\">\n                <input type=\"password\" id=\"newPassword\" placeholder='{{\"PLACEHOLDER.NEW_PWD\" | translate}}'\n                    required\n                    pattern=\"^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)[a-zA-Z\\d]{7,}$\"\n                    name=\"newPassword\"\n                    [(ngModel)]=\"newUser.password\"\n                    #newPassInput=\"ngModel\" size=\"28\">\n                <span class=\"tooltip-content\">\n                    {{'TOOLTIP.PASSWORD' | translate}}\n                </span>\n                </label>\n                <label *ngIf=\"isSelfRegistration\" role=\"tooltip\" aria-haspopup=\"true\" class=\"tooltip tooltip-bottom-left\">\n                    <clr-icon shape=\"info\" class=\"is-info\" size=\"24\"></clr-icon>\n                    <span class=\"tooltip-content\">\n                        {{'TOOLTIP.PASSWORD' | translate}}\n                    </span>\n                </label>\n            </div>\n            <div class=\"form-group\">\n                <label for=\"confirmPassword\" class=\"required\">{{'CHANGE_PWD.CONFIRM_PWD' | translate}}</label>\n                <label for=\"confirmPassword\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-bottom-left\" [class.invalid]=\"(confirmPassInput.invalid && (confirmPassInput.dirty || confirmPassInput.touched)) || (!confirmPassInput.invalid && confirmPassInput.value != newPassInput.value)\">\n                <input type=\"password\" id=\"confirmPassword\" placeholder='{{\"PLACEHOLDER.CONFIRM_PWD\" | translate}}'\n                    required\n                    pattern=\"^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)[a-zA-Z\\d]{7,}$\"\n                    name=\"confirmPassword\"\n                    [(ngModel)]=\"confirmedPwd\"\n                    #confirmPassInput=\"ngModel\" size=\"28\">\n                <span class=\"tooltip-content\">\n                    {{'TOOLTIP.CONFIRM_PWD' | translate}}\n                </span>\n            </label>\n            </div>\n            <div class=\"form-group\">\n                <label for=\"comment\" class=\"col-md-4\">{{'PROFILE.COMMENT' | translate}}</label>\n                <label for=\"comment\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-bottom-left\" [class.invalid]=\"commentInput.invalid && (commentInput.dirty || commentInput.touched)\">\n                <input type=\"text\" #commentInput=\"ngModel\" name=\"comment\" [(ngModel)]=\"newUser.comment\" maxLengthExt=\"20\" id=\"comment\" size=\"28\">\n                <span class=\"tooltip-content\">\n                    {{'TOOLTIP.COMMENT' | translate}}\n                </span>\n                </label>\n            </div>\n        </section>\n    </form>\n</div>\n<div style=\"height: 15px;\"></div>"

/***/ }),

/***/ 863:
/***/ (function(module, exports) {

module.exports = "<div class=\"wrapper-back\">\n    <div>\n        <clr-icon shape=\"warning\" class=\"is-warning\" size=\"96\"></clr-icon>\n        <span class=\"status-code\">404</span>\n        <span class=\"status-text\">{{'PAGE_NOT_FOUND.MAIN_TITLE' | translate}}</span>\n    </div>\n    <div class=\"status-subtitle\">\n        {{'PAGE_NOT_FOUND.SUB_TITLE' | translate}} <span class=\"second-number\">{{leftSeconds}}</span> {{'PAGE_NOT_FOUND.UNIT' | translate}}\n    </div>\n</div>"

/***/ }),

/***/ 864:
/***/ (function(module, exports) {

module.exports = "<div class=\"card card-block\">\n    <h3 class=\"card-title\">{{'STATISTICS.TITLE' | translate }}</h3>\n    <span class=\"card-text\">\n        <div class=\"row\">\n            <div class=\"col-xs-2 col-sm-2 col-md-2 col-lg-2 col-xl-2\">\n<span class=\"statistic-column-title\">{{'STATISTICS.PRO_ITEM' | translate }}</span>\n</div>\n<div class=\"col-xs-10 col-sm-10 col-md-10 col-lg-10 col-xl-10\">\n    <statistics [data]='{number: originalCopy.my_project_count, label: \"my\"}'></statistics>\n    <statistics [data]='{number: originalCopy.public_project_count, label: \"pub\"}'></statistics>\n    <statistics [data]='{number: originalCopy.total_project_count, label: \"total\"}' *ngIf=\"isValidSession\"></statistics>\n</div>\n</div>\n<div class=\"row\">\n    <div class=\"col-xs-2 col-sm-2 col-md-2 col-lg-2 col-xl-2\">\n        <span class=\"statistic-column-title\">{{'STATISTICS.REPO_ITEM' | translate }}</span>\n    </div>\n    <div class=\"col-xs-10 col-sm-10 col-md-10 col-lg-10 col-xl-10\">\n        <statistics [data]='{number: originalCopy.my_repo_count, label: \"my\"}'></statistics>\n        <statistics [data]='{number: originalCopy.public_repo_count, label: \"pub\"}'></statistics>\n        <statistics [data]='{number: originalCopy.total_repo_count, label: \"total\"}' *ngIf=\"isValidSession\"></statistics>\n    </div>\n</div>\n</span>\n</div>"

/***/ }),

/***/ 865:
/***/ (function(module, exports) {

module.exports = "<div class=\"statistic-wrapper\">\n    <span class=\"statistic-data\">{{data.number}}</span>\n    <span class=\"statistic-text\">{{data.label}}</span>\n</div>"

/***/ }),

/***/ 866:
/***/ (function(module, exports) {

module.exports = "<clr-modal [(clrModalOpen)]=\"opened\" [clrModalStaticBackdrop]=\"staticBackdrop\">\n    <h3 class=\"modal-title\">{{'USER.ADD_USER_TITLE' | translate}}</h3>\n    <div class=\"modal-body\">\n        <new-user-form (valueChange)=\"formValueChange($event)\"></new-user-form>\n        <inline-alert (confirmEvt)=\"confirmCancel($event)\"></inline-alert>\n    </div>\n    <div class=\"modal-footer\">\n        <span class=\"spinner spinner-inline\" style=\"top:8px;\" [hidden]=\"inProgress === false\"> </span>\n        <button type=\"button\" class=\"btn btn-outline\" (click)=\"close()\">{{'BUTTON.CANCEL' | translate}}</button>\n        <button type=\"button\" class=\"btn btn-primary\" [disabled]=\"!isValid || inProgress\" (click)=\"create()\">{{'BUTTON.OK' | translate}}</button>\n    </div>\n</clr-modal>"

/***/ }),

/***/ 867:
/***/ (function(module, exports) {

module.exports = "<div>\n    <h2 class=\"custom-h2\">{{'SIDE_NAV.SYSTEM_MGMT.USER' | translate}}</h2>\n    <div class=\"action-panel-pos\">\n        <span>\n            <clr-icon shape=\"plus\" class=\"is-highlight\" size=\"24\"></clr-icon>\n            <button type=\"submit\" class=\"btn btn-link custom-add-button\" (click)=\"addNewUser()\">{{'USER.ADD_ACTION' | translate}}</button>\n        </span>\n        <grid-filter class=\"filter-pos\" filterPlaceholder='{{\"USER.FILTER_PLACEHOLDER\" | translate}}' (filter)=\"doFilter($event)\"></grid-filter>\n        <span class=\"refresh-btn\" (click)=\"refreshUser()\">\n            <clr-icon shape=\"refresh\" [hidden]=\"inProgress\" ng-disabled=\"inProgress\"></clr-icon>\n            <span class=\"spinner spinner-inline\" [hidden]=\"inProgress === false\"></span>\n        </span>\n    </div>\n    <div>\n        <clr-datagrid>\n            <clr-dg-column>{{'USER.COLUMN_NAME' | translate}}</clr-dg-column>\n            <clr-dg-column>{{'USER.COLUMN_ADMIN' | translate}}</clr-dg-column>\n            <clr-dg-column>{{'USER.COLUMN_EMAIL' | translate}}</clr-dg-column>\n            <clr-dg-column>{{'USER.COLUMN_REG_NAME' | translate}}</clr-dg-column>\n            <clr-dg-row *clrDgItems=\"let user of users\" [clrDgItem]=\"user\">\n                <clr-dg-cell>{{user.username}}</clr-dg-cell>\n                <clr-dg-cell>{{isSystemAdmin(user)}}</clr-dg-cell>\n                <clr-dg-cell>{{user.email}}</clr-dg-cell>\n                <clr-dg-cell>\n                    {{user.creation_time}}\n                    <harbor-action-overflow>\n                        <a href=\"javascript:void(0)\" class=\"dropdown-item\" (click)=\"changeAdminRole(user)\">{{adminActions(user)}}</a>\n                        <div class=\"dropdown-divider\"></div>\n                        <a href=\"javascript:void(0)\" class=\"dropdown-item\" (click)=\"deleteUser(user)\">{{'USER.DEL_ACTION' | translate}}</a>\n                    </harbor-action-overflow>\n                </clr-dg-cell>\n            </clr-dg-row>\n            <clr-dg-footer>{{users.length}} {{'USER.ADD_ACTION' | translate}}</clr-dg-footer>\n        </clr-datagrid>\n    </div>\n    <new-user-modal (addNew)=\"addUserToList($event)\"></new-user-modal>\n</div>"

/***/ }),

/***/ 904:
/***/ (function(module, exports, __webpack_require__) {

module.exports = __webpack_require__(478);


/***/ }),

/***/ 95:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var core_1 = __webpack_require__(0);
var Subject_1 = __webpack_require__(25);
var SearchTriggerService = (function () {
    function SearchTriggerService() {
        this.searchTriggerSource = new Subject_1.Subject();
        this.searchCloseSource = new Subject_1.Subject();
        this.searchInputSource = new Subject_1.Subject();
        this.searchTriggerChan$ = this.searchTriggerSource.asObservable();
        this.searchCloseChan$ = this.searchCloseSource.asObservable();
        this.searchInputChan$ = this.searchInputSource.asObservable();
    }
    SearchTriggerService.prototype.triggerSearch = function (event) {
        this.searchTriggerSource.next(event);
    };
    //Set event to true for shell
    //set to false for search panel
    SearchTriggerService.prototype.closeSearch = function (event) {
        this.searchCloseSource.next(event);
    };
    //Notify the state change of search box in home start page
    SearchTriggerService.prototype.searchInputStat = function (event) {
        this.searchInputSource.next(event);
    };
    SearchTriggerService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [])
    ], SearchTriggerService);
    return SearchTriggerService;
}());
exports.SearchTriggerService = SearchTriggerService;
//# sourceMappingURL=/Users/vmware/harbor-clarity/harbor-app/src/search-trigger.service.js.map

/***/ })

},[904]);
//# sourceMappingURL=main.bundle.map