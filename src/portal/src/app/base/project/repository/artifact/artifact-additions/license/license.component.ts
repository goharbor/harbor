import { Component, Input, OnInit } from '@angular/core';
import { AdditionsService } from '../additions.service';
import { AdditionLink } from '../../../../../../../../ng-swagger-gen/models/addition-link';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';
import { finalize } from 'rxjs/operators';

@Component({
    selector: 'hbr-artifact-license',
    templateUrl: './license.component.html',
    styleUrls: ['./license.component.scss'],
})
export class ArtifactLicenseComponent implements OnInit {
    @Input() licenseLink: AdditionLink;
    license: string;
    loading: boolean = false;

    constructor(
        private errorHandler: ErrorHandler,
        private additionsService: AdditionsService
    ) {}

    ngOnInit(): void {
        this.getLicense();
    }

    getLicense() {
        if (
            this.licenseLink &&
            !this.licenseLink.absolute &&
            this.licenseLink.href
        ) {
            this.loading = true;
            this.additionsService
                .getDetailByLink(this.licenseLink.href, false, true)
                .pipe(finalize(() => (this.loading = false)))
                .subscribe(
                    res => {
                        this.license = res;
                    },
                    error => {
                        this.errorHandler.error(error);
                    }
                );
        }
    }
}
