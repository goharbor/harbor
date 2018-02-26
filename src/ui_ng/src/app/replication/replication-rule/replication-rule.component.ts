import {Component, OnInit, OnDestroy, ViewChild, ChangeDetectorRef, AfterViewInit} from '@angular/core';
import {ProjectService} from '../../project/project.service';
import {Project} from '../../project/project';
import {ActivatedRoute, Router} from '@angular/router';
import {FormArray, FormBuilder, FormGroup, Validators} from "@angular/forms";
import {ReplicationRuleServie} from "./replication-rule.service";
import {MessageHandlerService} from "../../shared/message-handler/message-handler.service";
import {Target, Filter, ReplicationRule} from "./replication-rule";
import {ConfirmationDialogService} from "../../shared/confirmation-dialog/confirmation-dialog.service";
import { ConfirmationTargets, ConfirmationState } from '../../shared/shared.const';
import {Subscription} from "rxjs/Subscription";
import {ConfirmationMessage} from "../../shared/confirmation-dialog/confirmation-message";
import {Subject} from "rxjs/Subject";
import {ListProjectModelComponent} from "./list-project-model/list-project-model.component";
import {toPromise, isEmptyObject, compareValue} from "harbor-ui/src/utils";
import {CreateEditEndpointComponent} from "harbor-ui/src/create-edit-endpoint/create-edit-endpoint.component";

const ONE_HOUR_SECONDS: number = 3600;
const ONE_DAY_SECONDS: number = 24 * ONE_HOUR_SECONDS;

@Component ({
    selector: 'repliction-rule',
    templateUrl: 'replication-rule.html',
    styleUrls: ['replication-rule.css']

})

export class ReplicationRuleComponent implements OnInit, OnDestroy {
    _localTime: Date = new Date();
    policyId: number;
    projectId: number;
    targetList: Target[] = [];
    isFilterHide: boolean = false;
    weeklySchedule: boolean;
    isScheduleOpt: boolean;
    isImmediate: boolean = false;
    noProjectInfo: string = "";
    noSelectedProject: boolean = true;
    noSelectedEndpoint: boolean = true;
    filterCount: number = 0;
    selectedprojectList: Project[] = [];
    triggerNames: string[] = ['Manual', 'Immediate', 'Scheduled'];
    scheduleNames: string[] = ['Daily', 'Weekly'];
    weekly: string[] = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'];
    filterSelect: string[] = ['repository', 'tag'];
    ruleNameTooltip: string = 'TOOLTIP.EMPTY';
    headerTitle: string = 'REPLICATION.ADD_POLICY';

    filterListData: {[key: string]: any}[] = [];
    inProgress: boolean = false;
    inNameChecking: boolean = false;
    isRuleNameExist: boolean = false;
    isSubmitOver: boolean = false;
    nameChecker: Subject<string> = new Subject<string>();

    confirmSub: Subscription;
    ruleForm: FormGroup;
    copyUpdateForm: ReplicationRule;
    emptyEndpoint = new Target();

    @ViewChild(ListProjectModelComponent)
    projectListModel: ListProjectModelComponent;

    @ViewChild(CreateEditEndpointComponent)
    createEditEndpointComponent: CreateEditEndpointComponent;

    baseFilterData(name: string, option: string[], state: boolean) {
        return {
            name: name,
            options: option,
            state: state,
            isValid: true
        };
    }

    constructor(public projectService: ProjectService,
                private router: Router,
                private fb: FormBuilder,
                private repService: ReplicationRuleServie,
                private route: ActivatedRoute,
                private msgHandler: MessageHandlerService,
                private confirmService: ConfirmationDialogService,
                public ref: ChangeDetectorRef) {
        this.createForm();
        Promise.all([this.repService.getEndpoints(), this.repService.listProjects()])
            .then(res => {
                if (!res[0]) {
                    this.noSelectedEndpoint = true;
                }else {
                    this.targetList = res[0];
                    if (!this.policyId) {
                        res[0].unshift(this.emptyEndpoint);
                        this.setTarget([res[0][0]]);
                    }
                }
                if (!res[1]) {
                    this.noProjectInfo = 'REPLICATION.NO_PROJECT_INFO';
                }else {
                    if (!this.policyId && !this.projectId) {
                        this.setProject([res[1][0]]);
                    }
                    if (!this.policyId && this.projectId) {
                        this.setProject( res[1].filter(rule => rule.project_id === this.projectId));
                        this.noSelectedProject = false;
                    }
                }
                if (!this.policyId) {
                    this.copyUpdateForm = Object.assign({}, this.ruleForm.value);
                }
            });
    }

    ngOnInit(): void {
       this.policyId = +this.route.snapshot.params['id'];
       this.projectId = +this.route.snapshot.params['projectId'];
       if (this.policyId) {
           this.headerTitle = 'REPLICATION.EDIT_POLICY_TITLE';
           this.repService.getReplicationRule(this.policyId)
               .then((response) => {
                    this.copyUpdateForm = Object.assign({}, response);
                    // set filter value is [] if callback fiter value is null.
                   this.copyUpdateForm.filters = response.filters ? response.filters : [];
                    this.updateForm(response);
               }).catch(error => {
               this.msgHandler.handleError(error);
           });
       }

       this.nameChecker.debounceTime(500).distinctUntilChanged().subscribe((ruleName: string) => {
           this.isRuleNameExist = false;
           this.inNameChecking = true;
           toPromise<ReplicationRule[]>(this.repService.getReplicationRules(0, ruleName))
               .then(response => {
                   if (response.some(rule => rule.name === ruleName)) {
                       this.ruleNameTooltip = 'TOOLTIP.RULE_USER_EXISTING';
                       this.isRuleNameExist = true;
                   }
                   this.inNameChecking = false;
               }).catch(() => {
               this.inNameChecking = false;
           });
       });
    }

    ngOnDestroy(): void {
        if (this.confirmSub) {
            this.confirmSub.unsubscribe();
        }
        if (this.nameChecker) {
            this.nameChecker.unsubscribe();
        }
    }

    get isVaild() {
        return !(this.isRuleNameExist || this.noSelectedProject || this.noSelectedEndpoint || this.inProgress || this.isSubmitOver);
    }

    createForm() {
        this.ruleForm = this.fb.group({
            name: ['', Validators.required],
            description: '',
            projects: this.fb.array([]),
            targets: this.fb.array([]),
            trigger: this.fb.group({
                kind: this.triggerNames[0],
                schedule_param: this.fb.group({
                    type: this.scheduleNames[0],
                    weekday: 1,
                    offtime: '08:00'
                }),
            }),
            filters: this.fb.array([]),
            replicate_existing_image_now: true,
            replicate_deletion: false
        });
    }

    updateForm(rule: ReplicationRule): void {
        rule.trigger = this.updateTrigger(rule.trigger);
        this.ruleForm.reset({
            name: rule.name,
            description: rule.description,
            trigger: rule.trigger,
            replicate_existing_image_now: rule.replicate_existing_image_now,
            replicate_deletion: rule.replicate_deletion
        });
        this.setProject(rule.projects);
        this.noSelectedProject = false;
        this.setTarget(rule.targets);
        this.noSelectedEndpoint = false;

        if (rule.filters) {
            this.setFilter(rule.filters);
            this.updateFilter(rule.filters);
        }

        // Force refresh view
        let hnd = setInterval(() => this.ref.markForCheck(), 100);
        setTimeout(() => clearInterval(hnd), 2000);
    }

    get projects(): FormArray {
        return this.ruleForm.get('projects') as FormArray;
    }
    setProject(projects: Project[]) {
        const projectFGs = projects.map(project => this.fb.group(project));
        const projectFormArray = this.fb.array(projectFGs);
        this.ruleForm.setControl('projects', projectFormArray);
    }

    get filters(): FormArray {
        return this.ruleForm.get('filters') as FormArray;
    }
    setFilter(filters: Filter[]) {
        const filterFGs = filters.map(filter => this.fb.group(filter));
        const filterFormArray = this.fb.array(filterFGs);
        this.ruleForm.setControl('filters', filterFormArray);
    }

    get targets(): FormArray {
        return this.ruleForm.get('targets') as FormArray;
    }
    setTarget(targets: Target[]) {
        const targetFGs = targets.map(target => this.fb.group(target));
        const targetFormArray = this.fb.array(targetFGs);
        this.ruleForm.setControl('targets', targetFormArray);
    }

    initFilter(name: string) {
        return this.fb.group({
            kind: name,
            pattern: ['', Validators.required]
        });
    }

    filterChange($event: any) {
        if ($event && $event.target['value']) {
            let id: number = $event.target.id;
            let name: string = $event.target.name;
            let value: string = $event.target['value'];

            this.filterListData.forEach((data, index) => {
                if (index === +id) {
                    data.name = $event.target.name = value;
                }else {
                    data.options.splice(data.options.indexOf(value), 1);
                }
                if (data.options.indexOf(name) === -1) {
                    data.options.push(name);
                }
            });
        }
    }

    targetChange($event: any) {
        if ($event && $event.target && event.target['value']) {
            if ($event.target['value'] === '-1') {
                this.noSelectedEndpoint = true;
                return;
            }
            let selecedTarget: Target = this.targetList.find(target => target.id === +$event.target['value']);
            this.setTarget([selecedTarget]);
            this.noSelectedEndpoint = false;
        }
    }

    openProjectModel(): void {
        this.projectListModel.openModel();
    }

    selectedProject(project: Project): void {
        if (!project) {
            this.noSelectedProject = true;
        }else {
            this.noSelectedProject = false;
            this.setProject([project]);
        }
    }

    addNewFilter(): void {
        if (this.filterCount === 0) {
            this.filterListData.push(this.baseFilterData(this.filterSelect[0], this.filterSelect.slice(), true));
            this.filters.push(this.initFilter(this.filterSelect[0]));

        }else {
            let nameArr: string[] = this.filterSelect.slice();
            this.filterListData.forEach(data => {
                nameArr.splice(nameArr.indexOf(data.name), 1);
            });
            // when add a new filter,the filterListData should change the options
            this.filterListData.filter((data) => {
                data.options.splice(data.options.indexOf(nameArr[0]), 1);
            });
            this.filterListData.push(this.baseFilterData(nameArr[0], nameArr, true));
            this.filters.push(this.initFilter(nameArr[0]));
        }
        this.filterCount += 1;
        if (this.filterCount >= this.filterSelect.length) {
            this.isFilterHide = true;
        }
    }

    // delete a filter
    deleteFilter(i: number): void {
        if (i || i === 0) {
            let delfilter = this.filterListData.splice(i, 1)[0];
            if (this.filterCount === this.filterSelect.length) {
                this.isFilterHide = false;
            }
            this.filterCount -= 1;
            if (this.filterListData.length) {
                let optionVal = delfilter.name;
                this.filterListData.filter(data => {
                    if (data.options.indexOf(optionVal) === -1) {
                        data.options.push(optionVal);
                    }
                });
            }
            const control = <FormArray>this.ruleForm.controls['filters'];
            control.removeAt(i);
        }
    }

    selectTrigger($event: any): void {
        if ($event && $event.target && $event.target['value']) {
            let val: string = $event.target['value'];
            if (val === this.triggerNames[2]) {
                this.isScheduleOpt = true;
                this.isImmediate = false;
            }
            if (val === this.triggerNames[1]) {
                this.isScheduleOpt = false;
                this.isImmediate = true;
            }
            if (val === this.triggerNames[0]) {
                this.isScheduleOpt = false;
                this.isImmediate = false;
            }
        }
    }

    // Replication Schedule select value exchange
    selectSchedule($event: any): void {
        if ($event && $event.target && $event.target['value']) {
            switch ($event.target['value']) {
                case this.scheduleNames[1]:
                    this.weeklySchedule = true;
                    this.ruleForm.patchValue({
                        trigger: {
                            schedule_param: {
                                weekday: 1,
                            }
                        }
                    })
                    break;
                case this.scheduleNames[0]:
                    this.weeklySchedule = false;
                    break;
            }
        }
    }

    checkRuleName(): void {
        let ruleName: string = this.ruleForm.controls['name'].value;
        if (ruleName) {
            this.nameChecker.next(ruleName);
        } else {
            this.ruleNameTooltip = 'TOOLTIP.EMPTY';
        }
    }

    updateFilter(filters: any) {
        let opt: string[] = this.filterSelect.slice();
        filters.forEach((filter: any) => {
            opt.splice(opt.indexOf(filter.kind), 1);
        })
        filters.forEach((filter: any) => {
            let option: string [] = opt.slice();
            option.unshift(filter.kind);
            this.filterListData.push(this.baseFilterData(filter.kind, option, true));
        });
        this.filterCount = filters.length;
        if (filters.length === this.filterSelect.length) {
            this.isFilterHide = true;
        }
    }

    updateTrigger(trigger: any) {
        if (trigger['schedule_param']) {
            this.isScheduleOpt = true;
            this.isImmediate = false;
            trigger['schedule_param']['offtime'] = this.getOfftime(trigger['schedule_param']['offtime']);
            if (trigger['schedule_param']['weekday']) {
                this.weeklySchedule = true;
            }else {
                // set default
                trigger['schedule_param']['weekday'] = 1;
            }
        }else {
            if (trigger['kind'] === this.triggerNames[0]) {
                this.isImmediate = false;
            }
            trigger['schedule_param'] = { type: this.scheduleNames[0],
                weekday: this.weekly[0],
                offtime: '08:00'};
        }
        return trigger;
    }

    setTriggerVaule(trigger: any) {
        if (!this.isScheduleOpt) {
            delete trigger['schedule_param'];
            return trigger;
        }else {
            if (!this.weeklySchedule) {
                delete trigger['schedule_param']['weekday'];
            }else {
                trigger['schedule_param']['weekday'] = +trigger['schedule_param']['weekday'];
            }
            trigger['schedule_param']['offtime'] = this.setOfftime(trigger['schedule_param']['offtime']);
            return trigger;
        }
    }

    public hasFormChange(): boolean {
        return !isEmptyObject(this.getChanges());
    }

    onSubmit() {
        // add new Replication rule
        this.inProgress = true;
        let copyRuleForm: ReplicationRule = this.ruleForm.value;
        copyRuleForm.trigger = this.setTriggerVaule(copyRuleForm.trigger);
        if (!this.policyId) {
            this.repService.createReplicationRule(copyRuleForm)
                .then(() => {
                    this.msgHandler.showSuccess('REPLICATION.CREATED_SUCCESS');
                    this.inProgress = false;
                    this.isSubmitOver = true;
                    setTimeout(() => {
                        this.copyUpdateForm = Object.assign({}, this.ruleForm.value);
                        if (this.projectId) {
                            this.router.navigate(['harbor/projects', this.projectId, 'replications']);
                        }else {
                            this.router.navigate(['/harbor/replications']);
                        }
                    }, 2000);

                }).catch((error: any) => {
                this.inProgress = false;
                this.msgHandler.handleError(error);
            });
        } else {
            this.repService.updateReplicationRule(this.policyId, this.ruleForm.value)
                .then(() => {
                    this.msgHandler.showSuccess('REPLICATION.UPDATED_SUCCESS');
                    this.inProgress = false;
                    this.isSubmitOver = true;
                    setTimeout(() => {
                        this.copyUpdateForm = Object.assign({}, this.ruleForm.value);
                        if (this.projectId) {
                            this.router.navigate(['harbor/projects', this.projectId, 'replications']);
                        }else {
                            this.router.navigate(['/harbor/replications']);
                        }
                    }, 2000);

                }).catch((error: any) => {
                this.inProgress = false;
                this.msgHandler.handleError(error);
            });
        }
    }

    openModal() {
        this.createEditEndpointComponent.openCreateEditTarget(true);
    }

    reload($event: boolean) {
        if ($event) {
            Promise.all([this.repService.getEndpoints()]).then(res => {
                this.targetList = res[0];
                this.setTarget([this.targetList[this.targetList.length - 1]]);
                this.noSelectedEndpoint = false;
            });
        }
    }

    onCancel(): void {
        this.router.navigate(['/harbor/replications']);
    }

    // UTC time
    public getOfftime(daily_time: any): string {

        let timeOffset: number = 0; // seconds
        if (daily_time && typeof daily_time === 'number') {
            timeOffset = +daily_time;
        }

        // Convert to current time
        let timezoneOffset: number = this._localTime.getTimezoneOffset();
        // Local time
        timeOffset = timeOffset - timezoneOffset * 60;
        if (timeOffset < 0) {
            timeOffset = timeOffset + ONE_DAY_SECONDS;
        }

        if (timeOffset >= ONE_DAY_SECONDS) {
            timeOffset -= ONE_DAY_SECONDS;
        }

        // To time string
        let hours: number = Math.floor(timeOffset / ONE_HOUR_SECONDS);
        let minutes: number = Math.floor((timeOffset - hours * ONE_HOUR_SECONDS) / 60);

        let timeStr: string = '' + hours;
        if (hours < 10) {
            timeStr = '0' + timeStr;
        }
        if (minutes < 10) {
            timeStr += ':0';
        } else {
            timeStr += ':';
        }
        timeStr += minutes;

        return timeStr;
    }
    public setOfftime(v: string) {
        if (!v || v === '') {
            return;
        }

        let values: string[] = v.split(':');
        if (!values || values.length !== 2) {
            return;
        }

        let hours: number = +values[0];
        let minutes: number = +values[1];
        // Convert to UTC time
        let timezoneOffset: number = this._localTime.getTimezoneOffset();
        let utcTimes: number = hours * ONE_HOUR_SECONDS + minutes * 60;
        utcTimes += timezoneOffset * 60;
        if (utcTimes < 0) {
            utcTimes += ONE_DAY_SECONDS;
        }

        if (utcTimes >= ONE_DAY_SECONDS) {
            utcTimes -= ONE_DAY_SECONDS;
        }

        return utcTimes;
    }

    backReplication(): void {
        this.router.navigate(['/harbor/replications']);
    }
    backProjectReplication(): void {
        this.router.navigate(['harbor/projects', this.projectId, 'replications']);
    }


    getChanges(): { [key: string]: any | any[] } {
        let changes: { [key: string]: any | any[] } = {};
        let ruleValue: { [key: string]: any | any[] } = this.ruleForm.value;
        if (!ruleValue || !this.copyUpdateForm) {
            return changes;
        }
        for (let prop in ruleValue) {
            let field = this.copyUpdateForm[prop];
            if (!compareValue(field, ruleValue[prop])) {
                changes[prop] = ruleValue[prop];
                //Number
                if (typeof field === "number") {
                    changes[prop] = +changes[prop];
                }

                //Trim string value
                if (typeof field === "string") {
                    changes[prop] = ('' + changes[prop]).trim();
                }
            }
        }

        return changes;
    }

}
