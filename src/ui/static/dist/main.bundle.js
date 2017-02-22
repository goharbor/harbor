webpackJsonp([0,4],{

/***/ 156:
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
var http_1 = __webpack_require__(79);
var base_service_1 = __webpack_require__(551);
var Observable_1 = __webpack_require__(8);
__webpack_require__(746);
__webpack_require__(747);
__webpack_require__(745);
var url_prefix = '';
var ProjectService = (function (_super) {
    __extends(ProjectService, _super);
    function ProjectService(http) {
        _super.call(this);
        this.http = http;
        this.headers = new http_1.Headers({ 'Content-type': 'application/json' });
        this.options = new http_1.RequestOptions({ 'headers': this.headers });
    }
    ProjectService.prototype.listProjects = function (name, isPublic) {
        return this.http
            .get(url_prefix + ("/api/projects?project_name=" + name + "&is_public=" + isPublic), this.options)
            .map(function (response) { return response.json(); })
            .catch(this.handleError);
    };
    ProjectService.prototype.createProject = function (name, isPublic) {
        return this.http
            .post(url_prefix + "/api/projects", JSON.stringify({ 'project_name': name, 'public': isPublic }), this.options)
            .map(function (response) { return response.status; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    ProjectService.prototype.toggleProjectPublic = function (projectId, isPublic) {
        return this.http
            .put(url_prefix + ("/api/projects/" + projectId + "/publicity"), { 'public': isPublic }, this.options)
            .map(function (response) { return response.status; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    ProjectService.prototype.deleteProject = function (projectId) {
        return this.http
            .delete(url_prefix + ("/api/projects/" + projectId))
            .map(function (response) { return response.status; })
            .catch(function (error) { return Observable_1.Observable.throw(error); });
    };
    ProjectService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof http_1.Http !== 'undefined' && http_1.Http) === 'function' && _a) || Object])
    ], ProjectService);
    return ProjectService;
    var _a;
}(base_service_1.BaseService));
exports.ProjectService = ProjectService;
//# sourceMappingURL=/clarity-seed/src/src/project.service.js.map

/***/ }),

/***/ 222:
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
var HarborShellComponent = (function () {
    function HarborShellComponent() {
    }
    HarborShellComponent = __decorate([
        core_1.Component({
            selector: 'harbor-shell',
            template: __webpack_require__(721)
        }), 
        __metadata('design:paramtypes', [])
    ], HarborShellComponent);
    return HarborShellComponent;
}());
exports.HarborShellComponent = HarborShellComponent;
//# sourceMappingURL=/clarity-seed/src/src/harbor-shell.component.js.map

/***/ }),

/***/ 223:
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
var Subject_1 = __webpack_require__(46);
var MessageService = (function () {
    function MessageService() {
        this.messageAnnouncedSource = new Subject_1.Subject();
        this.messageAnnounced$ = this.messageAnnouncedSource.asObservable();
    }
    MessageService.prototype.announceMessage = function (message) {
        this.messageAnnouncedSource.next(message);
    };
    MessageService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [])
    ], MessageService);
    return MessageService;
}());
exports.MessageService = MessageService;
//# sourceMappingURL=/clarity-seed/src/src/message.service.js.map

/***/ }),

/***/ 224:
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
var list_project_component_1 = __webpack_require__(347);
var create_project_component_1 = __webpack_require__(346);
var ProjectComponent = (function () {
    function ProjectComponent() {
        this.lastFilteredType = 0;
    }
    ProjectComponent.prototype.openModal = function () {
        this.creationProject.newProject();
    };
    ProjectComponent.prototype.deleteSelectedProjects = function () {
        this.listProjects.deleteSelectedProjects();
    };
    ProjectComponent.prototype.createProject = function (created) {
        console.log('Project has been created:' + created);
        this.listProjects.retrieve('', 0);
    };
    ProjectComponent.prototype.filterProjects = function (type) {
        this.lastFilteredType = type;
        this.listProjects.retrieve('', type);
        console.log('Projects were filtered by:' + type);
    };
    ProjectComponent.prototype.searchProjects = function (projectName) {
        console.log('Search for project name:' + projectName);
        this.listProjects.retrieve(projectName, this.lastFilteredType);
    };
    ProjectComponent.prototype.actionPerform = function (performed) {
        this.listProjects.retrieve('', 0);
    };
    ProjectComponent.prototype.ngOnInit = function () {
        this.listProjects.retrieve('', 0);
    };
    __decorate([
        core_1.ViewChild(list_project_component_1.ListProjectComponent), 
        __metadata('design:type', (typeof (_a = typeof list_project_component_1.ListProjectComponent !== 'undefined' && list_project_component_1.ListProjectComponent) === 'function' && _a) || Object)
    ], ProjectComponent.prototype, "listProjects", void 0);
    __decorate([
        core_1.ViewChild(create_project_component_1.CreateProjectComponent), 
        __metadata('design:type', (typeof (_b = typeof create_project_component_1.CreateProjectComponent !== 'undefined' && create_project_component_1.CreateProjectComponent) === 'function' && _b) || Object)
    ], ProjectComponent.prototype, "creationProject", void 0);
    ProjectComponent = __decorate([
        core_1.Component({
            selector: 'project',
            template: __webpack_require__(732),
            styles: [__webpack_require__(714)]
        }), 
        __metadata('design:paramtypes', [])
    ], ProjectComponent);
    return ProjectComponent;
    var _a, _b;
}());
exports.ProjectComponent = ProjectComponent;
//# sourceMappingURL=/clarity-seed/src/src/project.component.js.map

/***/ }),

/***/ 341:
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
var router_1 = __webpack_require__(107);
var core_2 = __webpack_require__(0);
var forms_1 = __webpack_require__(201);
var sign_in_service_1 = __webpack_require__(532);
var sign_in_credential_1 = __webpack_require__(531);
var session_service_1 = __webpack_require__(353);
//Define status flags for signing in states
exports.signInStatusNormal = 0;
exports.signInStatusOnGoing = 1;
exports.signInStatusError = -1;
var SignInComponent = (function () {
    function SignInComponent(signInService, router, session) {
        this.signInService = signInService;
        this.router = router;
        this.session = session;
        //Status flag
        this.signInStatus = 0;
        //Initialize sign in credential
        this.signInCredential = {
            principal: "",
            password: ""
        };
    }
    Object.defineProperty(SignInComponent.prototype, "statusError", {
        //For template accessing
        get: function () {
            return exports.signInStatusError;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(SignInComponent.prototype, "statusOnGoing", {
        get: function () {
            return exports.signInStatusOnGoing;
        },
        enumerable: true,
        configurable: true
    });
    //Validate the related fields
    SignInComponent.prototype.validate = function () {
        return true;
        //return this.signInForm.valid;
    };
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
        if (!this.validate()) {
            console.info("return");
            return;
        }
        //Start signing in progress
        this.signInStatus = exports.signInStatusOnGoing;
        //Call the service to send out the http request
        this.signInService.signIn(this.signInCredential)
            .then(function () {
            //Set status
            _this.signInStatus = exports.signInStatusNormal;
            //Validate the sign-in session
            _this.session.retrieveUser()
                .then(function () {
                //Routing to the right location
                var nextRoute = ["/harbor", "dashboard"];
                _this.router.navigate(nextRoute);
            })
                .catch(_this.handleError);
        })
            .catch(this.handleError);
    };
    //Help user navigate to the sign up
    SignInComponent.prototype.signUp = function () {
        var nextRoute = ["/harbor", "signup"];
        this.router.navigate(nextRoute);
    };
    __decorate([
        core_2.ViewChild('signInForm'), 
        __metadata('design:type', (typeof (_a = typeof forms_1.NgForm !== 'undefined' && forms_1.NgForm) === 'function' && _a) || Object)
    ], SignInComponent.prototype, "currentForm", void 0);
    __decorate([
        core_2.Input(), 
        __metadata('design:type', (typeof (_b = typeof sign_in_credential_1.SignInCredential !== 'undefined' && sign_in_credential_1.SignInCredential) === 'function' && _b) || Object)
    ], SignInComponent.prototype, "signInCredential", void 0);
    SignInComponent = __decorate([
        core_1.Component({
            selector: 'sign-in',
            template: __webpack_require__(716),
            styles: [__webpack_require__(711)],
            providers: [sign_in_service_1.SignInService]
        }), 
        __metadata('design:paramtypes', [(typeof (_c = typeof sign_in_service_1.SignInService !== 'undefined' && sign_in_service_1.SignInService) === 'function' && _c) || Object, (typeof (_d = typeof router_1.Router !== 'undefined' && router_1.Router) === 'function' && _d) || Object, (typeof (_e = typeof session_service_1.SessionService !== 'undefined' && session_service_1.SessionService) === 'function' && _e) || Object])
    ], SignInComponent);
    return SignInComponent;
    var _a, _b, _c, _d, _e;
}());
exports.SignInComponent = SignInComponent;
//# sourceMappingURL=/clarity-seed/src/src/sign-in.component.js.map

/***/ }),

/***/ 342:
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
// import { Router } from '@angular/router';
var AppComponent = (function () {
    function AppComponent() {
    }
    AppComponent = __decorate([
        core_1.Component({
            selector: 'harbor-app',
            template: __webpack_require__(717),
            styleUrls: []
        }), 
        __metadata('design:paramtypes', [])
    ], AppComponent);
    return AppComponent;
}());
exports.AppComponent = AppComponent;
//# sourceMappingURL=/clarity-seed/src/src/app.component.js.map

/***/ }),

/***/ 343:
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
var platform_browser_1 = __webpack_require__(106);
var core_1 = __webpack_require__(0);
var forms_1 = __webpack_require__(201);
var http_1 = __webpack_require__(79);
var clarity_angular_1 = __webpack_require__(557);
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
//# sourceMappingURL=/clarity-seed/src/src/core.module.js.map

/***/ }),

/***/ 344:
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
var DashboardComponent = (function () {
    function DashboardComponent() {
    }
    DashboardComponent.prototype.ngOnInit = function () {
        this.repositories = [
            { name: 'Ubuntu', version: '14.04', count: 1 },
            { name: 'MySQL', version: 'Latest', count: 2 },
            { name: 'Photon', version: '1.0', count: 3 }
        ];
    };
    DashboardComponent = __decorate([
        core_1.Component({
            selector: 'dashboard',
            template: __webpack_require__(723)
        }), 
        __metadata('design:paramtypes', [])
    ], DashboardComponent);
    return DashboardComponent;
}());
exports.DashboardComponent = DashboardComponent;
//# sourceMappingURL=/clarity-seed/src/src/dashboard.component.js.map

/***/ }),

/***/ 345:
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
var AuditLogComponent = (function () {
    function AuditLogComponent() {
    }
    AuditLogComponent.prototype.ngOnInit = function () {
        this.auditLogs = [
            { username: 'Admin', repoName: 'project01', tag: '', operation: 'create', timestamp: '2016-12-23 12:05:17' },
            { username: 'Admin', repoName: 'project01/ubuntu', tag: '14.04', operation: 'push', timestamp: '2016-12-30 14:52:23' },
            { username: 'user1', repoName: 'project01/mysql', tag: '5.6', operation: 'pull', timestamp: '2016-12-30 12:12:33' }
        ];
    };
    AuditLogComponent = __decorate([
        core_1.Component({
            template: __webpack_require__(725)
        }), 
        __metadata('design:paramtypes', [])
    ], AuditLogComponent);
    return AuditLogComponent;
}());
exports.AuditLogComponent = AuditLogComponent;
//# sourceMappingURL=/clarity-seed/src/src/audit-log.component.js.map

/***/ }),

/***/ 346:
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
var http_1 = __webpack_require__(79);
var project_1 = __webpack_require__(350);
var project_service_1 = __webpack_require__(156);
var message_service_1 = __webpack_require__(223);
var CreateProjectComponent = (function () {
    function CreateProjectComponent(projectService, messageService) {
        this.projectService = projectService;
        this.messageService = messageService;
        this.project = new project_1.Project();
        this.create = new core_1.EventEmitter();
    }
    CreateProjectComponent.prototype.onSubmit = function () {
        var _this = this;
        this.hasError = false;
        this.projectService
            .createProject(this.project.name, this.project.public ? 1 : 0)
            .subscribe(function (status) {
            _this.create.emit(true);
            _this.createProjectOpened = false;
        }, function (error) {
            _this.hasError = true;
            if (error instanceof http_1.Response) {
                switch (error.status) {
                    case 409:
                        _this.errorMessage = 'Project name already exists.';
                        break;
                    case 400:
                        _this.errorMessage = 'Project name is illegal.';
                        break;
                    default:
                        _this.errorMessage = 'Unknown error for project name.';
                        _this.messageService.announceMessage(_this.errorMessage);
                }
            }
        });
    };
    CreateProjectComponent.prototype.newProject = function () {
        this.hasError = false;
        this.project = new project_1.Project();
        this.createProjectOpened = true;
    };
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], CreateProjectComponent.prototype, "create", void 0);
    CreateProjectComponent = __decorate([
        core_1.Component({
            selector: 'create-project',
            template: __webpack_require__(727),
            styles: [__webpack_require__(712)]
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof project_service_1.ProjectService !== 'undefined' && project_service_1.ProjectService) === 'function' && _a) || Object, (typeof (_b = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _b) || Object])
    ], CreateProjectComponent);
    return CreateProjectComponent;
    var _a, _b;
}());
exports.CreateProjectComponent = CreateProjectComponent;
//# sourceMappingURL=/clarity-seed/src/src/create-project.component.js.map

/***/ }),

/***/ 347:
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
var project_service_1 = __webpack_require__(156);
var ListProjectComponent = (function () {
    function ListProjectComponent(projectService) {
        this.projectService = projectService;
        this.selected = [];
        this.actionPerform = new core_1.EventEmitter();
    }
    ListProjectComponent.prototype.retrieve = function (name, isPublic) {
        var _this = this;
        this.projectService
            .listProjects(name, isPublic)
            .subscribe(function (response) { return _this.projects = response; }, function (error) { return _this.errorMessage = error; });
    };
    ListProjectComponent.prototype.toggleProject = function (p) {
        this.projectService
            .toggleProjectPublic(p.project_id, p.public)
            .subscribe(function (response) { return console.log(response); }, function (error) { return console.log(error); });
    };
    ListProjectComponent.prototype.deleteProject = function (p) {
        var _this = this;
        this.projectService
            .deleteProject(p.project_id)
            .subscribe(function (response) {
            console.log(response);
            _this.actionPerform.emit(true);
        }, function (error) { return console.log(error); });
    };
    ListProjectComponent.prototype.deleteSelectedProjects = function () {
        var _this = this;
        this.selected.forEach(function (p) { return _this.deleteProject(p); });
    };
    ListProjectComponent.prototype.onEdit = function (p) {
    };
    ListProjectComponent.prototype.onDelete = function (p) {
    };
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], ListProjectComponent.prototype, "actionPerform", void 0);
    ListProjectComponent = __decorate([
        core_1.Component({
            selector: 'list-project',
            template: __webpack_require__(729)
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof project_service_1.ProjectService !== 'undefined' && project_service_1.ProjectService) === 'function' && _a) || Object])
    ], ListProjectComponent);
    return ListProjectComponent;
    var _a;
}());
exports.ListProjectComponent = ListProjectComponent;
//# sourceMappingURL=/clarity-seed/src/src/list-project.component.js.map

/***/ }),

/***/ 348:
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
var MemberComponent = (function () {
    function MemberComponent() {
    }
    MemberComponent.prototype.ngOnInit = function () {
        this.members = [
            { name: 'Admin', role: 'Sys admin' },
            { name: 'user01', role: 'Project Admin' },
            { name: 'user02', role: 'Developer' },
            { name: 'user03', role: 'Guest' }
        ];
    };
    MemberComponent = __decorate([
        core_1.Component({
            template: __webpack_require__(730)
        }), 
        __metadata('design:paramtypes', [])
    ], MemberComponent);
    return MemberComponent;
}());
exports.MemberComponent = MemberComponent;
//# sourceMappingURL=/clarity-seed/src/src/member.component.js.map

/***/ }),

/***/ 349:
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
var ProjectDetailComponent = (function () {
    function ProjectDetailComponent() {
    }
    ProjectDetailComponent = __decorate([
        core_1.Component({
            selector: 'project-detail',
            template: __webpack_require__(731),
            styles: [__webpack_require__(713)]
        }), 
        __metadata('design:paramtypes', [])
    ], ProjectDetailComponent);
    return ProjectDetailComponent;
}());
exports.ProjectDetailComponent = ProjectDetailComponent;
//# sourceMappingURL=/clarity-seed/src/src/project-detail.component.js.map

/***/ }),

/***/ 350:
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
//# sourceMappingURL=/clarity-seed/src/src/project.js.map

/***/ }),

/***/ 351:
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
var ReplicationComponent = (function () {
    function ReplicationComponent() {
    }
    ReplicationComponent.prototype.ngOnInit = function () {
        this.policies = [
            { name: 'sync_01', status: 'Disabled', destination: '10.117.5.135', lastStartTime: '2016-12-21 17:52:35', description: 'test' },
            { name: 'sync_02', status: 'Enabled', destination: '10.117.5.117', lastStartTime: '2016-12-21 12:22:47', description: 'test' },
        ];
        this.jobs = [
            { name: 'project01/ubuntu:14.04', status: 'Finished', operation: 'Transfer', creationTime: '2016-12-21 17:53:50', endTime: '2016-12-21 17:55:01' },
            { name: 'project01/mysql:5.6', status: 'Finished', operation: 'Transfer', creationTime: '2016-12-21 17:54:20', endTime: '2016-12-21 17:55:05' },
            { name: 'project01/photon:latest', status: 'Finished', operation: 'Transfer', creationTime: '2016-12-21 17:54:50', endTime: '2016-12-21 17:55:15' }
        ];
    };
    ReplicationComponent = __decorate([
        core_1.Component({
            selector: 'replicaton',
            template: __webpack_require__(734)
        }), 
        __metadata('design:paramtypes', [])
    ], ReplicationComponent);
    return ReplicationComponent;
}());
exports.ReplicationComponent = ReplicationComponent;
//# sourceMappingURL=/clarity-seed/src/src/replication.component.js.map

/***/ }),

/***/ 352:
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
var RepositoryComponent = (function () {
    function RepositoryComponent() {
    }
    RepositoryComponent.prototype.ngOnInit = function () {
        this.repos = [
            { name: 'ubuntu', status: 'ready', tag: '14.04', author: 'Admin', dockerVersion: '1.10.1', created: '2016-10-10', pullCommand: 'docker pull 10.117.5.61/project01/ubuntu:14.04' },
            { name: 'mysql', status: 'ready', tag: '5.6', author: 'docker', dockerVersion: '1.11.2', created: '2016-09-23', pullCommand: 'docker pull 10.117.5.61/project01/mysql:5.6' },
            { name: 'photon', status: 'ready', tag: 'latest', author: 'Admin', dockerVersion: '1.10.1', created: '2016-11-10', pullCommand: 'docker pull 10.117.5.61/project01/photon:latest' },
        ];
    };
    RepositoryComponent = __decorate([
        core_1.Component({
            selector: 'repository',
            template: __webpack_require__(735)
        }), 
        __metadata('design:paramtypes', [])
    ], RepositoryComponent);
    return RepositoryComponent;
}());
exports.RepositoryComponent = RepositoryComponent;
//# sourceMappingURL=/clarity-seed/src/src/repository.component.js.map

/***/ }),

/***/ 353:
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
var http_1 = __webpack_require__(79);
__webpack_require__(392);
var currentUserEndpint = "/api/users/current";
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
    }
    /**
     * Get the related information of current signed in user from backend
     *
     * @returns {Promise<any>}
     *
     * @memberOf SessionService
     */
    SessionService.prototype.retrieveUser = function () {
        var _this = this;
        return this.http.get(currentUserEndpint, { headers: this.headers }).toPromise()
            .then(function (response) { return _this.currentUser = response.json(); })
            .catch(function (error) {
            console.log("An error occurred when getting current user ", error); //TODO: Will replaced with general error handler
        });
    };
    /**
     * For getting info
     */
    SessionService.prototype.getCurrentUser = function () {
        return this.currentUser;
    };
    SessionService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof http_1.Http !== 'undefined' && http_1.Http) === 'function' && _a) || Object])
    ], SessionService);
    return SessionService;
    var _a;
}());
exports.SessionService = SessionService;
//# sourceMappingURL=/clarity-seed/src/src/session.service.js.map

/***/ }),

/***/ 407:
/***/ (function(module, exports) {

function webpackEmptyContext(req) {
	throw new Error("Cannot find module '" + req + "'.");
}
webpackEmptyContext.keys = function() { return []; };
webpackEmptyContext.resolve = webpackEmptyContext;
module.exports = webpackEmptyContext;
webpackEmptyContext.id = 407;


/***/ }),

/***/ 408:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

__webpack_require__(556);
var platform_browser_dynamic_1 = __webpack_require__(500);
var core_1 = __webpack_require__(0);
var environment_1 = __webpack_require__(555);
var _1 = __webpack_require__(554);
if (environment_1.environment.production) {
    core_1.enableProdMode();
}
platform_browser_dynamic_1.platformBrowserDynamic().bootstrapModule(_1.AppModule);
//# sourceMappingURL=/clarity-seed/src/src/src/main.js.map

/***/ }),

/***/ 530:
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
var shared_module_1 = __webpack_require__(55);
var router_1 = __webpack_require__(107);
var sign_in_component_1 = __webpack_require__(341);
var AccountModule = (function () {
    function AccountModule() {
    }
    AccountModule = __decorate([
        core_1.NgModule({
            imports: [
                shared_module_1.SharedModule,
                router_1.RouterModule
            ],
            declarations: [sign_in_component_1.SignInComponent],
            exports: [sign_in_component_1.SignInComponent]
        }), 
        __metadata('design:paramtypes', [])
    ], AccountModule);
    return AccountModule;
}());
exports.AccountModule = AccountModule;
//# sourceMappingURL=/clarity-seed/src/src/account.module.js.map

/***/ }),

/***/ 531:
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
//# sourceMappingURL=/clarity-seed/src/src/sign-in-credential.js.map

/***/ }),

/***/ 532:
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
var http_1 = __webpack_require__(79);
__webpack_require__(392);
var url_prefix = '';
var signInUrl = url_prefix + '/login';
/**
 *
 * Define a service to provide sign in methods
 *
 * @export
 * @class SignInService
 */
var SignInService = (function () {
    function SignInService(http) {
        this.http = http;
        this.headers = new http_1.Headers({
            "Content-Type": 'application/x-www-form-urlencoded'
        });
    }
    //Handle the related exceptions
    SignInService.prototype.handleError = function (error) {
        return Promise.reject(error.message || error);
    };
    //Submit signin form to backend (NOT restful service)
    SignInService.prototype.signIn = function (signInCredential) {
        //Build the form package
        var body = new http_1.URLSearchParams();
        body.set('principal', signInCredential.principal);
        body.set('password', signInCredential.password);
        //Trigger Http
        return this.http.post(signInUrl, body.toString(), { headers: this.headers })
            .toPromise()
            .then(function () { return null; })
            .catch(this.handleError);
    };
    SignInService = __decorate([
        core_1.Injectable(), 
        __metadata('design:paramtypes', [(typeof (_a = typeof http_1.Http !== 'undefined' && http_1.Http) === 'function' && _a) || Object])
    ], SignInService);
    return SignInService;
    var _a;
}());
exports.SignInService = SignInService;
//# sourceMappingURL=/clarity-seed/src/src/sign-in.service.js.map

/***/ }),

/***/ 533:
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
var app_component_1 = __webpack_require__(342);
var account_module_1 = __webpack_require__(530);
var base_module_1 = __webpack_require__(536);
var harbor_routing_module_1 = __webpack_require__(542);
var core_module_1 = __webpack_require__(343);
var AppModule = (function () {
    function AppModule() {
    }
    AppModule = __decorate([
        core_1.NgModule({
            declarations: [
                app_component_1.AppComponent,
            ],
            imports: [
                core_module_1.CoreModule,
                account_module_1.AccountModule,
                base_module_1.BaseModule,
                harbor_routing_module_1.HarborRoutingModule
            ],
            providers: [],
            bootstrap: [app_component_1.AppComponent]
        }), 
        __metadata('design:paramtypes', [])
    ], AppModule);
    return AppModule;
}());
exports.AppModule = AppModule;
//# sourceMappingURL=/clarity-seed/src/src/app.module.js.map

/***/ }),

/***/ 534:
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
var router_1 = __webpack_require__(107);
var harbor_shell_component_1 = __webpack_require__(222);
var dashboard_component_1 = __webpack_require__(344);
var project_component_1 = __webpack_require__(224);
var baseRoutes = [
    {
        path: 'harbor', component: harbor_shell_component_1.HarborShellComponent,
        children: [
            { path: 'dashboard', component: dashboard_component_1.DashboardComponent },
            { path: 'projects', component: project_component_1.ProjectComponent }
        ]
    }];
var BaseRoutingModule = (function () {
    function BaseRoutingModule() {
    }
    BaseRoutingModule = __decorate([
        core_1.NgModule({
            imports: [
                router_1.RouterModule.forChild(baseRoutes)
            ],
            exports: [router_1.RouterModule]
        }), 
        __metadata('design:paramtypes', [])
    ], BaseRoutingModule);
    return BaseRoutingModule;
}());
exports.BaseRoutingModule = BaseRoutingModule;
//# sourceMappingURL=/clarity-seed/src/src/base-routing.module.js.map

/***/ }),

/***/ 535:
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
var BaseSettingsComponent = (function () {
    function BaseSettingsComponent() {
    }
    BaseSettingsComponent.prototype.ngOnInit = function () {
    };
    BaseSettingsComponent = __decorate([
        core_1.Component({
            selector: "base-settings",
            template: __webpack_require__(718)
        }), 
        __metadata('design:paramtypes', [])
    ], BaseSettingsComponent);
    return BaseSettingsComponent;
}());
exports.BaseSettingsComponent = BaseSettingsComponent;
//# sourceMappingURL=/clarity-seed/src/src/base-settings.component.js.map

/***/ }),

/***/ 536:
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
var shared_module_1 = __webpack_require__(55);
var dashboard_module_1 = __webpack_require__(540);
var project_module_1 = __webpack_require__(547);
var user_module_1 = __webpack_require__(553);
var navigator_component_1 = __webpack_require__(539);
var global_search_component_1 = __webpack_require__(538);
var footer_component_1 = __webpack_require__(537);
var harbor_shell_component_1 = __webpack_require__(222);
var base_settings_component_1 = __webpack_require__(535);
var base_routing_module_1 = __webpack_require__(534);
var BaseModule = (function () {
    function BaseModule() {
    }
    BaseModule = __decorate([
        core_1.NgModule({
            imports: [
                shared_module_1.SharedModule,
                dashboard_module_1.DashboardModule,
                project_module_1.ProjectModule,
                user_module_1.UserModule,
                base_routing_module_1.BaseRoutingModule
            ],
            declarations: [
                navigator_component_1.NavigatorComponent,
                global_search_component_1.GlobalSearchComponent,
                base_settings_component_1.BaseSettingsComponent,
                footer_component_1.FooterComponent,
                harbor_shell_component_1.HarborShellComponent
            ],
            exports: [harbor_shell_component_1.HarborShellComponent]
        }), 
        __metadata('design:paramtypes', [])
    ], BaseModule);
    return BaseModule;
}());
exports.BaseModule = BaseModule;
//# sourceMappingURL=/clarity-seed/src/src/base.module.js.map

/***/ }),

/***/ 537:
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
            template: __webpack_require__(719)
        }), 
        __metadata('design:paramtypes', [])
    ], FooterComponent);
    return FooterComponent;
}());
exports.FooterComponent = FooterComponent;
//# sourceMappingURL=/clarity-seed/src/src/footer.component.js.map

/***/ }),

/***/ 538:
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
var GlobalSearchComponent = (function () {
    function GlobalSearchComponent() {
    }
    GlobalSearchComponent = __decorate([
        core_1.Component({
            selector: 'global-search',
            template: __webpack_require__(720)
        }), 
        __metadata('design:paramtypes', [])
    ], GlobalSearchComponent);
    return GlobalSearchComponent;
}());
exports.GlobalSearchComponent = GlobalSearchComponent;
//# sourceMappingURL=/clarity-seed/src/src/global-search.component.js.map

/***/ }),

/***/ 539:
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
var NavigatorComponent = (function () {
    function NavigatorComponent() {
    }
    NavigatorComponent = __decorate([
        core_1.Component({
            selector: 'navigator',
            template: __webpack_require__(722)
        }), 
        __metadata('design:paramtypes', [])
    ], NavigatorComponent);
    return NavigatorComponent;
}());
exports.NavigatorComponent = NavigatorComponent;
//# sourceMappingURL=/clarity-seed/src/src/navigator.component.js.map

/***/ }),

/***/ 540:
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
var dashboard_component_1 = __webpack_require__(344);
var shared_module_1 = __webpack_require__(55);
var DashboardModule = (function () {
    function DashboardModule() {
    }
    DashboardModule = __decorate([
        core_1.NgModule({
            imports: [shared_module_1.SharedModule],
            declarations: [dashboard_component_1.DashboardComponent],
            exports: [dashboard_component_1.DashboardComponent]
        }), 
        __metadata('design:paramtypes', [])
    ], DashboardModule);
    return DashboardModule;
}());
exports.DashboardModule = DashboardModule;
//# sourceMappingURL=/clarity-seed/src/src/dashboard.module.js.map

/***/ }),

/***/ 541:
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
var message_service_1 = __webpack_require__(223);
var MessageComponent = (function () {
    function MessageComponent(messageService) {
        var _this = this;
        messageService.messageAnnounced$.subscribe(function (message) {
            _this.globalMessageOpened = true;
            _this.globalMessage = message;
            console.log('received message:' + message);
        });
    }
    MessageComponent.prototype.onClose = function () {
        this.globalMessageOpened = false;
    };
    MessageComponent = __decorate([
        core_1.Component({
            selector: 'global-message',
            template: __webpack_require__(724)
        }), 
        __metadata('design:paramtypes', [(typeof (_a = typeof message_service_1.MessageService !== 'undefined' && message_service_1.MessageService) === 'function' && _a) || Object])
    ], MessageComponent);
    return MessageComponent;
    var _a;
}());
exports.MessageComponent = MessageComponent;
//# sourceMappingURL=/clarity-seed/src/src/message.component.js.map

/***/ }),

/***/ 542:
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
var router_1 = __webpack_require__(107);
var sign_in_component_1 = __webpack_require__(341);
var harborRoutes = [
    { path: '', redirectTo: '/sign-in', pathMatch: 'full' },
    { path: 'sign-in', component: sign_in_component_1.SignInComponent }
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
//# sourceMappingURL=/clarity-seed/src/src/harbor-routing.module.js.map

/***/ }),

/***/ 543:
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
var audit_log_component_1 = __webpack_require__(345);
var shared_module_1 = __webpack_require__(55);
var LogModule = (function () {
    function LogModule() {
    }
    LogModule = __decorate([
        core_1.NgModule({
            imports: [shared_module_1.SharedModule],
            declarations: [audit_log_component_1.AuditLogComponent],
            exports: [audit_log_component_1.AuditLogComponent]
        }), 
        __metadata('design:paramtypes', [])
    ], LogModule);
    return LogModule;
}());
exports.LogModule = LogModule;
//# sourceMappingURL=/clarity-seed/src/src/log.module.js.map

/***/ }),

/***/ 544:
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
var project_1 = __webpack_require__(350);
var project_service_1 = __webpack_require__(156);
var ActionProjectComponent = (function () {
    function ActionProjectComponent(projectService) {
        this.projectService = projectService;
        this.togglePublic = new core_1.EventEmitter();
        this.deleteProject = new core_1.EventEmitter();
    }
    ActionProjectComponent.prototype.toggle = function () {
        if (this.project) {
            this.project.public === 0 ? this.project.public = 1 : this.project.public = 0;
            this.togglePublic.emit(this.project);
        }
    };
    ActionProjectComponent.prototype.delete = function () {
        if (this.project) {
            this.deleteProject.emit(this.project);
        }
    };
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], ActionProjectComponent.prototype, "togglePublic", void 0);
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], ActionProjectComponent.prototype, "deleteProject", void 0);
    __decorate([
        core_1.Input(), 
        __metadata('design:type', (typeof (_a = typeof project_1.Project !== 'undefined' && project_1.Project) === 'function' && _a) || Object)
    ], ActionProjectComponent.prototype, "project", void 0);
    ActionProjectComponent = __decorate([
        core_1.Component({
            selector: 'action-project',
            template: __webpack_require__(726)
        }), 
        __metadata('design:paramtypes', [(typeof (_b = typeof project_service_1.ProjectService !== 'undefined' && project_service_1.ProjectService) === 'function' && _b) || Object])
    ], ActionProjectComponent);
    return ActionProjectComponent;
    var _a, _b;
}());
exports.ActionProjectComponent = ActionProjectComponent;
//# sourceMappingURL=/clarity-seed/src/src/action-project.component.js.map

/***/ }),

/***/ 545:
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
exports.projectTypes = [
    { 'key': 0, 'value': 'My Projects' },
    { 'key': 1, 'value': 'Public Projects' }
];
var FilterProjectComponent = (function () {
    function FilterProjectComponent() {
        this.filter = new core_1.EventEmitter();
        this.types = exports.projectTypes;
        this.currentType = exports.projectTypes[0];
    }
    FilterProjectComponent.prototype.doFilter = function (type) {
        console.log('Filtered projects by:' + type);
        this.currentType = exports.projectTypes.find(function (item) { return item.key === type; });
        this.filter.emit(type);
    };
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], FilterProjectComponent.prototype, "filter", void 0);
    FilterProjectComponent = __decorate([
        core_1.Component({
            selector: 'filter-project',
            template: __webpack_require__(728)
        }), 
        __metadata('design:paramtypes', [])
    ], FilterProjectComponent);
    return FilterProjectComponent;
}());
exports.FilterProjectComponent = FilterProjectComponent;
//# sourceMappingURL=/clarity-seed/src/src/filter-project.component.js.map

/***/ }),

/***/ 546:
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
var router_1 = __webpack_require__(107);
var harbor_shell_component_1 = __webpack_require__(222);
var project_component_1 = __webpack_require__(224);
var project_detail_component_1 = __webpack_require__(349);
var repository_component_1 = __webpack_require__(352);
var replication_component_1 = __webpack_require__(351);
var member_component_1 = __webpack_require__(348);
var audit_log_component_1 = __webpack_require__(345);
var projectRoutes = [
    { path: 'harbor',
        component: harbor_shell_component_1.HarborShellComponent,
        children: [
            { path: 'projects', component: project_component_1.ProjectComponent },
            {
                path: 'projects/:id',
                component: project_detail_component_1.ProjectDetailComponent,
                children: [
                    { path: 'repository', component: repository_component_1.RepositoryComponent },
                    { path: 'replication', component: replication_component_1.ReplicationComponent },
                    { path: 'member', component: member_component_1.MemberComponent },
                    { path: 'log', component: audit_log_component_1.AuditLogComponent }
                ]
            }
        ]
    }
];
var ProjectRoutingModule = (function () {
    function ProjectRoutingModule() {
    }
    ProjectRoutingModule = __decorate([
        core_1.NgModule({
            imports: [
                router_1.RouterModule.forChild(projectRoutes)
            ],
            exports: [router_1.RouterModule]
        }), 
        __metadata('design:paramtypes', [])
    ], ProjectRoutingModule);
    return ProjectRoutingModule;
}());
exports.ProjectRoutingModule = ProjectRoutingModule;
//# sourceMappingURL=/clarity-seed/src/src/project-routing.module.js.map

/***/ }),

/***/ 547:
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
var shared_module_1 = __webpack_require__(55);
var repository_module_1 = __webpack_require__(550);
var replication_module_1 = __webpack_require__(549);
var log_module_1 = __webpack_require__(543);
var project_component_1 = __webpack_require__(224);
var create_project_component_1 = __webpack_require__(346);
var search_project_component_1 = __webpack_require__(548);
var filter_project_component_1 = __webpack_require__(545);
var action_project_component_1 = __webpack_require__(544);
var list_project_component_1 = __webpack_require__(347);
var project_detail_component_1 = __webpack_require__(349);
var member_component_1 = __webpack_require__(348);
var project_routing_module_1 = __webpack_require__(546);
var project_service_1 = __webpack_require__(156);
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
                project_routing_module_1.ProjectRoutingModule
            ],
            declarations: [
                project_component_1.ProjectComponent,
                create_project_component_1.CreateProjectComponent,
                search_project_component_1.SearchProjectComponent,
                filter_project_component_1.FilterProjectComponent,
                action_project_component_1.ActionProjectComponent,
                list_project_component_1.ListProjectComponent,
                project_detail_component_1.ProjectDetailComponent,
                member_component_1.MemberComponent
            ],
            exports: [list_project_component_1.ListProjectComponent],
            providers: [project_service_1.ProjectService]
        }), 
        __metadata('design:paramtypes', [])
    ], ProjectModule);
    return ProjectModule;
}());
exports.ProjectModule = ProjectModule;
//# sourceMappingURL=/clarity-seed/src/src/project.module.js.map

/***/ }),

/***/ 548:
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
var SearchProjectComponent = (function () {
    function SearchProjectComponent() {
        this.search = new core_1.EventEmitter();
    }
    SearchProjectComponent.prototype.doSearch = function (projectName) {
        this.search.emit(projectName);
    };
    __decorate([
        core_1.Output(), 
        __metadata('design:type', Object)
    ], SearchProjectComponent.prototype, "search", void 0);
    SearchProjectComponent = __decorate([
        core_1.Component({
            selector: 'search-project',
            template: __webpack_require__(733)
        }), 
        __metadata('design:paramtypes', [])
    ], SearchProjectComponent);
    return SearchProjectComponent;
}());
exports.SearchProjectComponent = SearchProjectComponent;
//# sourceMappingURL=/clarity-seed/src/src/search-project.component.js.map

/***/ }),

/***/ 549:
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
var replication_component_1 = __webpack_require__(351);
var shared_module_1 = __webpack_require__(55);
var ReplicationModule = (function () {
    function ReplicationModule() {
    }
    ReplicationModule = __decorate([
        core_1.NgModule({
            imports: [shared_module_1.SharedModule],
            declarations: [replication_component_1.ReplicationComponent],
            exports: [replication_component_1.ReplicationComponent]
        }), 
        __metadata('design:paramtypes', [])
    ], ReplicationModule);
    return ReplicationModule;
}());
exports.ReplicationModule = ReplicationModule;
//# sourceMappingURL=/clarity-seed/src/src/replication.module.js.map

/***/ }),

/***/ 55:
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
var core_module_1 = __webpack_require__(343);
var session_service_1 = __webpack_require__(353);
var message_component_1 = __webpack_require__(541);
var message_service_1 = __webpack_require__(223);
var SharedModule = (function () {
    function SharedModule() {
    }
    SharedModule = __decorate([
        core_1.NgModule({
            imports: [
                core_module_1.CoreModule
            ],
            declarations: [
                message_component_1.MessageComponent
            ],
            exports: [
                core_module_1.CoreModule,
                message_component_1.MessageComponent
            ],
            providers: [session_service_1.SessionService, message_service_1.MessageService]
        }), 
        __metadata('design:paramtypes', [])
    ], SharedModule);
    return SharedModule;
}());
exports.SharedModule = SharedModule;
//# sourceMappingURL=/clarity-seed/src/src/shared.module.js.map

/***/ }),

/***/ 550:
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
var repository_component_1 = __webpack_require__(352);
var shared_module_1 = __webpack_require__(55);
var RepositoryModule = (function () {
    function RepositoryModule() {
    }
    RepositoryModule = __decorate([
        core_1.NgModule({
            imports: [shared_module_1.SharedModule],
            declarations: [repository_component_1.RepositoryComponent],
            exports: [repository_component_1.RepositoryComponent]
        }), 
        __metadata('design:paramtypes', [])
    ], RepositoryModule);
    return RepositoryModule;
}());
exports.RepositoryModule = RepositoryModule;
//# sourceMappingURL=/clarity-seed/src/src/repository.module.js.map

/***/ }),

/***/ 551:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var http_1 = __webpack_require__(79);
var Observable_1 = __webpack_require__(8);
var BaseService = (function () {
    function BaseService() {
    }
    BaseService.prototype.handleError = function (error) {
        // In a real world app, we might use a remote logging infrastructure
        var errMsg;
        if (error instanceof http_1.Response) {
            var body = error.json() || '';
            var err = body.error || JSON.stringify(body);
            errMsg = error.status + " - " + (error.statusText || '') + " " + err;
        }
        else {
            errMsg = error.message ? error.message : error.toString();
        }
        console.error(errMsg);
        return Observable_1.Observable.throw(errMsg);
    };
    return BaseService;
}());
exports.BaseService = BaseService;
//# sourceMappingURL=/clarity-seed/src/src/base.service.js.map

/***/ }),

/***/ 552:
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
var UserComponent = (function () {
    function UserComponent() {
    }
    UserComponent = __decorate([
        core_1.Component({
            selector: 'harbor-user',
            template: __webpack_require__(736)
        }), 
        __metadata('design:paramtypes', [])
    ], UserComponent);
    return UserComponent;
}());
exports.UserComponent = UserComponent;
//# sourceMappingURL=/clarity-seed/src/src/user.component.js.map

/***/ }),

/***/ 553:
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
var shared_module_1 = __webpack_require__(55);
var user_component_1 = __webpack_require__(552);
var UserModule = (function () {
    function UserModule() {
    }
    UserModule = __decorate([
        core_1.NgModule({
            imports: [
                shared_module_1.SharedModule
            ],
            declarations: [
                user_component_1.UserComponent
            ],
            exports: [
                user_component_1.UserComponent
            ]
        }), 
        __metadata('design:paramtypes', [])
    ], UserModule);
    return UserModule;
}());
exports.UserModule = UserModule;
//# sourceMappingURL=/clarity-seed/src/src/user.module.js.map

/***/ }),

/***/ 554:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

function __export(m) {
    for (var p in m) if (!exports.hasOwnProperty(p)) exports[p] = m[p];
}
__export(__webpack_require__(342));
__export(__webpack_require__(533));
//# sourceMappingURL=/clarity-seed/src/src/src/app/index.js.map

/***/ }),

/***/ 555:
/***/ (function(module, exports, __webpack_require__) {

"use strict";
// The file contents for the current environment will overwrite these during build.
// The build system defaults to the dev environment which uses `environment.ts`, but if you do
// `ng build --env=prod` then `environment.prod.ts` will be used instead.
// The list of which env maps to which file can be found in `angular-cli.json`.

exports.environment = {
    production: false
};
//# sourceMappingURL=/clarity-seed/src/src/src/environments/environment.js.map

/***/ }),

/***/ 556:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

// This file includes polyfills needed by Angular 2 and is loaded before
// the app. You can add your own extra polyfills to this file.
__webpack_require__(571);
__webpack_require__(564);
__webpack_require__(560);
__webpack_require__(566);
__webpack_require__(565);
__webpack_require__(563);
__webpack_require__(562);
__webpack_require__(570);
__webpack_require__(559);
__webpack_require__(558);
__webpack_require__(568);
__webpack_require__(561);
__webpack_require__(569);
__webpack_require__(567);
__webpack_require__(572);
__webpack_require__(761);
//# sourceMappingURL=/clarity-seed/src/src/src/polyfills.js.map

/***/ }),

/***/ 711:
/***/ (function(module, exports, __webpack_require__) {

exports = module.exports = __webpack_require__(58)();
// imports


// module
exports.push([module.i, ".progress-size-small {\n    height: 0.5em !important;\n}\n\n.visibility-hidden {\n    visibility: hidden;\n}", ""]);

// exports


/*** EXPORTS FROM exports-loader ***/
module.exports = module.exports.toString();

/***/ }),

/***/ 712:
/***/ (function(module, exports, __webpack_require__) {

exports = module.exports = __webpack_require__(58)();
// imports


// module
exports.push([module.i, "", ""]);

// exports


/*** EXPORTS FROM exports-loader ***/
module.exports = module.exports.toString();

/***/ }),

/***/ 713:
/***/ (function(module, exports, __webpack_require__) {

exports = module.exports = __webpack_require__(58)();
// imports


// module
exports.push([module.i, ".display-in-line {\n    display: inline-block;\n}\n\n.project-title {\n    margin-left: 10px; \n}\n\n.pull-right {\n    float: right !important;\n}", ""]);

// exports


/*** EXPORTS FROM exports-loader ***/
module.exports = module.exports.toString();

/***/ }),

/***/ 714:
/***/ (function(module, exports, __webpack_require__) {

exports = module.exports = __webpack_require__(58)();
// imports


// module
exports.push([module.i, ".my-project-pull-right {\n    float: right;\n}", ""]);

// exports


/*** EXPORTS FROM exports-loader ***/
module.exports = module.exports.toString();

/***/ }),

/***/ 716:
/***/ (function(module, exports) {

module.exports = "<div class=\"login-wrapper\">\n    <form #signInForm=\"ngForm\" class=\"login\">\n        <label class=\"title\">\n        VMware Harbor<span class=\"trademark\">&#8482;</span>\n    </label>\n        <div class=\"login-group\">\n            <label for=\"username\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-left\" [class.invalid]=\"userNameInput.invalid && (userNameInput.dirty || userNameInput.touched)\">\n                <input class=\"username\" type=\"text\" required\n                [(ngModel)]=\"signInCredential.principal\" \n                name=\"login_username\" id=\"login_username\" placeholder=\"Username\"\n                #userNameInput='ngModel'>\n                <span class=\"tooltip-content\">\n                    Username is required!\n                </span>\n            </label>\n            <label for=\"username\" aria-haspopup=\"true\" role=\"tooltip\" class=\"tooltip tooltip-validation tooltip-md tooltip-top-left\" [class.invalid]=\"passwordInput.invalid && (passwordInput.dirty || passwordInput.touched)\">\n                <input class=\"password\" type=\"password\" required \n                [(ngModel)]=\"signInCredential.password\" \n                name=\"login_password\" id=\"login_password\" placeholder=\"Password\" \n                #passwordInput=\"ngModel\">\n                <span class=\"tooltip-content\">\n                    Password is required!\n                </span>\n            </label>\n            <div class=\"checkbox\">\n                <input type=\"checkbox\" id=\"rememberme\">\n                <label for=\"rememberme\">\n            Remember me\n        </label>\n            </div>\n            <div [class.visibility-hidden]=\"signInStatus != statusError\" class=\"error active\">\n                Invalid user name or password\n            </div>\n            <button [class.visibility-hidden]=\"signInStatus === statusOnGoing\" [disabled]=\"signInStatus === statusOnGoing\" type=\"submit\" class=\"btn btn-primary\" (click)=\"signIn()\">LOG IN</button>\n            <div [class.visibility-hidden]=\"signInStatus != statusOnGoing\" class=\"progress loop progress-size-small\"><progress></progress></div>\n            <a href=\"javascript:void(0)\" class=\"signup\" (click)=\"signUp()\">Sign up for an account</a>\n        </div>\n    </form>\n</div>"

/***/ }),

/***/ 717:
/***/ (function(module, exports) {

module.exports = "<router-outlet></router-outlet>"

/***/ }),

/***/ 718:
/***/ (function(module, exports) {

module.exports = "<clr-dropdown [clrMenuPosition]=\"'bottom-right'\">\n    <button class=\"nav-text\" clrDropdownToggle>\n          <!--<clr-icon shape=\"user\" class=\"is-inverse\" size=\"24\"></clr-icon>-->\n          <span>Administrator</span>\n          <clr-icon shape=\"caret down\"></clr-icon>\n      </button>\n    <div class=\"dropdown-menu\">\n        <a href=\"javascript:void(0)\" clrDropdownItem>Add User</a>\n        <a href=\"javascript:void(0)\" clrDropdownItem>Account Setting</a>\n        <a href=\"javascript:void(0)\" clrDropdownItem>About</a>\n    </div>\n</clr-dropdown>"

/***/ }),

/***/ 719:
/***/ (function(module, exports) {

module.exports = ""

/***/ }),

/***/ 720:
/***/ (function(module, exports) {

module.exports = "<a class=\"nav-link\">\n  <span class=\"nav-text\">\n    <clr-icon shape=\"search\"></clr-icon>\n  </span>\n</a>"

/***/ }),

/***/ 721:
/***/ (function(module, exports) {

module.exports = "<clr-main-container>\n  <clr-modal [(clrModalOpen)]=\"account_settings_opened\">\n    <h3 class=\"modal-title\">Accout Settings</h3>\n    <div class=\"modal-body\">\n      <form>\n        <section class=\"form-block\">\n          <div class=\"form-group\">\n            <label for=\"account_settings_username\" class=\"col-md-4\">Username</label>\n            <input type=\"text\" class=\"col-md-8\" id=\"account_settings_username\" size=\"20\"> \n          </div>\n          <div class=\"form-group\">\n            <label for=\"account_settings_email\" class=\"col-md-4\">Email</label>\n            <input type=\"text\" class=\"col-md-8\" id=\"account_settings_email\" size=\"20\"> \n          </div>\n          <div class=\"form-group\">\n            <label for=\"account_settings_full_name\" class=\"col-md-4\">Full name</label>\n            <input type=\"text\" class=\"col-md-8\" id=\"account_settings_full_name\" size=\"20\"> \n          </div>\n          <div class=\"form-group\">\n            <label for=\"account_settings_comments\" class=\"col-md-4\">Comments</label>\n            <input type=\"text\" class=\"col-md-8\" id=\"account_settings_comments\" size=\"20\"> \n          </div>\n          <div class=\"form-group\">\n            <button type=\"button\" class=\"col-md-4\" class=\"btn btn-outline\">Change Password</button>\n          </div>\n        </section>\n      </form>\n    </div>\n    <div class=\"modal-footer\">\n      <button type=\"button\" class=\"btn btn-outline\" (click)=\"account_settings_opened = false\">Cancel</button>\n      <button type=\"button\" class=\"btn btn-primary\" (click)=\"account_settings_opened = false\">Ok</button>\n    </div>\n  </clr-modal>\n  <navigator></navigator>\n  <global-message></global-message>\n  <div class=\"content-container\">\n    <div class=\"content-area\">  \n      <router-outlet></router-outlet>\n    </div>\n  </div>\n</clr-main-container>\n"

/***/ }),

/***/ 722:
/***/ (function(module, exports) {

module.exports = "<clr-header class=\"header-4 header\">\n    <div class=\"branding\">\n        <a href=\"#\" class=\"nav-link\">\n            <clr-icon shape=\"vm-bug\"></clr-icon>\n            <span class=\"title\">Harbor</span>\n        </a>\n    </div>\n    <div class=\"divider\"></div>\n    <div class=\"header-nav\">\n        <a class=\"nav-link\" id=\"dashboard-link\" [routerLink]=\"['/harbor', 'dashboard']\" routerLinkActive=\"active\">\n            <span class=\"nav-text\">Dashboard</span>\n        </a>\n        <a class=\"nav-link\" id=\"dashboard-link\" [routerLink]=\"['/harbor', 'projects']\" routerLinkActive=\"active\">\n            <span class=\"nav-text\">Project</span>\n        </a>\n    </div>\n    <div class=\"header-actions\">\n        <clr-dropdown [clrMenuPosition]=\"'bottom-right'\">\n            <!--<clr-icon shape=\"user\" class=\"is-inverse\" size=\"24\"></clr-icon>-->\n            <button class=\"nav-text\" clrDropdownToggle>\n              <span>Administrator</span>\n              <clr-icon shape=\"caret down\"></clr-icon>\n            </button>\n            <div class=\"dropdown-menu\">\n                <a href=\"javascript:void(0)\" clrDropdownItem>Add User</a>\n                <a href=\"javascript:void(0)\" clrDropdownItem>Account Setting</a>\n                <a href=\"javascript:void(0)\" clrDropdownItem>About</a>\n            </div>\n        </clr-dropdown>\n        <global-search></global-search>\n        <clr-dropdown class=\"dropdown bottom-right\">\n            <button class=\"nav-text\" clrDropdownToggle>\n                <clr-icon shape=\"cog\"></clr-icon>\n            </button>\n            <div class=\"dropdown-menu\">\n                <a href=\"javascript:void(0)\" clrDropdownItem>Preferences</a>\n                <a href=\"javascript:void(0)\" clrDropdownItem>Log out</a>\n            </div>\n        </clr-dropdown>\n    </div>\n</clr-header>"

/***/ }),

/***/ 723:
/***/ (function(module, exports) {

module.exports = "<h3>Dashboard</h3>\n<div class=\"row\">\n    <div class=\"col-lg-4 col-md-12 col-sm-12 col-xs-12\">\n        <div class=\"card\">\n            <div class=\"card-block\">\n                <h1 class=\"card-title\">Why user Harbor?</h1>\n                <p class=\"card-text\">\n                    Project Harbor is an enterprise-class registry server, which extends the open source Docker Registry server by adding the functionality usually required by an enterprise, such as security, control, and management. Harbor is primarily designed to be a private registry - providing the needed security and control that enterprises require. It also helps minimize ...\n                </p>\n            </div>\n            <div class=\"card-footer\">\n                <a href=\"...\" class=\"btn btn-sm btn-link\">View all</a>\n            </div>\n        </div>\n    </div>\n    <div class=\"col-lg-4 col-md-12 col-sm-12 col-xs-12\">\n        <div class=\"card\">\n            <div class=\"card-block\">\n                <h1 class=\"card-title\">Getting started</h1>\n                  <ul class=\"list\" style=\"list-style-type: none;\">\n                     <li><img src=\"../../images/Step1.png\" style=\"width: 19%; height: auto;\"/><a style=\"margin: 30px;\" href=\"\">Anonymous repository access</a></li>\n                     <li><img src=\"../../images/Step2.png\" style=\"width: 19%; height: auto;\"/><a style=\"margin: 30px;\" href=\"\">Repositories managed by project</a></li>\n                     <li><img src=\"../../images/Step3.png\" style=\"width: 19%; height: auto;\"/><a style=\"margin: 30px;\" href=\"\">Role based access control</a></li>\n                  </ul>\n                \n            </div>\n        </div>\n    </div>\n    <div class=\"col-lg-4 col-md-12 col-sm-12 col-xs-12\">\n        <div class=\"card\">\n            <div class=\"card-block\">\n                <h1 class=\"card-title\">Activities</h1>\n                <p class=\"card-text\">\n                    ...\n                </p>\n            </div>\n        </div>\n    </div>\n</div>\n<div class=\"row\">\n  <div class=\"col-lg-8 col-md-8 col-sm-12 col-xs-12\">\n    <clr-datagrid>\n        <clr-dg-column>Name</clr-dg-column>\n        <clr-dg-column>Version</clr-dg-column>\n        <clr-dg-column>Count</clr-dg-column>\n        <clr-dg-row *ngFor=\"let r of repositories\">\n        <clr-dg-cell>{{r.name}}</clr-dg-cell>\n        <clr-dg-cell>{{r.version}}</clr-dg-cell>\n        <clr-dg-cell>{{r.count}}</clr-dg-cell>\n        </clr-dg-row>\n        <clr-dg-footer>{{repositories.length}} item(s)</clr-dg-footer>\n    </clr-datagrid>\n  </div>\n</div>"

/***/ }),

/***/ 724:
/***/ (function(module, exports) {

module.exports = "<clr-alert [clrAlertType]=\"'alert-danger'\" [clrAlertAppLevel]=\"true\" [(clrAlertClosed)]=\"!globalMessageOpened\" (clrAlertClosedChange)=\"onClose()\">\n  <div class=\"alert-item\">\n    <span class=\"alert-text\">\n      {{globalMessage}}\n    </span>\n  </div>\n</clr-alert>"

/***/ }),

/***/ 725:
/***/ (function(module, exports) {

module.exports = "<div class=\"row\">\n  <div class=\"col-lg-12 col-md-12 col-sm-12 col-xs-12\">  \n    <div class=\"row flex-items-xs-between\">\n      <div class=\"col-md-2 push-md-8\">\n        <input type=\"text\" placeholder=\"Search for user\">\n      </div>\n    </div>\n    <clr-datagrid>\n      <clr-dg-column>Username</clr-dg-column>\n      <clr-dg-column>Repository Name</clr-dg-column>\n      <clr-dg-column>Tag</clr-dg-column>\n      <clr-dg-column>Operation</clr-dg-column>\n      <clr-dg-column>Timestamp</clr-dg-column>\n      <clr-dg-row *ngFor=\"let l of auditLogs\">\n        <clr-dg-cell>{{l.username}}</clr-dg-cell>\n        <clr-dg-cell>{{l.repoName}}</clr-dg-cell>\n        <clr-dg-cell>{{l.tag}}</clr-dg-cell>\n        <clr-dg-cell>{{l.operation}}</clr-dg-cell>\n        <clr-dg-cell>{{l.timestamp}}</clr-dg-cell>\n      </clr-dg-row>\n      <clr-dg-footer>{{auditLogs.length}} item(s)</clr-dg-footer>\n    </clr-datagrid>\n  </div>\n</div>"

/***/ }),

/***/ 726:
/***/ (function(module, exports) {

module.exports = "<clr-dropdown [clrMenuPosition]=\"'bottom-right'\" [clrCloseMenuOnItemClick]=\"true\">\n  <button clrDropdownToggle>\n    <clr-icon shape=\"ellipses-vertical\"></clr-icon>\n  </button>\n  <div class=\"dropdown-menu\">\n    <a href=\"javascript:void(0)\" clrDropdownItem>New Policy</a>\n    <a href=\"javascript:void(0)\" clrDropdownItem (click)=\"toggle()\">Make {{project.public === 0 ? 'Public' : 'Private'}} </a>\n    <div class=\"dropdown-divider\"></div>\n    <a href=\"javascript:void(0)\" clrDropdownItem (click)=\"delete()\">Delete</a>\n  </div>\n</clr-dropdown>"

/***/ }),

/***/ 727:
/***/ (function(module, exports) {

module.exports = "<clr-modal [(clrModalOpen)]=\"createProjectOpened\">\n  <h3 class=\"modal-title\">New Project</h3>\n  <div class=\"modal-body\">\n    <form>\n      <section class=\"form-block\">\n        <div class=\"form-group\">\n          <label for=\"create_project_name\" class=\"col-md-4\">Project Name</label>\n          <label for=\"create_project_name\" aria-haspopup=\"true\" role=\"tooltip\" [class.invalid]=\"hasError\" [class.valid]=\"!hasError\" class=\"tooltip tooltip-validation tooltip-sm tooltip-bottom-right\">\n            <input type=\"text\" id=\"create_project_name\"  [(ngModel)]=\"project.name\" name=\"name\" size=\"20\" (keyup)=\"hasError=false;\">\n            <span class=\"tooltip-content\">\n              {{errorMessage}}\n            </span>\n          </label>\n        </div>\n        <div class=\"form-group\">\n          <label class=\"col-md-4\">Public</label>\n          <div class=\"checkbox-inline\">\n            <input type=\"checkbox\" id=\"create_project_public\" [(ngModel)]=\"project.public\" name=\"public\">\n            <label for=\"create_project_public\"></label>\n          </div>\n        </div>\n      </section>\n    </form>\n  </div>\n  <div class=\"modal-footer\">\n    <button type=\"button\" class=\"btn btn-outline\" (click)=\"createProjectOpened = false\">Cancel</button>\n    <button type=\"button\" class=\"btn btn-primary\" (click)=\"onSubmit()\">Ok</button>\n  </div>\n</clr-modal>\n"

/***/ }),

/***/ 728:
/***/ (function(module, exports) {

module.exports = "<clr-dropdown [clrMenuPosition]=\"'bottom-left'\">\n  <button class=\"btn btn-sm btn-link\" clrDropdownToggle>\n    {{currentType.value}}\n    <clr-icon shape=\"caret down\"></clr-icon>\n  </button>\n  <div class=\"dropdown-menu\">\n    <a href=\"javascript:void(0)\" clrDropdownItem *ngFor=\"let p of types\" (click)=\"doFilter(p.key)\">{{p.value}}</a>\n  </div>\n</clr-dropdown>"

/***/ }),

/***/ 729:
/***/ (function(module, exports) {

module.exports = "<clr-datagrid [(clrDgSelected)]=\"selected\">\n  <clr-dg-column>Name</clr-dg-column>\n  <clr-dg-column>Public/Private</clr-dg-column>\n  <clr-dg-column>Repositories</clr-dg-column>\n  <clr-dg-column>Creation time</clr-dg-column>\n  <clr-dg-column>Description</clr-dg-column> \n  <clr-dg-row *clrDgItems=\"let p of projects\" [clrDgItem]=\"p\" [(clrDgSelected)]=\"p.selected\">\n    <!--<clr-dg-action-overflow>\n      <button class=\"action-item\" (click)=\"onEdit(p)\">Edit</button>\n      <button class=\"action-item\" (click)=\"onDelete(p)\">Delete</button>\n    </clr-dg-action-overflow>-->\n    <clr-dg-cell><a [routerLink]=\"['/harbor', 'projects', p.id, 'repository']\" >{{p.name}}</a></clr-dg-cell>\n    <clr-dg-cell>{{p.public == 1 ? 'Public': 'Private'}}</clr-dg-cell>\n    <clr-dg-cell>{{p.repo_count}}</clr-dg-cell>\n    <clr-dg-cell>{{p.creation_time}}</clr-dg-cell>\n    <clr-dg-cell>\n      {{p.description}}\n      <span style=\"float: right;\">\n        <action-project (togglePublic)=\"toggleProject($event)\" (deleteProject)=\"deleteProject($event)\" [project]=\"p\"></action-project>\n      </span>\n    </clr-dg-cell>\n  \n    \n  \n  </clr-dg-row>\n  <clr-dg-footer>{{ (projects ? projects.length : 0) }} item(s)</clr-dg-footer>\n</clr-datagrid>"

/***/ }),

/***/ 730:
/***/ (function(module, exports) {

module.exports = "<div class=\"row\">\n  <div class=\"col-lg-12 col-md-12 col-sm-12 col-xs-12\">  \n    <div class=\"row flex-items-xs-between\">\n      <div class=\"col-xs-4\">\n        <button class=\"btn btn-sm\">new user</button>\n      </div>\n      <div class=\"col-xs-4\">\n        <input type=\"text\" placeholder=\"Search for users\">\n      </div>\n    </div>\n    <clr-datagrid>\n      <clr-dg-column>Name</clr-dg-column>\n      <clr-dg-column>Role</clr-dg-column>\n      <clr-dg-column>Action</clr-dg-column>\n      <clr-dg-row *ngFor=\"let u of members\">\n        <clr-dg-cell>{{u.name}}</clr-dg-cell>\n        <clr-dg-cell>{{u.role}}</clr-dg-cell>\n        <clr-dg-cell>\n          <clr-dropdown [clrMenuPosition]=\"'bottom-left'\">\n            <button class=\"btn btn-sm btn-link\" clrDropdownToggle>\n              Actions\n              <clr-icon shape=\"caret down\"></clr-icon>\n            </button>\n            <div class=\"dropdown-menu\">\n              <a href=\"javascript:void(0)\" clrDropdownItem>Project Admin</a>\n              <a href=\"javascript:void(0)\" clrDropdownItem>Developer</a>\n              <a href=\"javascript:void(0)\" clrDropdownItem>Guest</a>\n            </div>\n          </clr-dropdown>\n        </clr-dg-cell>\n      </clr-dg-row>\n      <clr-dg-footer>{{members.length}} item(s)</clr-dg-footer>\n    </clr-datagrid>\n  </div>\n</div>"

/***/ }),

/***/ 731:
/***/ (function(module, exports) {

module.exports = "<h1 class=\"display-in-line\">Project 01</h1><h6 class=\"display-in-line project-title\">PROJECT</h6>\n<nav class=\"subnav\">\n  <ul class=\"nav\">\n    <li class=\"nav-item\">\n      <a class=\"nav-link\" routerLink=\"repository\" routerLinkActive=\"active\">Repositories</a>\n    </li>\n    <li class=\"nav-item\">\n      <a class=\"nav-link\" routerLink=\"replication\" routerLinkActive=\"active\">Replication</a>\n    </li>\n    <li class=\"nav-item\">\n      <a class=\"nav-link\" routerLink=\"member\" routerLinkActive=\"active\">Users</a>\n    </li>\n    <li class=\"nav-item\">\n      <a class=\"nav-link\" routerLink=\"log\" routerLinkActive=\"active\">Logs</a>\n    </li>\n  </ul>\n</nav>\n<router-outlet></router-outlet>\n"

/***/ }),

/***/ 732:
/***/ (function(module, exports) {

module.exports = "<h3>Projects</h3>\n<div class=\"row flex-items-xs-between\">\n  <div class=\"col-xs-4\">\n    <button class=\"btn btn-link\" (click)=\"openModal()\"><clr-icon shape=\"add\"></clr-icon>New Project</button>\n    <button class=\"btn btn-link\" (click)=\"deleteSelectedProjects()\"><clr-icon shape=\"close\"></clr-icon>Delete</button>\n    <create-project (create)=\"createProject($event)\" (openModal)=\"openModal($event)\"></create-project>\n  </div>\n  <div class=\"col-xs-4\">\n    <filter-project (filter)=\"filterProjects($event)\"></filter-project>\n    <search-project (search)=\"searchProjects($event)\"></search-project>\n  </div>\n  <div class=\"col-lg-12 col-md-12 col-sm-12 col-xs-12\">\n    <list-project (actionPerform)=\"actionPerform($event)\"></list-project>\n  </div>\n</div>"

/***/ }),

/***/ 733:
/***/ (function(module, exports) {

module.exports = "<input type=\"text\" placeholder=\"Search for projects\" #searchProject (keyup.enter)=\"doSearch(searchProject.value)\" >"

/***/ }),

/***/ 734:
/***/ (function(module, exports) {

module.exports = "<clr-modal [(clrModalOpen)]=\"create_policy_opened\">\n  <h3 class=\"modal-title\">Add Policy</h3>\n  <div class=\"modal-body\">\n    <form>\n      <section class=\"form-block\">\n        <div class=\"form-group\">\n          <label for=\"policy_name\" class=\"col-md-4\">Name</label>\n          <input type=\"text\" class=\"col-md-8\" id=\"policy_name\" size=\"20\"> \n        </div>\n        <div class=\"form-group\">\n          <label for=\"policy_description\" class=\"col-md-4\">Description</label>\n          <input type=\"text\" class=\"col-md-8\" id=\"policy_description\" size=\"20\"> \n        </div>\n        <div class=\"form-group\">\n          <label class=\"col-md-4\">Enable</label>\n          <div class=\"checkbox-inline\">\n            <input type=\"checkbox\" id=\"policy_enable\">\n            <label for=\"policy_enable\"></label>\n          </div>\n        </div>\n        <div class=\"form-group\">\n          <label for=\"destination_name\" class=\"col-md-4\">Destination name</label>\n          <div class=\"select\">\n            <select id=\"destination_name\">\n              <option>10.117.5.114</option>\n              <option>10.117.5.61</option>\n            </select>\n          </div>\n          <div class=\"checkbox-inline\">\n            <input type=\"checkbox\" id=\"check_new\">\n            <label for=\"check_new\">New destination</label>\n          </div>\n        </div>\n        <div class=\"form-group\">\n          <label for=\"destination_url\" class=\"col-md-4\">Destination URL</label>\n          <input type=\"text\" class=\"col-md-8\" id=\"destination_url\" size=\"20\">\n        </div>\n        <div class=\"form-group\">\n          <label for=\"destination_username\" class=\"col-md-4\">Username</label>\n          <input type=\"text\" class=\"col-md-8\" id=\"destination_username\" size=\"20\">\n        </div>\n        <div class=\"form-group\">\n          <label for=\"destination_password\" class=\"col-md-4\">Password</label>\n          <input type=\"text\" class=\"col-md-8\" id=\"destination_password\" size=\"20\">\n        </div>\n      </section>\n    </form>\n  </div>\n  <div class=\"modal-footer\">\n      <button type=\"button\" class=\"btn btn-outline\">Test Connection</button>\n      <button type=\"button\" class=\"btn btn-outline\" (click)=\"create_policy_opened = false\">Cancel</button>\n      <button type=\"button\" class=\"btn btn-primary\" (click)=\"create_policy_opened = false\">Ok</button>\n  </div>\n</clr-modal>\n<div class=\"row\">\n  <div class=\"col-lg-12 col-md-12 col-sm-12 col-xs-12\">\n    <div class=\"row flex-items-xs-between\">\n      <div class=\"col-xs-4\">\n        <button class=\"btn btn-sm\" (click)=\"create_policy_opened = true\">New Policy</button>\n      </div>\n      <div class=\"col-xs-4\">\n        <input type=\"text\" placeholder=\"Search for policies\">\n      </div>\n    </div>\n    <clr-datagrid>\n      <clr-dg-column>Name</clr-dg-column>\n      <clr-dg-column>Status</clr-dg-column>\n      <clr-dg-column>Destination</clr-dg-column>      \n      <clr-dg-column>Last start time</clr-dg-column>\n      <clr-dg-column>Description</clr-dg-column>\n      <clr-dg-column>Action</clr-dg-column>\n      <clr-dg-row *ngFor=\"let p of policies\">\n        <clr-dg-cell>{{p.name}}</clr-dg-cell>\n        <clr-dg-cell>{{p.status}}</clr-dg-cell>\n        <clr-dg-cell>{{p.destination}}</clr-dg-cell>\n        <clr-dg-cell>{{p.lastStartTime}}</clr-dg-cell>\n        <clr-dg-cell>{{p.description}}</clr-dg-cell>\n        <clr-dg-cell>\n          <clr-dropdown [clrMenuPosition]=\"'bottom-left'\">\n              <button class=\"btn btn-sm btn-link\" clrDropdownToggle>\n                Actions\n                <clr-icon shape=\"caret down\"></clr-icon>\n              </button>\n              <div class=\"dropdown-menu\">\n                <a href=\"javascript:void(0)\" clrDropdownItem>Enable</a>\n                <a href=\"javascript:void(0)\" clrDropdownItem>Disable</a>\n              </div>\n            </clr-dropdown>\n        </clr-dg-cell>\n      </clr-dg-row>\n      <clr-dg-footer>{{policies.length}} item(s)</clr-dg-footer>\n    </clr-datagrid>\n    <div class=\"row flex-items-xs-between flex-items-xs-bottom\">\n      <div class=\"col-xs-4\">\n        <span>Replication Jobs for 'project01/sync_01'</span>\n      </div>\n      <div class=\"col-xs-4\">\n        <clr-dropdown [clrMenuPosition]=\"'bottom-left'\">\n          <button class=\"btn btn-sm btn-outline-primary\" clrDropdownToggle>\n            All\n            <clr-icon shape=\"caret down\"></clr-icon>\n          </button>\n          <div class=\"dropdown-menu\">\n            <a href=\"javascript:void(0)\" clrDropdownItem>Finished</a>            \n            <a href=\"javascript:void(0)\" clrDropdownItem>Running</a>\n            <a href=\"javascript:void(0)\" clrDropdownItem>Error</a>\n            <a href=\"javascript:void(0)\" clrDropdownItem>Stopped</a>\n            <a href=\"javascript:void(0)\" clrDropdownItem>Retrying</a>\n          </div>\n        </clr-dropdown>\n        <input type=\"text\" placeholder=\"Search for jobs\">\n      </div>\n    </div>\n    <clr-datagrid>\n      <clr-dg-column>Name</clr-dg-column>\n      <clr-dg-column>Status</clr-dg-column>\n      <clr-dg-column>Operation</clr-dg-column>      \n      <clr-dg-column>Creation time</clr-dg-column>\n      <clr-dg-column>End time</clr-dg-column>\n      <clr-dg-column>Logs</clr-dg-column>\n      <clr-dg-row *ngFor=\"let j of jobs\">\n        <clr-dg-cell>{{j.name}}</clr-dg-cell>\n        <clr-dg-cell>{{j.status}}</clr-dg-cell>\n        <clr-dg-cell>{{j.operation}}</clr-dg-cell>\n        <clr-dg-cell>{{j.creationTime}}</clr-dg-cell>\n        <clr-dg-cell>{{j.endTime}}</clr-dg-cell>\n        <clr-dg-cell></clr-dg-cell>\n      </clr-dg-row>\n      <clr-dg-footer>{{jobs.length}} item(s)</clr-dg-footer>\n    </clr-datagrid>     \n  </div>\n</div>"

/***/ }),

/***/ 735:
/***/ (function(module, exports) {

module.exports = "<div class=\"row\">\n  <div class=\"col-lg-12 col-md-12 col-sm-12 col-xs-12\">  \n    <div class=\"row flex-items-lg-right\">\n      <div class=\"col-lg-3 col-md-3 col-sm-12 col-xs-12\">\n        <clr-dropdown [clrMenuPosition]=\"'bottom-left'\">\n          <button class=\"btn btn-sm btn-outline-primary\" clrDropdownToggle>\n            My Projects\n            <clr-icon shape=\"caret down\"></clr-icon>\n          </button>\n          <div class=\"dropdown-menu\">\n            <a href=\"#/project\" clrDropdownItem>My Projects</a>\n            <a href=\"#/project\" clrDropdownItem>Public Projects</a>\n          </div>\n        </clr-dropdown>\n        <input type=\"text\" placeholder=\"Search for projects\">\n      </div>\n    </div>\n    <clr-datagrid>\n      <clr-dg-column>Name</clr-dg-column>\n      <clr-dg-column>Status</clr-dg-column>\n      <clr-dg-column>Tag</clr-dg-column>\n      <clr-dg-column>Author</clr-dg-column>\n      <clr-dg-column>Docker version</clr-dg-column>\n      <clr-dg-column>Created</clr-dg-column>\n      <clr-dg-column>Pull Command</clr-dg-column>\n      <clr-dg-row *ngFor=\"let r of repos\">\n        <clr-dg-cell>{{r.name}}</clr-dg-cell>\n        <clr-dg-cell>{{r.status}}</clr-dg-cell>\n        <clr-dg-cell>{{r.tag}}</clr-dg-cell>\n        <clr-dg-cell>{{r.author}}</clr-dg-cell>\n        <clr-dg-cell>{{r.dockerVersion}}</clr-dg-cell>\n        <clr-dg-cell>{{r.created}}</clr-dg-cell>\n        <clr-dg-cell>{{r.pullCommand}}</clr-dg-cell>\n      </clr-dg-row>\n      <clr-dg-footer>{{repos.length}} item(s)</clr-dg-footer>\n    </clr-datagrid>\n  </div>\n</div>"

/***/ }),

/***/ 736:
/***/ (function(module, exports) {

module.exports = ""

/***/ }),

/***/ 763:
/***/ (function(module, exports, __webpack_require__) {

module.exports = __webpack_require__(408);


/***/ })

},[763]);
//# sourceMappingURL=main.bundle.js.map