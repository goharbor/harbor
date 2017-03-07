import { Component, OnInit, OnDestroy } from '@angular/core';
import { Router } from '@angular/router';

const defaultInterval = 1000;
const defaultLeftTime = 5;
 
@Component({
    selector: 'page-not-found',
    templateUrl: "not-found.component.html",
    styleUrls: ['not-found.component.css']
})
export class PageNotFoundComponent implements OnInit, OnDestroy{
    private leftSeconds: number = defaultLeftTime;
    private timeInterval: any = null;

    constructor(private router: Router){}

    ngOnInit(): void {
        if(!this.timeInterval){
            this.timeInterval = setInterval(interval => {
                this.leftSeconds--;
                if(this.leftSeconds <= 0){
                    this.router.navigate(['harbor']);
                    clearInterval(this.timeInterval);
                }
            }, defaultInterval);
        }
    }

    ngOnDestroy(): void {
        if(this.timeInterval){
             clearInterval(this.timeInterval);
        }
    }
}