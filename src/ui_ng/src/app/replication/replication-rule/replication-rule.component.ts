import {Component, OnInit, OnDestroy, ViewChild, ChangeDetectorRef, AfterViewInit} from '@angular/core';
import {ProjectService} from '../../project/project.service';
import {Project} from '../../project/project';
import {compareValue, toPromise} from 'harbor-ui/src/utils';
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


const ONE_HOUR_SECONDS: number = 3600;
const ONE_DAY_SECONDS: number = 24 * ONE_HOUR_SECONDS;

@Component ({
    selector: 'repliction-rule',
    templateUrl: 'replication-rule.html',
    styleUrls: ['replication-rule.css']

})
export class ReplicationRuleComponent implements OnInit, AfterViewInit, OnDestroy {
    timerHandler: any;
    _localTime: Date = new Date();
    policyId: number;
    projectList: Project[] = [];
    targetList: Target[] = [];
    isFilterHide: boolean = false;
    weeklySchedule: boolean;
    isScheduleOpt: boolean;
    filterCount: number = 0;
    triggerNames: string[] = ['immediate', 'schedule', 'manual'];
    scheduleNames: string[] = ['daily', 'weekly'];
    weekly: string[] = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'];
    filterSelect: string[] = ['repository', 'tag'];
    ruleNameTooltip: string = 'TOOLTIP.EMPTY';

    filterListData: {[key: string]: any}[] = [];
    inProgress: boolean = false;
    inNameChecking: boolean = false;
    isBackReplication: boolean = false;
    isRuleNameExist: boolean = false;
    nameChecker: Subject<string> = new Subject<string>();

    confirmSub: Subscription;
    ruleForm: FormGroup;
    copyUpdateForm: ReplicationRule;

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
                if (!res[0].length || !res[1].length) {
                    this.msgHandler.error('should have project and target first');
                    this.router.navigate(['/harbor/replications']);
                };
                if (res[0].length && res[1].length) {
                    this.projectList = res[1];
                    this.setProject([this.projectList[0]]);
                    this.targetList = res[0];
                    this.setTarget([this.targetList[0]]);
                }
            });
    }

    ngOnInit(): void {
       this.policyId = +this.route.snapshot.params['id'];
       if (this.policyId) {
           this.repService.getReplicationRule(this.policyId)
               .then((response) => {
                    this.copyUpdateForm = Object.assign({}, response);
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
        this.confirmSub = this.confirmService.confirmationConfirm$.subscribe(confirmation => {
            if (confirmation &&
                confirmation.state === ConfirmationState.CONFIRMED) {
                if (confirmation.source === ConfirmationTargets.CONFIG) {
                    if (this.policyId) {
                        this.updateForm(this.copyUpdateForm);
                    } else {
                        this.initFom();
                    }
                    if (this.isBackReplication) {
                        this.router.navigate(['/harbor/replications']);
                    }
                }
            }
        });
    }

    get hasFormChange() {
        if (this.copyUpdateForm) {
          return  !compareValue(this.copyUpdateForm, this.ruleForm.value);
        }
        return this.ruleForm.touched && this.ruleForm.dirty;
    }

    ngAfterViewInit(): void {
    }

    ngOnDestroy(): void {
        if (this.confirmSub) {
            this.confirmSub.unsubscribe();
        }
        if (this.nameChecker) {
            this.nameChecker.unsubscribe();
        }
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
            replicate_deletion: true
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
        this.setTarget(rule.targets);
        if (rule.filters) {
            this.setFilter(rule.filters);
            this.updateFilter(rule.filters);
        }
    }

    initFom(): void {
        this.ruleForm.reset({
            name: '',
            description: '',
            trigger: {kind: this.triggerNames[0], schedule_param: {
                type: this.scheduleNames[0],
                weekday: 1,
                offtime: '08:00'
            }},
            replicate_existing_image_now: true,
            replicate_deletion: true
        });
        this.setProject([this.projectList[0]]);
        this.setTarget([this.targetList[0]]);
        this.setFilter([]);

        this.isFilterHide = false;
        this.filterListData = [];
        this.isScheduleOpt = false;
        this.weeklySchedule = false;
        this.isRuleNameExist = true;
        this.ruleNameTooltip = 'TOOLTIP.EMPTY';
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

    projectChange($event: any) {
        if ($event && $event.target && event.target['value']) {
            let selecedProject: Project = this.projectList.find(project => project.project_id === +$event.target['value']);
            this.setProject([selecedProject]);
        }
    }

    targetChange($event: any) {
        if ($event && $event.target && event.target['value']) {
            let selecedTarget: Target = this.targetList.find(target => target.id === +$event.target['value']);
            this.setTarget([selecedTarget]);
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
            if ($event.target['value'] === this.triggerNames[1]) {
                this.isScheduleOpt = true;
            } else {
                this.isScheduleOpt = false;
            }
        }
    }

    // Replication Schedule select exchange
    selectSchedule($event: any): void {
        if ($event && $event.target && $event.target['value']) {
            switch ($event.target['value']) {
                case this.scheduleNames[1]:
                    this.weeklySchedule = true;
                    break;
                case this.scheduleNames[0]:
/*                    this.dailySchedule = true;*/
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
            trigger['schedule_param']['offtime'] = this.getOfftime(trigger['schedule_param']['offtime']);
            if (trigger['schedule_param']['weekday']) {
                this.weeklySchedule = true;
            }else {
                // set default
                trigger['schedule_param']['weekday'] = 1;
            }
        }else {
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


    onSubmit() {
        // add new Replication rule
        let copyRuleForm: ReplicationRule = this.ruleForm.value;
        copyRuleForm.trigger = this.setTriggerVaule(copyRuleForm.trigger);
        if (!this.policyId) {
            this.repService.createReplicationRule(copyRuleForm)
                .then(() => {
                this.msgHandler.showSuccess('REPLICATION.CREATED_SUCCESS');
                this.inProgress = false;
                setTimeout(() => {
                    this.router.navigate(['/harbor/replications']);
                }, 2000);

            }).catch((error: any) => {
                this.inProgress = false;
                this.msgHandler.handleError(error);
            });
        } else {
            this.repService.updateReplicationRule(this.policyId, this.ruleForm.value)
                .then(() => {
                this.msgHandler.showSuccess('REPLICATION.CREATED_SUCCESS');
                this.inProgress = false;
                setTimeout(() => {
                    this.router.navigate(['/harbor/replications']);
                }, 2000);

            }).catch((error: any) => {
                this.inProgress = false;
                this.msgHandler.handleError(error);
            });
        }
        this.inProgress = true;
    }

    onCancel(): void {

        console.log(this.ruleForm.valid, this.isRuleNameExist , !this.hasFormChange)
        if (this.ruleForm.dirty) {
            let msg = new ConfirmationMessage(
                'CONFIG.CONFIRM_TITLE',
                'CONFIG.CONFIRM_SUMMARY',
                '',
                null,
                ConfirmationTargets.CONFIG
            );

            this.confirmService.openComfirmDialog(msg);
        }
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
        this.isBackReplication = true;
        if (this.ruleForm.dirty) {
            this.onCancel();
        } else {
            this.router.navigate(['/harbor/replications']);
        }
    }
}
