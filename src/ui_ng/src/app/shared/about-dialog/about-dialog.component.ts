import { Component } from '@angular/core';
 
@Component({
    selector: 'about-dialog',
    templateUrl: "about-dialog.component.html",
    styleUrls: ["about-dialog.component.css"]
})
export class AboutDialogComponent {
    private opened: boolean = false;
    private version: string ="0.4.1";
    private build: string ="4276418";

    public open(): void {
        this.opened = true;
    }

    public close(): void {
        this.opened = false;
    }
}