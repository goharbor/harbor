import {
    Component, Input, OnChanges,
    OnInit, SimpleChanges
} from "@angular/core";
import { LastTrigger } from "../webhook";


@Component({
    selector: 'last-trigger',
    templateUrl: 'last-trigger.component.html',
    styleUrls: ['./last-trigger.component.scss']
})
export class LastTriggerComponent implements  OnInit , OnChanges {
    @Input()  inputLastTriggers: LastTrigger[];
    @Input()  webhookName: string;
    lastTriggers: LastTrigger[] = [];
    constructor() {
    }
    ngOnChanges(changes: SimpleChanges): void {
       if (changes && changes['inputLastTriggers']) {
           this.lastTriggers = [];
           this.inputLastTriggers.forEach(item => {
             if (this.webhookName === item.policy_name) {
                 this.lastTriggers.push(item);
             }
           });
       }
    }
    ngOnInit(): void {
    }
}
