import { Component } from '@angular/core';
import { SessionService } from '../../../shared/services/session.service';

@Component({
    selector: 'app-gc-page',
    templateUrl: './gc-page.component.html',
    styleUrls: ['./gc-page.component.scss'],
})
export class GcPageComponent {
    inProgress: boolean = true;
    constructor(private session: SessionService) {}

    public get hasAdminRole(): boolean {
        return (
            this.session.getCurrentUser() &&
            this.session.getCurrentUser().has_admin_role
        );
    }

    getGcStatus(status: boolean) {
        this.inProgress = status;
    }
}
