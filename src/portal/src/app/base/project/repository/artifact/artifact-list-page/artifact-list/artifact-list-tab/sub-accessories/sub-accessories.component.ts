import {
    AfterViewInit,
    ChangeDetectionStrategy,
    ChangeDetectorRef,
    Component,
    Input,
    OnDestroy,
    OnInit,
} from '@angular/core';
import {
    clone,
    dbEncodeURIComponent,
    formatSize,
} from '../../../../../../../../shared/units/utils';
import { ActivatedRoute, Router } from '@angular/router';
import { Accessory } from 'ng-swagger-gen/models/accessory';
import { ArtifactService as NewArtifactService } from 'ng-swagger-gen/services/artifact.service';
import { ErrorHandler } from '../../../../../../../../shared/units/error-handler';
import { finalize } from 'rxjs/operators';
import { SafeUrl } from '@angular/platform-browser';
import { ArtifactService } from '../../../../artifact.service';
import {
    AccessoryFront,
    AccessoryQueryParams,
    artifactDefault,
} from '../../../../artifact';
import {
    EventService,
    HarborEvent,
} from '../../../../../../../../services/event-service/event.service';
import { Subscription } from 'rxjs';

export const ACCESSORY_PAGE_SIZE: number = 5;

@Component({
    selector: 'sub-accessories',
    templateUrl: 'sub-accessories.component.html',
    styleUrls: ['./sub-accessories.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush, // use OnPush Strategy to avoid ExpressionChangedAfterItHasBeenCheckedError
})
export class SubAccessoriesComponent
    implements OnInit, AfterViewInit, OnDestroy
{
    @Input()
    projectName: string;
    @Input()
    repositoryName: string;
    @Input()
    artifactDigest: string;
    @Input()
    accessories: Accessory[] = [];
    currentPage: number = 1;
    @Input()
    total: number = 0;
    pageSize: number = ACCESSORY_PAGE_SIZE;
    page: number = 1;
    displayedAccessories: AccessoryFront[] = [];
    loading: boolean = false;
    iconSub: Subscription;
    constructor(
        private activatedRoute: ActivatedRoute,
        private router: Router,
        private newArtifactService: NewArtifactService,
        private artifactService: ArtifactService,
        private errorHandlerService: ErrorHandler,
        private cdf: ChangeDetectorRef,
        private event: EventService
    ) {}

    ngAfterViewInit(): void {
        this.cdf.detectChanges();
    }

    ngOnInit(): void {
        if (!this.iconSub) {
            this.iconSub = this.event.subscribe(
                HarborEvent.RETRIEVED_ICON,
                () => {
                    this.cdf.detectChanges();
                }
            );
        }
        this.displayedAccessories = clone(this.accessories);
    }

    ngOnDestroy() {
        if (this.iconSub) {
            this.iconSub.unsubscribe();
            this.iconSub = null;
        }
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
        this.router.navigate(relativeRouterLink, {
            relativeTo: this.activatedRoute,
            queryParams: {
                [AccessoryQueryParams.ACCESSORY_TYPE]: accessory.type,
            },
        });
    }

    delete(a: Accessory) {
        this.event.publish(HarborEvent.DELETE_ACCESSORY, a);
    }

    clrLoad() {
        if (this.currentPage === 1) {
            this.displayedAccessories = clone(this.accessories);
            this.getIconFromBackend();
            this.getAccessoriesAsync(this.displayedAccessories);
            return;
        }
        this.loading = true;
        const listTagParams: NewArtifactService.ListAccessoriesParams = {
            projectName: this.projectName,
            repositoryName: dbEncodeURIComponent(this.repositoryName),
            reference: this.artifactDigest,
            page: this.currentPage,
            pageSize: ACCESSORY_PAGE_SIZE,
        };
        this.newArtifactService
            .listAccessories(listTagParams)
            .pipe(finalize(() => (this.loading = false)))
            .subscribe(
                res => {
                    this.displayedAccessories = res;
                    this.cdf.detectChanges();
                    this.getIconFromBackend();
                    this.getAccessoriesAsync(this.displayedAccessories);
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

    // get accessories
    getAccessoriesAsync(artifacts: AccessoryFront[]) {
        if (artifacts && artifacts.length) {
            artifacts.forEach(item => {
                const listTagParams: NewArtifactService.ListAccessoriesParams =
                    {
                        projectName: this.projectName,
                        repositoryName: dbEncodeURIComponent(
                            this.repositoryName
                        ),
                        reference: item.digest,
                        page: 1,
                        pageSize: ACCESSORY_PAGE_SIZE,
                    };
                this.newArtifactService
                    .listAccessoriesResponse(listTagParams)
                    .subscribe(res => {
                        if (res.headers) {
                            let xHeader: string =
                                res.headers.get('x-total-count');
                            if (xHeader) {
                                item.accessoryNumber = Number.parseInt(
                                    xHeader,
                                    10
                                );
                            }
                        }
                        item.accessories = res.body;
                        this.cdf.detectChanges();
                    });
            });
        }
    }

    copyDigest(a: Accessory) {
        this.event.publish(HarborEvent.COPY_DIGEST, a);
    }
}
