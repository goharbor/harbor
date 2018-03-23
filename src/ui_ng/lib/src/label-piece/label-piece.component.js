var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { Component, Input } from '@angular/core';
import 'rxjs/add/operator/debounceTime';
import 'rxjs/add/operator/distinctUntilChanged';
import { LABEL_PIEICE_TEMPLATE, LABEL_PIEICE_STYLES } from './label-piece.template';
var LabelPieceComponent = (function () {
    function LabelPieceComponent() {
    }
    LabelPieceComponent.prototype.ngOnInit = function () {
    };
    return LabelPieceComponent;
}());
__decorate([
    Input(),
    __metadata("design:type", Object)
], LabelPieceComponent.prototype, "label", void 0);
LabelPieceComponent = __decorate([
    Component({
        selector: 'hbr-label-piece',
        styles: [LABEL_PIEICE_STYLES],
        template: LABEL_PIEICE_TEMPLATE
    })
], LabelPieceComponent);
export { LabelPieceComponent };
//# sourceMappingURL=label-piece.component.js.map