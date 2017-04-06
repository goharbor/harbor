import { Component } from '@angular/core';

import { AppConfigService } from '../../app-config.service';

@Component({
    selector: 'about-dialog',
    templateUrl: "about-dialog.component.html",
    styleUrls: ["about-dialog.component.css"]
})
export class AboutDialogComponent {
    private opened: boolean = false;
    private build: string = "4276418";

    constructor(private appConfigService: AppConfigService) { }

    public get version(): string {
        let appConfig = this.appConfigService.getConfig();
        return appConfig?appConfig.harbor_version: "n/a";
    }

    public open(): void {
        this.opened = true;
    }

    public close(): void {
        this.opened = false;
    }
}