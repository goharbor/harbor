import { Component, EventEmitter, Input, OnInit, Output } from "@angular/core";
import { clone, dbEncodeURIComponent, formatSize } from "../../../../../../../../shared/units/utils";
import { UN_LOGGED_PARAM, YES } from "../../../../../../../../account/sign-in/sign-in.service";
import { ActivatedRoute, Router } from "@angular/router";
import { Accessory } from "ng-swagger-gen/models/accessory";
import { ArtifactService as NewArtifactService } from "ng-swagger-gen/services/artifact.service";
import { ErrorHandler } from "../../../../../../../../shared/units/error-handler";
import { finalize } from "rxjs/operators";
import { SafeUrl } from '@angular/platform-browser';
import { ArtifactService } from "../../../../artifact.service";
import { AccessoryQueryParams, artifactDefault } from '../../../../artifact';

export const ACCESSORY_PAGE_SIZE: number = 5;

@Component({
    selector: 'sub-accessories',
    templateUrl: 'sub-accessories.component.html',
    styleUrls: ['./sub-accessories.component.scss']
})
export class SubAccessoriesComponent implements OnInit {
    @Input()
    projectName: string;
    @Input()
    repositoryName: string;
    @Input()
    artifactDigest: string;
    @Input()
    accessories: Accessory[] = [];
    @Output()
    deleteAccessory: EventEmitter<Accessory> = new EventEmitter<Accessory>();
    currentPage: number = 1;
    @Input()
    total: number = 0;
    pageSize: number = ACCESSORY_PAGE_SIZE;
    page: number = 1;
    displayedAccessories: Accessory[] = [];
    loading: boolean = false;

    constructor(private activatedRoute: ActivatedRoute,
                private router: Router,
                private newArtifactService: NewArtifactService,
                private artifactService: ArtifactService,
                private errorHandlerService: ErrorHandler) {
    }

    ngOnInit(): void {
        this.displayedAccessories = clone(this.accessories);
    }

    size(size: number) {
        return formatSize(size.toString());
    }

    getIcon(icon: string): SafeUrl {
        return this.artifactService.getIcon(icon);
    }

    showDefaultIcon(event: any) {
        if (event && event.target) {
            event.target.src = artifactDefault;
        }
    }

    goIntoArtifactSummaryPage(accessory: Accessory): void {
        const relativeRouterLink: string[] = ['artifacts', accessory.digest];
        if (this.activatedRoute.snapshot.queryParams[UN_LOGGED_PARAM] === YES) {
            this.router.navigate(relativeRouterLink, {
                relativeTo: this.activatedRoute, queryParams: {
                    [UN_LOGGED_PARAM]: YES, [AccessoryQueryParams.ACCESSORY_TYPE]: accessory.type
                }
            });
        } else {
            this.router.navigate(relativeRouterLink, {
                relativeTo: this.activatedRoute, queryParams: {
                    [AccessoryQueryParams.ACCESSORY_TYPE]: accessory.type
                }
            });
        }
    }

    delete(a: Accessory) {
        this.deleteAccessory.emit(a);
    }

    clrLoad() {
        if (this.currentPage === 1) {
            this.displayedAccessories = clone(this.accessories);
            this.getIconFromBackend();
            return;
        }
        this.loading = true;
        const listTagParams: NewArtifactService.ListAccessoriesParams = {
            projectName: this.projectName,
            repositoryName: dbEncodeURIComponent(this.repositoryName),
            reference: this.artifactDigest,
            page: 1,
            pageSize: ACCESSORY_PAGE_SIZE
        };
        this.newArtifactService.listAccessories(listTagParams)
            .pipe(finalize(() => this.loading = false))
            .subscribe(
                res => {
                    this.displayedAccessories = res;
                    this.getIconFromBackend();
                },
                error => {
                    this.errorHandlerService.error(error);
                }
            );
    }
    getIconFromBackend() {
        if (this.displayedAccessories?.length) {
            this.artifactService.getIconsFromBackEnd(this.displayedAccessories);
        }
    }

    get dashLineHeight() {
        // fixed height 27 plus each row height 40
        return 27 + this.displayedAccessories?.length * 40;
    }
}
