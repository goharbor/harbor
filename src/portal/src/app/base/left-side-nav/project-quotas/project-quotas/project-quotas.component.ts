import {
    Component,
    Input,
    Output,
    EventEmitter,
    ViewChild,
    SimpleChanges,
    OnChanges,
} from '@angular/core';
import { Configuration } from '../../config/config';
import { State, QuotaHardLimitInterface } from '../../../../shared/services';
import {
    clone,
    isEmpty,
    getChanges,
    getSuitableUnit,
    calculatePage,
    getByte,
    GetIntegerAndUnit,
    getPageSizeFromLocalStorage,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../../shared/units/utils';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import {
    QuotaUnits,
    QuotaUnlimited,
    QUOTA_DANGER_COEFFICIENT,
    QUOTA_WARNING_COEFFICIENT,
} from '../../../../shared/entities/shared.const';
import { EditProjectQuotasComponent } from './edit-project-quotas/edit-project-quotas.component';
import { TranslateService } from '@ngx-translate/core';
import { forkJoin, of } from 'rxjs';
import { Router } from '@angular/router';
import { finalize, mergeMap } from 'rxjs/operators';
import { ClrDatagridStateInterface } from '@clr/angular';
import { ConfigurationService } from '../../../../services/config.service';
import { QuotaService } from '../../../../../../ng-swagger-gen/services/quota.service';
import { QuotaUpdateReq } from '../../../../../../ng-swagger-gen/models/quota-update-req';
import { ProjectService } from '../../../../../../ng-swagger-gen/services/project.service';
import { Quota } from '../../../../../../ng-swagger-gen/models/quota';
import { FilterComponent } from '../../../../shared/components/filter/filter.component';

const QuotaType = 'project';

@Component({
    selector: 'project-quotas',
    templateUrl: './project-quotas.component.html',
    styleUrls: ['./project-quotas.component.scss'],
})
export class ProjectQuotasComponent implements OnChanges {
    config: Configuration = new Configuration();
    @ViewChild('editProjectQuotas')
    editQuotaDialog: EditProjectQuotasComponent;
    loading = true;
    quotaHardLimitValue: QuotaHardLimitInterface;
    currentState: State;

    @Output() configChange: EventEmitter<Configuration> =
        new EventEmitter<Configuration>();
    @Output() refreshAllconfig: EventEmitter<Configuration> =
        new EventEmitter<Configuration>();
    quotaList: Quota[] = [];
    originalConfig: Configuration;
    currentPage = 1;
    totalCount = 0;
    pageSize = getPageSizeFromLocalStorage(
        PageSizeMapKeys.PROJECT_QUOTA_COMPONENT
    );
    quotaDangerCoefficient: number = QUOTA_DANGER_COEFFICIENT;
    quotaWarningCoefficient: number = QUOTA_WARNING_COEFFICIENT;
    @Input()
    get allConfig(): Configuration {
        return this.config;
    }
    set allConfig(cfg: Configuration) {
        this.config = cfg;
        this.configChange.emit(this.config);
    }
    selectedRow: Quota[] = [];
    @ViewChild(FilterComponent)
    filterComponent: FilterComponent;
    constructor(
        private configService: ConfigurationService,
        private quotaService: QuotaService,
        private translate: TranslateService,
        private router: Router,
        private errorHandler: ErrorHandler,
        private projectService: ProjectService
    ) {}

    editQuota() {
        if (this.selectedRow && this.selectedRow.length === 1) {
            const defaultTexts = [
                this.translate.get('QUOTA.EDIT_PROJECT_QUOTAS'),
                this.translate.get('QUOTA.SET_QUOTAS', {
                    params: this.selectedRow[0].ref.name,
                }),
                this.translate.get('QUOTA.STORAGE_QUOTA'),
            ];
            forkJoin(...defaultTexts).subscribe(res => {
                const defaultTextsObj = {
                    editQuota: res[0],
                    setQuota: res[1],
                    storageQuota: res[2],
                    quotaHardLimitValue: this.selectedRow[0],
                    isSystemDefaultQuota: false,
                };
                this.editQuotaDialog.openEditQuotaModal(defaultTextsObj);
            });
        }
    }

    editDefaultQuota(quotaHardLimitValue: QuotaHardLimitInterface) {
        const defaultTexts = [
            this.translate.get('QUOTA.EDIT_DEFAULT_PROJECT_QUOTAS'),
            this.translate.get('QUOTA.SET_DEFAULT_QUOTAS'),
            this.translate.get('QUOTA.STORAGE_DEFAULT_QUOTA'),
        ];
        forkJoin(...defaultTexts).subscribe(res => {
            const defaultTextsObj = {
                editQuota: res[0],
                setQuota: res[1],
                storageQuota: res[2],
                quotaHardLimitValue: quotaHardLimitValue,
                isSystemDefaultQuota: true,
            };
            this.editQuotaDialog.openEditQuotaModal(defaultTextsObj);
        });
    }
    public getChanges() {
        let allChanges = getChanges(this.originalConfig, this.config);
        if (allChanges) {
            return this.getQuotaChanges(allChanges);
        }
        return null;
    }

    getQuotaChanges(allChanges) {
        let changes = {};
        for (let prop in allChanges) {
            if (prop === 'storage_per_project') {
                changes[prop] = allChanges[prop];
            }
        }
        return changes;
    }

    public saveConfig(configQuota): void {
        this.allConfig.storage_per_project.value =
            +configQuota.storage === QuotaUnlimited
                ? configQuota.storage
                : getByte(configQuota.storage, configQuota.storageUnit);
        let changes = this.getChanges();
        if (!isEmpty(changes)) {
            this.loading = true;
            this.configService
                .saveConfiguration(changes)
                .pipe(
                    finalize(() => {
                        this.loading = false;
                        this.editQuotaDialog.openEditQuota = false;
                    })
                )
                .subscribe(
                    response => {
                        this.refreshAllconfig.emit();
                        this.errorHandler.info('CONFIG.SAVE_SUCCESS');
                    },
                    error => {
                        this.errorHandler.error(error);
                    }
                );
        } else {
            // Inprop situation, should not come here
            this.translate.get('CONFIG.NO_CHANGE').subscribe(res => {
                this.editQuotaDialog.inlineAlert.showInlineError(res);
            });
        }
    }

    confirmEdit(event) {
        if (event.isSystemDefaultQuota) {
            this.saveConfig(event.formValue);
        } else {
            this.saveCurrentQuota(event);
        }
    }
    saveCurrentQuota(event) {
        let storage =
            +event.formValue.storage === QuotaUnlimited
                ? +event.formValue.storage
                : getByte(
                      +event.formValue.storage,
                      event.formValue.storageUnit
                  );
        let rep: QuotaUpdateReq = { hard: { storage } };
        this.loading = true;
        this.quotaService.updateQuota({ id: event.id, hard: rep }).subscribe(
            res => {
                this.editQuotaDialog.openEditQuota = false;
                this.getQuotaList(this.currentState);
                this.errorHandler.info('QUOTA.SAVE_SUCCESS');
            },
            error => {
                this.editQuotaDialog.inlineAlert.showInlineError(error);
                this.loading = false;
            }
        );
    }

    getquotaHardLimitValue() {
        const storageNumberAndUnit = this.allConfig.storage_per_project
            ? this.allConfig.storage_per_project.value
            : QuotaUnlimited;
        const storageLimit = storageNumberAndUnit;
        const storageUnit = this.getIntegerAndUnit(
            storageNumberAndUnit,
            0
        ).partCharacterHard;
        this.quotaHardLimitValue = { storageLimit, storageUnit };
    }
    getQuotaList(state: ClrDatagridStateInterface) {
        if (!state || !state.page) {
            return;
        }
        this.pageSize = state.page.size;
        setPageSizeToLocalStorage(
            PageSizeMapKeys.PROJECT_QUOTA_COMPONENT,
            this.pageSize
        );
        // Keep state for future filtering and sorting
        this.currentState = state;

        let pageNumber: number = calculatePage(state);
        if (pageNumber <= 0) {
            pageNumber = 1;
        }
        this.loading = true;
        this.quotaService
            .listQuotasResponse({
                reference: QuotaType,
                page: pageNumber,
                pageSize: this.pageSize,
            })
            .pipe(
                finalize(() => {
                    this.loading = false;
                    this.selectedRow = [];
                })
            )
            .subscribe(
                res => {
                    if (res.headers) {
                        let xHeader: string = res.headers.get('X-Total-Count');
                        if (xHeader) {
                            this.totalCount = parseInt(xHeader, 0);
                        }
                    }
                    this.quotaList = res.body.filter(quota => {
                        return quota.ref !== null;
                    }) as Quota[];
                },
                error => {
                    this.errorHandler.error(error);
                }
            );
    }
    ngOnChanges(changes: SimpleChanges): void {
        if (changes && changes['allConfig']) {
            this.originalConfig = clone(this.config);
            this.getquotaHardLimitValue();
        }
    }
    getSuitableUnit(value) {
        const QuotaUnitsCopy = clone(QuotaUnits);
        return getSuitableUnit(value, QuotaUnitsCopy);
    }
    getIntegerAndUnit(valueHard, valueUsed) {
        return GetIntegerAndUnit(
            valueHard,
            clone(QuotaUnits),
            valueUsed,
            clone(QuotaUnits)
        );
    }

    goToLink(proId) {
        let linkUrl = ['harbor', 'projects', proId];
        this.router.navigate(linkUrl);
    }
    refresh() {
        if (this.filterComponent) {
            this.filterComponent.currentValue = null;
        }
        this.currentPage = 1;
        this.selectedRow = [];
        this.getQuotaList(this.currentState);
    }
    doSearch(name: string) {
        if (name) {
            // should query project by name first, then query quota by referenceId(project_id)
            this.projectService
                .listProjects({
                    withDetail: false,
                    q: encodeURIComponent(`name=${name}`),
                })
                .pipe(
                    mergeMap(projects => {
                        if (projects && projects.length) {
                            return this.quotaService.listQuotas({
                                referenceId: projects[0].project_id.toString(),
                            });
                        }
                        return of([]);
                    })
                )
                .subscribe(res => {
                    this.quotaList = res;
                });
        } else {
            this.refresh();
        }
    }
}
