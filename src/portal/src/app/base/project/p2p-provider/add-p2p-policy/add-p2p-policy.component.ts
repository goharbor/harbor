import {
    Component,
    EventEmitter,
    Input,
    OnDestroy,
    OnInit,
    Output,
    ViewChild,
} from '@angular/core';
import { PreheatPolicy } from '../../../../../../ng-swagger-gen/models/preheat-policy';
import { NgForm } from '@angular/forms';
import { OriginCron, ProjectService } from '../../../../shared/services';
import { CronScheduleComponent } from '../../../../shared/components/cron-schedule';
import { PreheatService } from '../../../../../../ng-swagger-gen/services/preheat.service';
import {
    debounceTime,
    distinctUntilChanged,
    filter,
    finalize,
    switchMap,
} from 'rxjs/operators';
import { deleteEmptyKey } from '../../../../shared/units/utils';
import { ClrLoadingState } from '@clr/angular';
import { SessionService } from '../../../../shared/services/session.service';
import { Project } from '../../project';
import { ActivatedRoute } from '@angular/router';
import {
    FILTER_TYPE,
    PROJECT_SEVERITY_LEVEL_MAP,
    TRIGGER,
    TRIGGER_I18N_MAP,
} from '../p2p-provider.service';
import { ProviderUnderProject } from '../../../../../../ng-swagger-gen/models/provider-under-project';
import { AppConfigService } from '../../../../services/app-config.service';
import { Subject, Subscription } from 'rxjs';
import { InlineAlertComponent } from '../../../../shared/components/inline-alert/inline-alert.component';

const SCHEDULE_TYPE = {
    NONE: 'None',
    DAILY: 'Daily',
    WEEKLY: 'Weekly',
    HOURLY: 'Hourly',
    CUSTOM: 'Custom',
};
const TRUE: string = 'true';
@Component({
    selector: 'add-p2p-policy',
    templateUrl: './add-p2p-policy.component.html',
    styleUrls: ['./add-p2p-policy.component.scss'],
})
export class AddP2pPolicyComponent implements OnInit, OnDestroy {
    severityOptions = [
        { severity: 5, severityLevel: 'VULNERABILITY.SEVERITY.CRITICAL' },
        { severity: 4, severityLevel: 'VULNERABILITY.SEVERITY.HIGH' },
        { severity: 3, severityLevel: 'VULNERABILITY.SEVERITY.MEDIUM' },
        { severity: 2, severityLevel: 'VULNERABILITY.SEVERITY.LOW' },
        { severity: 0, severityLevel: 'VULNERABILITY.SEVERITY.NONE' },
    ];
    isEdit: boolean;
    isOpen: boolean = false;
    closable: boolean = false;
    staticBackdrop: boolean = true;
    projectName: string;
    projectId: number;
    @Output() notify = new EventEmitter<boolean>();

    @ViewChild(InlineAlertComponent)
    inlineAlert: InlineAlertComponent;
    policy: PreheatPolicy = {};
    repos: string;
    tags: string;
    onlySignedImages: boolean = false;
    severity: number;
    labels: string;
    triggerType: string = TRIGGER.MANUAL;
    cron: string;
    @ViewChild('policyForm', { static: true }) currentForm: NgForm;
    loading: boolean = false;
    @ViewChild('cronScheduleComponent')
    cronScheduleComponent: CronScheduleComponent;
    buttonStatus: ClrLoadingState = ClrLoadingState.DEFAULT;
    originPolicyForEdit: PreheatPolicy;
    originReposForEdit: string;
    originTagsForEdit: string;
    originOnlySignedImagesForEdit: boolean;
    originSeverityForEdit: number;
    originLabelsForEdit: string;
    originTriggerTypeForEdit: string;
    originCronForEdit: string;
    @Input()
    providers: ProviderUnderProject[] = [];
    preventVul: boolean = false;
    projectSeverity: string;
    triggers: string[] = [
        TRIGGER.MANUAL,
        TRIGGER.SCHEDULED,
        TRIGGER.EVENT_BASED,
    ];
    enableContentTrust: boolean = false;
    private _nameSubject: Subject<string> = new Subject<string>();
    private _nameSubscription: Subscription;
    isNameExisting: boolean = false;
    checkNameOnGoing: boolean = false;
    @Output()
    hasInit: EventEmitter<boolean> = new EventEmitter<boolean>();
    constructor(
        private preheatService: PreheatService,
        private session: SessionService,
        private route: ActivatedRoute,
        private appConfigService: AppConfigService,
        private projectService: ProjectService
    ) {}

    ngOnInit() {
        const resolverData = this.route.snapshot.parent.parent.data;
        if (resolverData) {
            const project = <Project>resolverData['projectResolver'];
            this.projectName = project.name;
            this.projectId = project.project_id;
            // get latest project info
            this.getProject();
        }
        this.subscribeName();
    }
    ngOnDestroy() {
        if (this._nameSubscription) {
            this._nameSubscription.unsubscribe();
            this._nameSubscription = null;
        }
    }
    subscribeName() {
        if (!this._nameSubscription) {
            this._nameSubscription = this._nameSubject
                .pipe(
                    debounceTime(500),
                    distinctUntilChanged(),
                    filter(name => {
                        if (
                            this.isEdit &&
                            this.originPolicyForEdit &&
                            this.originPolicyForEdit.name === name
                        ) {
                            return false;
                        }
                        return name.length > 0;
                    }),
                    switchMap(name => {
                        this.isNameExisting = false;
                        this.checkNameOnGoing = true;
                        return this.preheatService
                            .ListPolicies({
                                projectName: this.projectName,
                                q: encodeURIComponent(`name=${name}`),
                            })
                            .pipe(
                                finalize(() => (this.checkNameOnGoing = false))
                            );
                    })
                )
                .subscribe(res => {
                    if (res && res.length > 0) {
                        this.isNameExisting = true;
                    }
                });
        }
    }
    inputName() {
        this._nameSubject.next(this.policy.name);
    }
    getProject() {
        this.projectService.getProject(this.projectId).subscribe(project => {
            if (project && project.metadata) {
                this.preventVul = project.metadata.prevent_vul === TRUE;
                this.projectSeverity = project.metadata.severity;
                this.enableContentTrust =
                    project.metadata.enable_content_trust === TRUE;
                this.severity =
                    PROJECT_SEVERITY_LEVEL_MAP[this.projectSeverity];
            }
            this.hasInit.emit(true);
        });
    }

    resetForAdd() {
        this.inlineAlert.close();
        this.policy = {};
        this.repos = null;
        this.tags = null;
        this.labels = null;
        this.cron = null;
        this.isNameExisting = false;
        if (this.providers && this.providers.length) {
            this.providers.forEach(item => {
                if (item.default) {
                    this.policy.provider_id = item.id;
                }
            });
        }
        this.currentForm.reset({
            triggerType: 'manual',
            severity: PROJECT_SEVERITY_LEVEL_MAP[this.projectSeverity],
            onlySignedImages: this.enableContentTrust,
            provider: this.policy.provider_id,
        });
    }

    setCron(event: any) {
        this.cron = event;
        this.cronScheduleComponent.resetSchedule();
    }

    getCron(): OriginCron {
        const originCron: OriginCron = {
            type: SCHEDULE_TYPE.NONE,
            cron: '',
        };
        originCron.cron = this.cron;
        if (
            originCron.cron === '' ||
            originCron.cron === null ||
            originCron.cron === undefined
        ) {
            originCron.type = SCHEDULE_TYPE.NONE;
        } else if (originCron.cron === '0 0 * * * *') {
            originCron.type = SCHEDULE_TYPE.HOURLY;
        } else if (originCron.cron === '0 0 0 * * *') {
            originCron.type = SCHEDULE_TYPE.DAILY;
        } else if (originCron.cron === '0 0 0 * * 0') {
            originCron.type = SCHEDULE_TYPE.WEEKLY;
        } else {
            originCron.type = SCHEDULE_TYPE.CUSTOM;
        }
        return originCron;
    }

    onCancel() {
        this.isOpen = false;
    }

    closeModal() {
        this.isOpen = false;
    }

    addOrSave(isAdd: boolean) {
        const policy: PreheatPolicy = {};
        Object.assign(policy, this.policy);
        policy.provider_id = +policy.provider_id;
        const filters: any[] = [];
        if (this.repos) {
            if (
                this.repos.indexOf(',') !== -1 &&
                this.repos.indexOf('{') === -1 &&
                this.repos.indexOf('}') === -1
            ) {
                filters.push({
                    type: FILTER_TYPE.REPOS,
                    value: `{${this.repos}}`,
                });
            } else {
                filters.push({ type: FILTER_TYPE.REPOS, value: this.repos });
            }
        }
        if (this.tags) {
            if (
                this.tags.indexOf(',') !== -1 &&
                this.tags.indexOf('{') === -1 &&
                this.tags.indexOf('}') === -1
            ) {
                filters.push({
                    type: FILTER_TYPE.TAG,
                    value: `{${this.tags}}`,
                });
            } else {
                filters.push({ type: FILTER_TYPE.TAG, value: this.tags });
            }
        }
        if (this.labels) {
            if (
                this.labels.indexOf(',') !== -1 &&
                this.labels.indexOf('{') === -1 &&
                this.labels.indexOf('}') === -1
            ) {
                filters.push({
                    type: FILTER_TYPE.LABEL,
                    value: `{${this.labels}}`,
                });
            } else {
                filters.push({ type: FILTER_TYPE.LABEL, value: this.labels });
            }
        }
        policy.filters = JSON.stringify(filters);
        const trigger: any = {
            type: this.triggerType ? this.triggerType : TRIGGER.MANUAL,
            trigger_setting: {
                cron:
                    !this.triggerType ||
                    this.triggerType === TRIGGER.MANUAL ||
                    this.triggerType === TRIGGER.EVENT_BASED
                        ? ''
                        : this.cron,
            },
        };
        policy.trigger = JSON.stringify(trigger);
        this.loading = true;
        this.buttonStatus = ClrLoadingState.LOADING;
        deleteEmptyKey(policy);
        if (isAdd) {
            policy.project_id = this.projectId;
            policy.enabled = true;
            this.preheatService
                .CreatePolicy({ projectName: this.projectName, policy: policy })
                .pipe(finalize(() => (this.loading = false)))
                .subscribe(
                    response => {
                        this.buttonStatus = ClrLoadingState.SUCCESS;
                        this.closeModal();
                        this.notify.emit(isAdd);
                    },
                    error => {
                        this.inlineAlert.showInlineError(error);
                        this.buttonStatus = ClrLoadingState.ERROR;
                    }
                );
        } else {
            policy.id = this.originPolicyForEdit.id;
            this.preheatService
                .UpdatePolicy({
                    projectName: this.projectName,
                    preheatPolicyName: this.originPolicyForEdit.name,
                    policy: policy,
                })
                .pipe(finalize(() => (this.loading = false)))
                .subscribe(
                    response => {
                        this.buttonStatus = ClrLoadingState.SUCCESS;
                        this.closeModal();
                        this.notify.emit(isAdd);
                    },
                    error => {
                        this.inlineAlert.showInlineError(error);
                        this.buttonStatus = ClrLoadingState.ERROR;
                    }
                );
        }
    }

    valid(): boolean {
        if (this.triggerType === TRIGGER.SCHEDULED && !this.cron) {
            return false;
        }
        return this.currentForm.valid;
    }

    compare(): boolean {
        if (this.projectSeverity && this.preventVul) {
            if (
                PROJECT_SEVERITY_LEVEL_MAP[this.projectSeverity] >
                (this.severity ? this.severity : 0)
            ) {
                return true;
            }
        }
        return false;
    }

    hasChange(): boolean {
        // eslint-disable-next-line eqeqeq
        if (this.policy.provider_id != this.originPolicyForEdit.provider_id) {
            return true;
        }
        // eslint-disable-next-line eqeqeq
        if (this.policy.name != this.originPolicyForEdit.name) {
            return true;
        }
        if (
            (this.policy.description || this.originPolicyForEdit.description) &&
            // eslint-disable-next-line eqeqeq
            this.policy.description != this.originPolicyForEdit.description
        ) {
            return true;
        }
        // eslint-disable-next-line eqeqeq
        if (this.originReposForEdit != this.repos) {
            return true;
        }
        // eslint-disable-next-line eqeqeq
        if (this.originTagsForEdit != this.tags) {
            return true;
        }
        // eslint-disable-next-line eqeqeq
        if (this.originOnlySignedImagesForEdit != this.onlySignedImages) {
            return true;
        }
        // eslint-disable-next-line eqeqeq
        if (this.originLabelsForEdit != this.labels) {
            return true;
        }
        // eslint-disable-next-line eqeqeq
        if (this.originSeverityForEdit != this.severity) {
            return true;
        }
        // eslint-disable-next-line eqeqeq
        if (this.originTriggerTypeForEdit != this.triggerType) {
            return true;
        }
        // eslint-disable-next-line eqeqeq
        return this.originCronForEdit != this.cron;
    }
    isSystemAdmin(): boolean {
        const account = this.session.getCurrentUser();
        return account != null && account.has_admin_role;
    }

    getTriggerTypeI18n(triggerType): string {
        if (triggerType) {
            return TRIGGER_I18N_MAP[triggerType];
        }
        return '';
    }
    showCron(): boolean {
        if (this.triggerType) {
            return this.triggerType === TRIGGER.SCHEDULED;
        }
        return false;
    }
    withNotary(): boolean {
        return this.appConfigService.getConfig().with_notary;
    }
    showExplainForEventBased(): boolean {
        return this.triggerType === TRIGGER.EVENT_BASED;
    }
}
