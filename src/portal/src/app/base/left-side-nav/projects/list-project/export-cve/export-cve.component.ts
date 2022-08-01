import {
    Component,
    ElementRef,
    EventEmitter,
    Output,
    ViewChild,
} from '@angular/core';
import { Label } from 'ng-swagger-gen/models/label';
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
import { LabelService } from 'src/app/shared/services/label.service';

const SUPPORTED_MIME_TYPE: string =
    'application/vnd.security.vulnerability.report; version=1.1';
@Component({
    selector: 'export-cve',
    templateUrl: './export-cve.component.html',
    styleUrls: ['./export-cve.component.scss'],
})
export class ExportCveComponent {
    @Output() triggerExportSuccess = new EventEmitter<void>();
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
                    this.triggerExportSuccess.emit();
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
            if (item.id === l.id) {
                flag = true;
            }
        });
        return flag;
    }
    selectOrUnselect(l: Label) {
        if (this.isSelected(l)) {
            this.selectedLabels = this.selectedLabels.filter(
                item => item.id !== l.id
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
            .getAllGlobalAndSpecificProjectLabels(
                this.selectedProjects[0].project_id
            )
            .pipe(finalize(() => (this.loadingAllLabels = false)))
            .subscribe(res => {
                this.allLabels = res;
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
