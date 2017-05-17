webpackJsonp([1,4],{

/***/ 141:
/***/ (function(module, exports) {

function webpackEmptyContext(req) {
	throw new Error("Cannot find module '" + req + "'.");
}
webpackEmptyContext.keys = function() { return []; };
webpackEmptyContext.resolve = webpackEmptyContext;
module.exports = webpackEmptyContext;
webpackEmptyContext.id = 141;


/***/ }),

/***/ 142:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
__webpack_require__(162);
var platform_browser_dynamic_1 = __webpack_require__(157);
var core_1 = __webpack_require__(10);
var environment_1 = __webpack_require__(161);
var _1 = __webpack_require__(160);
if (environment_1.environment.production) {
    core_1.enableProdMode();
}
platform_browser_dynamic_1.platformBrowserDynamic().bootstrapModule(_1.AppModule);
//# sourceMappingURL=/Users/druk/Sites/harbor/src/src/src/main.js.map

/***/ }),

/***/ 158:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
Object.defineProperty(exports, "__esModule", { value: true });
var platform_browser_1 = __webpack_require__(33);
var core_1 = __webpack_require__(10);
var forms_1 = __webpack_require__(88);
var http_1 = __webpack_require__(156);
var clarity_angular_1 = __webpack_require__(90);
var app_component_1 = __webpack_require__(89);
var utils_module_1 = __webpack_require__(165);
var app_routing_1 = __webpack_require__(159);
var AppModule = (function () {
    function AppModule() {
    }
    return AppModule;
}());
AppModule = __decorate([
    core_1.NgModule({
        declarations: [
            app_component_1.AppComponent
        ],
        imports: [
            platform_browser_1.BrowserModule,
            forms_1.FormsModule,
            http_1.HttpModule,
            clarity_angular_1.ClarityModule.forRoot(),
            utils_module_1.UtilsModule,
            app_routing_1.ROUTING
        ],
        providers: [],
        bootstrap: [app_component_1.AppComponent]
    })
], AppModule);
exports.AppModule = AppModule;
//# sourceMappingURL=/Users/druk/Sites/harbor/src/src/src/app/app.module.js.map

/***/ }),

/***/ 159:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var router_1 = __webpack_require__(47);
exports.ROUTES = [
    { path: '', redirectTo: 'home', pathMatch: 'full' }
];
exports.ROUTING = router_1.RouterModule.forRoot(exports.ROUTES);
//# sourceMappingURL=/Users/druk/Sites/harbor/src/src/src/app/app.routing.js.map

/***/ }),

/***/ 160:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

function __export(m) {
    for (var p in m) if (!exports.hasOwnProperty(p)) exports[p] = m[p];
}
Object.defineProperty(exports, "__esModule", { value: true });
__export(__webpack_require__(89));
__export(__webpack_require__(158));
//# sourceMappingURL=/Users/druk/Sites/harbor/src/src/src/app/index.js.map

/***/ }),

/***/ 161:
/***/ (function(module, exports, __webpack_require__) {

"use strict";
// The file contents for the current environment will overwrite these during build.
// The build system defaults to the dev environment which uses `environment.ts`, but if you do
// `ng build --env=prod` then `environment.prod.ts` will be used instead.
// The list of which env maps to which file can be found in `angular-cli.json`.

Object.defineProperty(exports, "__esModule", { value: true });
exports.environment = {
    production: true
};
//# sourceMappingURL=/Users/druk/Sites/harbor/src/src/src/environments/environment.js.map

/***/ }),

/***/ 162:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
// This file includes polyfills needed by Angular 2 and is loaded before
// the app. You can add your own extra polyfills to this file.
__webpack_require__(179);
__webpack_require__(172);
__webpack_require__(168);
__webpack_require__(174);
__webpack_require__(173);
__webpack_require__(171);
__webpack_require__(170);
__webpack_require__(178);
__webpack_require__(167);
__webpack_require__(166);
__webpack_require__(176);
__webpack_require__(169);
__webpack_require__(177);
__webpack_require__(175);
__webpack_require__(180);
__webpack_require__(358);
//# sourceMappingURL=/Users/druk/Sites/harbor/src/src/src/polyfills.js.map

/***/ }),

/***/ 163:
/***/ (function(module, exports, __webpack_require__) {

"use strict";
/*
 * Hack while waiting for https://github.com/angular/angular/issues/6595 to be fixed.
 */

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
Object.defineProperty(exports, "__esModule", { value: true });
var core_1 = __webpack_require__(10);
var router_1 = __webpack_require__(47);
var HashListener = (function () {
    function HashListener(route) {
        var _this = this;
        this.route = route;
        this.sub = this.route.fragment.subscribe(function (f) {
            _this.scrollToAnchor(f, false);
        });
    }
    HashListener.prototype.ngOnInit = function () {
        this.scrollToAnchor(this.route.snapshot.fragment, false);
    };
    HashListener.prototype.scrollToAnchor = function (hash, smooth) {
        if (smooth === void 0) { smooth = true; }
        if (hash) {
            var element = document.querySelector("#" + hash);
            if (element) {
                element.scrollIntoView({
                    behavior: smooth ? "smooth" : "instant",
                    block: "start"
                });
            }
        }
    };
    HashListener.prototype.ngOnDestroy = function () {
        this.sub.unsubscribe();
    };
    return HashListener;
}());
HashListener = __decorate([
    core_1.Directive({
        selector: "[hash-listener]",
        host: {
            "[style.position]": "'relative'"
        }
    }),
    __metadata("design:paramtypes", [typeof (_a = typeof router_1.ActivatedRoute !== "undefined" && router_1.ActivatedRoute) === "function" && _a || Object])
], HashListener);
exports.HashListener = HashListener;
var _a;
//# sourceMappingURL=/Users/druk/Sites/harbor/src/src/src/utils/hash-listener.directive.js.map

/***/ }),

/***/ 164:
/***/ (function(module, exports, __webpack_require__) {

"use strict";
/*
 * Hack while waiting for https://github.com/angular/angular/issues/6595 to be fixed.
 */

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
Object.defineProperty(exports, "__esModule", { value: true });
var core_1 = __webpack_require__(10);
var router_1 = __webpack_require__(47);
var ScrollSpy = (function () {
    function ScrollSpy(renderer) {
        this.renderer = renderer;
        this.anchors = [];
        this.throttle = false;
    }
    Object.defineProperty(ScrollSpy.prototype, "links", {
        set: function (routerLinks) {
            var _this = this;
            this.anchors = routerLinks.map(function (routerLink) { return "#" + routerLink.fragment; });
            this.sub = routerLinks.changes.subscribe(function () {
                _this.anchors = routerLinks.map(function (routerLink) { return "#" + routerLink.fragment; });
            });
        },
        enumerable: true,
        configurable: true
    });
    ScrollSpy.prototype.handleEvent = function () {
        var _this = this;
        this.scrollPosition = this.scrollable.scrollTop;
        if (!this.throttle) {
            window.requestAnimationFrame(function () {
                var currentLinkIndex = _this.findCurrentAnchor() || 0;
                _this.linkElements.forEach(function (link, index) {
                    _this.renderer.setElementClass(link.nativeElement, "active", index === currentLinkIndex);
                });
                _this.throttle = false;
            });
        }
        this.throttle = true;
    };
    ScrollSpy.prototype.findCurrentAnchor = function () {
        for (var i = this.anchors.length - 1; i >= 0; i--) {
            var anchor = this.anchors[i];
            if (this.scrollable.querySelector(anchor) && this.scrollable.querySelector(anchor).offsetTop <= this.scrollPosition) {
                return i;
            }
        }
    };
    ScrollSpy.prototype.ngOnInit = function () {
        this.scrollable.addEventListener("scroll", this);
    };
    ScrollSpy.prototype.ngOnDestroy = function () {
        this.scrollable.removeEventListener("scroll", this);
        if (this.sub) {
            this.sub.unsubscribe();
        }
    };
    return ScrollSpy;
}());
__decorate([
    core_1.Input("scrollspy"),
    __metadata("design:type", Object)
], ScrollSpy.prototype, "scrollable", void 0);
__decorate([
    core_1.ContentChildren(router_1.RouterLinkWithHref, { descendants: true }),
    __metadata("design:type", typeof (_a = typeof core_1.QueryList !== "undefined" && core_1.QueryList) === "function" && _a || Object),
    __metadata("design:paramtypes", [typeof (_b = typeof core_1.QueryList !== "undefined" && core_1.QueryList) === "function" && _b || Object])
], ScrollSpy.prototype, "links", null);
__decorate([
    core_1.ContentChildren(router_1.RouterLinkWithHref, { descendants: true, read: core_1.ElementRef }),
    __metadata("design:type", typeof (_c = typeof core_1.QueryList !== "undefined" && core_1.QueryList) === "function" && _c || Object)
], ScrollSpy.prototype, "linkElements", void 0);
ScrollSpy = __decorate([
    core_1.Directive({
        selector: "[scrollspy]",
    }),
    __metadata("design:paramtypes", [typeof (_d = typeof core_1.Renderer !== "undefined" && core_1.Renderer) === "function" && _d || Object])
], ScrollSpy);
exports.ScrollSpy = ScrollSpy;
var _a, _b, _c, _d;
//# sourceMappingURL=/Users/druk/Sites/harbor/src/src/src/utils/scrollspy.directive.js.map

/***/ }),

/***/ 165:
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
Object.defineProperty(exports, "__esModule", { value: true });
var core_1 = __webpack_require__(10);
var hash_listener_directive_1 = __webpack_require__(163);
var scrollspy_directive_1 = __webpack_require__(164);
var clarity_angular_1 = __webpack_require__(90);
var common_1 = __webpack_require__(40);
var UtilsModule = (function () {
    function UtilsModule() {
    }
    return UtilsModule;
}());
UtilsModule = __decorate([
    core_1.NgModule({
        imports: [
            common_1.CommonModule,
            clarity_angular_1.ClarityModule.forChild()
        ],
        declarations: [
            hash_listener_directive_1.HashListener,
            scrollspy_directive_1.ScrollSpy
        ],
        exports: [
            hash_listener_directive_1.HashListener,
            scrollspy_directive_1.ScrollSpy
        ]
    })
], UtilsModule);
exports.UtilsModule = UtilsModule;
//# sourceMappingURL=/Users/druk/Sites/harbor/src/src/src/utils/utils.module.js.map

/***/ }),

/***/ 320:
/***/ (function(module, exports, __webpack_require__) {

exports = module.exports = __webpack_require__(38)(false);
// imports


// module
exports.push([module.i, ".clr-icon.clr-clarity-logo {\n  background-image: url(/harbor/images/vmw_oss.svg); }\n\n.hero {\n  background-color: #ddd;\n  text-align: center;\n  padding-bottom: 3em;\n  padding-top: 3em;\n  width: 100%; }\n  .hero .btn-custom {\n    display: inline-block;\n    text-align: center;\n    margin: auto; }\n\n.hero-image img {\n  max-width: 360px; }\n\n.icon {\n  display: inline-block;\n  height: 32px;\n  vertical-align: middle;\n  width: 32px; }\n  .icon.icon-github {\n    background: url(/harbor/images/github_icon.svg) no-repeat left -2px; }\n\n.nav-group label {\n  display: block;\n  margin-bottom: 1em; }\n\n.sidenav .nav-link {\n  padding: 3px 6px; }\n  .sidenav .nav-link:hover {\n    background: #eee; }\n  .sidenav .nav-link.active {\n    background: #d9e4ea;\n    color: #000; }\n\n.section {\n  padding: .5em 0; }\n\n.contributor {\n  border-radius: 50%;\n  border: 1px solid #ccc;\n  margin-bottom: 1.5em;\n  margin-right: 2.5em;\n  max-width: 104px;\n  text-decoration: none; }\n\n#license {\n  padding-bottom: 48vh; }\n", ""]);

// exports


/*** EXPORTS FROM exports-loader ***/
module.exports = module.exports.toString();

/***/ }),

/***/ 329:
/***/ (function(module, exports) {

module.exports = "<clr-main-container>\n    <header class=\"header header-6\">\n        <div class=\"branding\">\n            <a href=\"https://vmware.github.io/\" class=\"nav-link\">\n                <span class=\"clr-icon clr-clarity-logo\"></span>\n                <span class=\"title\">VMware&reg; Open Source Program Office</span>\n            </a>\n        </div>\n    </header>\n    <div class=\"hero\">\n        <div class=\"hero-image\"><img src=\"/images/harbor.png\" alt=\"\"></div>\n        <h3>An Enterprise-class Container Registry Server based on Docker Distribution</h3>\n        <p><a href=\"https://github.com/vmware/harbor\" class=\"btn btn-primary\"><i class=\"icon icon-github\"></i> Fork Harbor&trade;</a></p>\n    </div>\n    <div class=\"content-container\">\n        <div id=\"content-area\" class=\"content-area\" hash-listener #scrollable>\n            <div id=\"overview\" class=\"section\">\n                <h2>What is Harbor&trade;</h2>\n\n                <p>Project Harbor&trade; is an enterprise-class registry server that stores and distributes Docker images. Harbor&trade; extends the open source Docker Distribution by adding the functionalities usually required by an enterprise, such as security, identity and management. As an enterprise private registry, Harbor&trade; offers better performance and security. Having a registry closer to the build and run environment improves the image transfer efficiency. Harbor&trade; supports the setup of multiple registries and has images replicated between them. With Harbor&trade;, the images are stored within the private registry, keeping the bits and intellectual properties behind the company firewall. In addition, Harbor&trade; offers advanced security features, such as user management, access control and activity auditing.</p>\n\n                <br>\n\n                <ul>\n                    <li><strong>Role Based Access Control</strong> - Users and docker repositories are organized via \"projects\", a user can have different permission for images under a namespace.</li>\n                    <li><strong>Image replication</strong> - Images can be replicated (synchronized) between multiple registry instances. Great for load balancing, high availability, hybrid and multi-cloud scenarios.</li>\n                    <li><strong>Graphical user portal</strong> - User can easily browse, search docker repositories, manage projects/namespaces.</li>\n                    <li><strong>AD/LDAP support</strong> - Harbor&trade; integrates with existing enterprise AD/LDAP for user authentication and management.</li>\n                    <li><strong>Auditing</strong> - All the operations to the repositories are tracked and can be used for auditing purpose.</li>\n                    <li><strong>Internationalization</strong> - Already localized for English, Chinese, German, Japanese and Russian. More languages can be added.</li>\n                    <li><strong>RESTful API</strong> - RESTful APIs are provided for most administrative operations of Harbor&trade;. The integration with other management softwares becomes easy.</li>\n                    <li><strong>Easy deployment</strong> - Provide both an online and offline installer. Besides, a virtual appliance for vSphere platform (OVA) is available.</li>\n                </ul>\n\n                <p>See the <a href=\"https://github.com/vmware/harbor/blob/master/README.md\">README</a> for more information.</p>\n            </div>\n\n            <div id=\"gettingHarbor\" class=\"section\">\n                <h2>Getting Harbor&trade;</h2>\n\n                <p>Harbor&trade; can be installed on a Linux host. It can also be deployed as a virtual appliance on vSphere. Please download Harbor&trade; from the release page.</p>\n                <p>Refer to <a href=\"https://github.com/vmware/harbor/tree/master/docs\">Harbor’s documents</a> for more information.</p>\n            </div>\n\n            <div id=\"gettingStarted\" class=\"section\">\n                <h2>Getting Started</h2>\n                <p>We've provided a guide to help get you started:</p>\n\n                <a href=\"https://github.com/vmware/admiral/wiki/Developer-Guide\" class=\"btn btn-outline\">README</a>\n                <a href=\"https://github.com/vmware/harbor/blob/master/docs/installation_guide.md\" class=\"btn btn-outline\">Install Guide of Harbor&trade;</a>\n                <a href=\"https://github.com/vmware/harbor/blob/master/docs/user_guide.md\" class=\"btn btn-outline\">User Guide of Harbor&trade;</a>\n                <a href=\"https://github.com/vmware/harbor/blob/master/docs/installation_guide_ova.md\" title=\"Installation Guide of Harbor&trade; Virtual Appliance\" class=\"btn btn-outline\">Installation Guide of HVA</a>\n                <a href=\"https://github.com/vmware/harbor/blob/master/docs/user_guide_ova.md\" title=\"User Guide of Harbor&trade; Virtual Appliance\" class=\"btn btn-outline\">User Guide of HVA</a>\n            </div>\n\n            <div id=\"contributors\" class=\"section\">\n                <h2>Contributors</h2>\n\n                <p>\n                    <a title=\"reasonerjt\" href=\"https://github.com/reasonerjt\"><img class=\"contributor\" alt=\"reasonerjt\" src=\"https://avatars3.githubusercontent.com/u/2390463?v=3\" ></a>\n                    <a title=\"wknet123\" href=\"https://github.com/wknet123\"><img class=\"contributor\" alt=\"wknet123\" src=\"https://avatars0.githubusercontent.com/u/5027302?v=3\" ></a>\n                    <a title=\"ywk253100\" href=\"https://github.com/ywk253100\"><img class=\"contributor\" alt=\"ywk253100\" src=\"https://avatars0.githubusercontent.com/u/5835782?v=3\" ></a>\n                    <a title=\"hainingzhang\" href=\"https://github.com/hainingzhang\"><img class=\"contributor\" alt=\"hainingzhang\" src=\"https://avatars1.githubusercontent.com/u/2161887?v=3\" ></a>\n                    <a title=\"steven-zou\" href=\"https://github.com/steven-zou\"><img class=\"contributor\" alt=\"steven-zou\" src=\"https://avatars3.githubusercontent.com/u/5753287?v=3\" ></a>\n                    <a title=\"wemeya\" href=\"https://github.com/wemeya\"><img class=\"contributor\" alt=\"wemeya\" src=\"https://avatars2.githubusercontent.com/u/12540577?v=3\" ></a>\n                    <a title=\"yhua123\" href=\"https://github.com/yhua123\"><img class=\"contributor\" alt=\"yhua123\" src=\"https://avatars1.githubusercontent.com/u/19166125?v=3\" ></a>\n                    <a title=\"wy65701436\" href=\"https://github.com/wy65701436\"><img class=\"contributor\" alt=\"wy65701436\" src=\"https://avatars0.githubusercontent.com/u/2841473?v=3\" ></a>\n                    <a title=\"invalid-email-address\" href=\"https://github.com/invalid-email-address\"><img class=\"contributor\" alt=\"invalid-email-address\" src=\"https://avatars3.githubusercontent.com/u/148100?v=3\" ></a>\n                    <a title=\"saga92\" href=\"https://github.com/saga92\"><img class=\"contributor\" alt=\"saga92\" src=\"https://avatars1.githubusercontent.com/u/5730235?v=3\" ></a>\n                    <a title=\"xiahaoshawn\" href=\"https://github.com/xiahaoshawn\"><img class=\"contributor\" alt=\"xiahaoshawn\" src=\"https://avatars0.githubusercontent.com/u/10750864?v=3\" ></a>\n                    <a title=\"Erkak\" href=\"https://github.com/Erkak\"><img class=\"contributor\" alt=\"Erkak\" src=\"https://avatars2.githubusercontent.com/u/15937486?v=3\" ></a>\n                    <a title=\"hmwenchen\" href=\"https://github.com/hmwenchen\"><img class=\"contributor\" alt=\"hmwenchen\" src=\"https://avatars3.githubusercontent.com/u/16629561?v=3\" ></a>\n                    <a title=\"perhapszzy\" href=\"https://github.com/perhapszzy\"><img class=\"contributor\" alt=\"perhapszzy\" src=\"https://avatars1.githubusercontent.com/u/7953637?v=3\" ></a>\n                    <a title=\"zgdxiaoxiao\" href=\"https://github.com/zgdxiaoxiao\"><img class=\"contributor\" alt=\"zgdxiaoxiao\" src=\"https://avatars3.githubusercontent.com/u/19501217?v=3\" ></a>\n                    <a title=\"victoriazhengwf\" href=\"https://github.com/victoriazhengwf\"><img class=\"contributor\" alt=\"victoriazhengwf\" src=\"https://avatars0.githubusercontent.com/u/17972009?v=3\" ></a>\n                    <a title=\"rikatz\" href=\"https://github.com/rikatz\"><img class=\"contributor\" alt=\"rikatz\" src=\"https://avatars3.githubusercontent.com/u/7182341?v=3\" ></a>\n                    <a title=\"senk\" href=\"https://github.com/senk\"><img class=\"contributor\" alt=\"senk\" src=\"https://avatars1.githubusercontent.com/u/710568?v=3\" ></a>\n                    <a title=\"AlexZeitler\" href=\"https://github.com/AlexZeitler\"><img class=\"contributor\" alt=\"AlexZeitler\" src=\"https://avatars2.githubusercontent.com/u/287480?v=3\" ></a>\n                    <a title=\"ScorpioCPH\" href=\"https://github.com/ScorpioCPH\"><img class=\"contributor\" alt=\"ScorpioCPH\" src=\"https://avatars1.githubusercontent.com/u/5319646?v=3\" ></a>\n                    <a title=\"redkafei\" href=\"https://github.com/redkafei\"><img class=\"contributor\" alt=\"redkafei\" src=\"https://avatars1.githubusercontent.com/u/8327386?v=3\" ></a>\n                    <a title=\"int32bit\" href=\"https://github.com/int32bit\"><img class=\"contributor\" alt=\"int32bit\" src=\"https://avatars2.githubusercontent.com/u/5260798?v=3\" ></a>\n                    <a title=\"tobegit3hub\" href=\"https://github.com/tobegit3hub\"><img class=\"contributor\" alt=\"tobegit3hub\" src=\"https://avatars3.githubusercontent.com/u/2715000?v=3\" ></a>\n                    <a title=\"amandaz\" href=\"https://github.com/amandaz\"><img class=\"contributor\" alt=\"amandaz\" src=\"https://avatars0.githubusercontent.com/u/2898608?v=3\" ></a>\n                    <a title=\"laz2\" href=\"https://github.com/laz2\"><img class=\"contributor\" alt=\"laz2\" src=\"https://avatars2.githubusercontent.com/u/800356?v=3\" ></a>\n                    <a title=\"nagarjung\" href=\"https://github.com/nagarjung\"><img class=\"contributor\" alt=\"nagarjung\" src=\"https://avatars1.githubusercontent.com/u/9403528?v=3\" ></a>\n                    <a title=\"alanwooo\" href=\"https://github.com/alanwooo\"><img class=\"contributor\" alt=\"alanwooo\" src=\"https://avatars2.githubusercontent.com/u/12868735?v=3\" ></a>\n                    <a title=\"liubin\" href=\"https://github.com/liubin\"><img class=\"contributor\" alt=\"liubin\" src=\"https://avatars2.githubusercontent.com/u/1212008?v=3\" ></a>\n                    <a title=\"feilengcui008\" href=\"https://github.com/feilengcui008\"><img class=\"contributor\" alt=\"feilengcui008\" src=\"https://avatars3.githubusercontent.com/u/4131736?v=3\" ></a>\n                    <a title=\"sigsbee\" href=\"https://github.com/sigsbee\"><img class=\"contributor\" alt=\"sigsbee\" src=\"https://avatars1.githubusercontent.com/u/23101283?v=3\" ></a>\n                    </p>\n            </div>\n\n            <div id=\"contributing\" class=\"section\">\n                <h2>Contributing</h2>\n\n                <p>We welcome contributions from the community. If you wish to contribute code and you have not signed our contributor license agreement (CLA), our bot will update the issue when you open a pull request. For any questions about the CLA process, please refer to our <a href=\"https://cla.vmware.com/faq\">FAQ</a>.</p>\n            </div>\n\n            <div id=\"license\" class=\"section\">\n                <h2>License</h2>\n\n                <p>Harbor&trade; is available under the <a href=\"https://github.com/vmware/harbor/blob/master/LICENSE\">Apache 2 license</a>.</p>\n            </div>\n        </div>\n        <nav class=\"sidenav\" [clr-nav-level]=\"2\">\n            <section class=\"sidenav-content\">\n                <section class=\"nav-group\" [scrollspy]=\"scrollable\">\n                    <label><a class=\"nav-link active\" routerLink=\".\" routerLinkActive=\"active\" fragment=\"overview\">Overview</a></label>\n                    <label class=\"bump-down\"><a class=\"nav-link\" routerLink=\".\" fragment=\"gettingHarbor\">Getting Harbor&trade;</a></label>\n                    <label class=\"bump-down\"><a class=\"nav-link\" routerLink=\".\" fragment=\"gettingStarted\">Getting Started</a></label>\n                    <label class=\"bump-down\"><a class=\"nav-link\" routerLink=\".\" fragment=\"contributors\">Contributors</a></label>\n                    <label class=\"bump-down\"><a class=\"nav-link\" routerLink=\".\" fragment=\"contributing\">Contributing</a></label>\n                    <label class=\"bump-down\"><a class=\"nav-link\" routerLink=\".\" fragment=\"license\">License</a></label>\n                    <label class=\"bump-down\"><a class=\"nav-link\" href=\"https://vmware.github.io/harbor/cn/\">中文版</a></label>\n                </section>\n            </section>\n        </nav>\n    </div>\n</clr-main-container>\n"

/***/ }),

/***/ 360:
/***/ (function(module, exports, __webpack_require__) {

module.exports = __webpack_require__(142);


/***/ }),

/***/ 89:
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
Object.defineProperty(exports, "__esModule", { value: true });
var core_1 = __webpack_require__(10);
var router_1 = __webpack_require__(47);
var AppComponent = (function () {
    function AppComponent(router) {
        this.router = router;
    }
    return AppComponent;
}());
AppComponent = __decorate([
    core_1.Component({
        selector: 'my-app',
        template: __webpack_require__(329),
        styles: [__webpack_require__(320)]
    }),
    __metadata("design:paramtypes", [typeof (_a = typeof router_1.Router !== "undefined" && router_1.Router) === "function" && _a || Object])
], AppComponent);
exports.AppComponent = AppComponent;
var _a;
//# sourceMappingURL=/Users/druk/Sites/harbor/src/src/src/app/app.component.js.map

/***/ })

},[360]);
//# sourceMappingURL=main.bundle.js.map