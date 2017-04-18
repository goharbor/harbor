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


@Component({
    selector: 'grid-filter',
    templateUrl: 'filter.component.html',
    styleUrls: ['filter.component.css']
})

export class FilterComponent implements OnInit {
    
    private placeHolder: string = "";
    private filterTerms = new Subject<string>();

    @Output("filter") private filterEvt = new EventEmitter<string>();

    @Input() currentValue;
    @Input("filterPlaceholder")
    public set flPlaceholder(placeHolder: string) {
        this.placeHolder = placeHolder;
    }

    ngOnInit(): void {
        this.filterTerms
        .debounceTime(500)
        //.distinctUntilChanged()
        .subscribe(terms => {
            this.filterEvt.emit(terms);
        });
        
    }

    valueChange(): void {
        //Send out filter terms
        this.filterTerms.next(this.currentValue.trim());
    }
}