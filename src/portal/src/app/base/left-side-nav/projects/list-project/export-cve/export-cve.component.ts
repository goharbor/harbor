import { Component, ElementRef, ViewChild } from '@angular/core';
import { Label } from 'ng-swagger-gen/models/label';
import { LabelService } from 'ng-swagger-gen/services/label.service';
import { forkJoin, Observable } from 'rxjs';
import { finalize } from 'rxjs/operators';
import { Project } from 'src/app/base/project/project';
import { NgForm } from '@angular/forms';
import { ClrLoadingState } from '@clr/angular';
import { InlineAlertComponent } from '../../../../../shared/components/inline-alert/inline-alert.component';
import { ScanDataExportService } from '../../../../../../../ng-swagger-gen/services/scan-data-export.service';
import { MessageHandlerService } from '../../../../../shared/services/message-handler.service';
import {
    EventService,
    HarborEvent,
} from '../../../../../services/event-service/event.service';

const PAGE_SIZE: number = 100;
const SUPPORTED_MIME_TYPE: string =
    'application/vnd.security.vulnerability.report; version=1.1';
@Component({
    selector: 'export-cve',
    templateUrl: './export-cve.component.html',
    styleUrls: ['./export-cve.component.scss'],
})
export class ExportCveComponent {
    selectedProjects: Project[] = [];
    opened: boolean = false;
    loading: boolean = false;
    repos: string;
    tags: string;
    CVEIds: string;
    selectedLabels: Label[] = [];
    loadingAllLabels: boolean = false;
    allLabels: Label[] = [];
    @ViewChild('names', { static: true })
    namesSpan: ElementRef;
    @ViewChild('exportCVEForm', { static: true }) currentForm: NgForm;
    saveBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    @ViewChild(InlineAlertComponent)
    inlineAlertComponent: InlineAlertComponent;
    constructor(
        private labelService: LabelService,
        private scanDataExportService: ScanDataExportService,
        private msgHandler: MessageHandlerService,
        private event: EventService
    ) {}
    reset() {
        this.inlineAlertComponent?.close();
        this.selectedProjects = [];
        this.repos = null;
        this.tags = null;
        this.selectedLabels = [];
        this.CVEIds = null;
        this.currentForm?.reset();
        this.allLabels = [];
    }
    open(projects: Project[]) {
        this.reset();
        this.opened = true;
        this.selectedProjects = projects;
        this.getAllLabels();
    }

    close() {
        this.opened = false;
    }

    cancel() {
        this.close();
    }

    save() {
        this.loading = true;
        this.saveBtnState = ClrLoadingState.LOADING;
        const param: ScanDataExportService.ExportScanDataParams = {
            criteria: {
                projects: this.selectedProjects.map(item => item.project_id),
                labels: this.selectedLabels.map(item => item.id),
                repositories: this.handleBrace(this.repos),
                tags: this.handleBrace(this.tags),
                cveIds: this.handleBrace(this.CVEIds),
            },
            XScanDataType: SUPPORTED_MIME_TYPE,
        };
        this.scanDataExportService
            .exportScanData(param)
            .pipe(
                finalize(() => {
                    this.loading = false;
                    this.saveBtnState = ClrLoadingState.DEFAULT;
                })
            )
            .subscribe(
                res => {
                    this.msgHandler.showSuccess(
                        'CVE_EXPORT.TRIGGER_EXPORT_SUCCESS'
                    );
                    this.event.publish(HarborEvent.REFRESH_EXPORT_JOBS);
                    this.close();
                },
                err => {
                    this.inlineAlertComponent.showInlineError(err);
                }
            );
    }
    inputName() {}
    isSelected(l: Label): boolean {
        let flag: boolean = false;
        this.selectedLabels.forEach(item => {
            if (item.name === l.name) {
                flag = true;
            }
        });
        return flag;
    }
    selectOrUnselect(l: Label) {
        if (this.isSelected(l)) {
            this.selectedLabels = this.selectedLabels.filter(
                item => item.name !== l.name
            );
        } else {
            this.selectedLabels.push(l);
        }
    }

    getProjectNames(): string {
        if (this.selectedProjects?.length) {
            const names: string[] = [];
            this.selectedProjects.forEach(item => {
                names.push(item.name);
            });
            return names.join(', ');
        }
        return 'CVE_EXPORT.ALL_PROJECTS';
    }

    isOverflow(): boolean {
        return !(
            this.namesSpan?.nativeElement?.clientWidth >=
            this.namesSpan?.nativeElement?.scrollWidth
        );
    }
    getAllLabels(): void {
        // get all global labels
        this.loadingAllLabels = true;
        this.labelService
            .ListLabelsResponse({
                pageSize: PAGE_SIZE,
                page: 1,
                scope: 'g',
            })
            .pipe(finalize(() => (this.loadingAllLabels = false)))
            .subscribe(res => {
                if (res.headers) {
                    const xHeader: string = res.headers.get('X-Total-Count');
                    const totalCount = parseInt(xHeader, 0);
                    let arr = res.body || [];
                    if (totalCount <= 100) {
                        // already gotten all global labels
                        if (arr && arr.length) {
                            arr.forEach(data => {
                                this.allLabels.push(data);
                            });
                        }
                    } else {
                        // get all the global labels in specified times
                        const times: number = Math.ceil(totalCount / PAGE_SIZE);
                        const observableList: Observable<Label[]>[] = [];
                        for (let i = 2; i <= times; i++) {
                            observableList.push(
                                this.labelService.ListLabels({
                                    page: i,
                                    pageSize: PAGE_SIZE,
                                    scope: 'g',
                                })
                            );
                        }
                        this.loadingAllLabels = true;
                        forkJoin(observableList)
                            .pipe(
                                finalize(() => (this.loadingAllLabels = false))
                            )
                            .subscribe(response => {
                                if (response && response.length) {
                                    response.forEach(item => {
                                        arr = arr.concat(item);
                                    });
                                    arr.forEach(data => {
                                        this.allLabels.push(data);
                                    });
                                }
                            });
                    }
                }
            });
    }
    handleBrace(originStr: string): string {
        if (originStr) {
            if (
                originStr.indexOf(',') !== -1 &&
                originStr.indexOf('{') === -1 &&
                originStr.indexOf('}') === -1
            ) {
                return `{${originStr}}`;
            } else {
                return originStr;
            }
        }
        return null;
    }
}
