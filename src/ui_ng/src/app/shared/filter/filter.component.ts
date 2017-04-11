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
    
    placeHolder: string = "";
    filterTerms = new Subject<string>();

    @Output("filter") filterEvt = new EventEmitter<string>();

    @Input() currentValue: any;
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