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
import { Component, Input, Output, OnInit, EventEmitter } from '@angular/core';
import { Subject } from 'rxjs/Subject';
import { Observable } from 'rxjs/Observable';

import 'rxjs/add/operator/debounceTime';
import 'rxjs/add/operator/distinctUntilChanged';

import { LABEL_PIEICE_TEMPLATE, LABEL_PIEICE_STYLES } from './label-piece.template';
import {Label} from "../service/interface";


@Component({
    selector: 'hbr-label-piece',
    styles: [LABEL_PIEICE_STYLES],
    template: LABEL_PIEICE_TEMPLATE
})

export class LabelPieceComponent implements OnInit {
    @Input() label: Label;
    @Input() labelWidth: number;

    ngOnInit(): void {
    }
}