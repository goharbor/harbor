import {
    AfterViewInit,
    Component,
    ElementRef,
    EventEmitter,
    Input,
    OnChanges,
    OnInit,
    Output,
    SimpleChanges,
    ViewChild,
} from '@angular/core';
import {
    convertKey,
    hasPermission,
    isCandidate,
} from '../../../base/left-side-nav/system-robot-accounts/system-robot-util';
import { Access } from '../../../../../ng-swagger-gen/models/access';
import { Permission } from '../../../../../ng-swagger-gen/models/permission';

enum Position {
    UP = 'left-bottom',
    DOWN = 'left-top',
}

@Component({
    selector: 'robot-permissions-panel',
    templateUrl: './robot-permissions-panel.component.html',
    styleUrls: ['./robot-permissions-panel.component.scss'],
})
export class RobotPermissionsPanelComponent
    implements AfterViewInit, OnChanges
{
    modalOpen: boolean = false;

    @Input()
    mode: PermissionSelectPanelModes = PermissionSelectPanelModes.NORMAL;

    @Input()
    dropdownPosition: string = 'bottom-left';

    @Input()
    usedInDatagrid: boolean = false;

    @Input()
    candidatePermissions: Permission[] = [];

    candidateActions: string[] = [];
    candidateResources: string[] = [];

    @Input()
    permissionsModel!: Access[];
    @Output()
    permissionsModelChange = new EventEmitter<Access[]>();

    @ViewChild('dropdown')
    clrDropdown: ElementRef;

    ngAfterViewInit() {
        setTimeout(() => {
            if (this.clrDropdown && this.usedInDatagrid) {
                if (
                    this.clrDropdown.nativeElement.getBoundingClientRect().y <
                    488
                ) {
                    this.dropdownPosition = Position.DOWN;
                } else {
                    this.dropdownPosition = Position.UP;
                }
            }
        });
    }

    ngOnChanges(changes: SimpleChanges) {
        if (changes && changes['candidatePermissions']) {
            this.initCandidates();
        }
    }

    initCandidates() {
        this.candidateActions = [];
        this.candidateResources = [];
        this.candidatePermissions?.forEach(item => {
            if (this.candidateResources.indexOf(item?.resource) === -1) {
                this.candidateResources.push(item?.resource);
            }
            if (this.candidateActions.indexOf(item?.action) === -1) {
                this.candidateActions.push(item?.action);
            }
        });
        this.candidateActions.sort();
        this.candidateResources.sort();
    }

    isCandidate(resource: string, action: string): boolean {
        return isCandidate(this.candidatePermissions, { resource, action });
    }

    getCheckBoxValue(resource: string, action: string): boolean {
        return hasPermission(this.permissionsModel, { resource, action });
    }

    setCheckBoxValue(resource: string, action: string, value: boolean) {
        if (value) {
            if (!this.permissionsModel) {
                this.permissionsModel = [];
            }
            this.permissionsModel.push({ resource, action });
        } else {
            this.permissionsModel = this.permissionsModel.filter(item => {
                return item.resource !== resource || item.action !== action;
            });
        }
        this.permissionsModelChange.emit(this.permissionsModel);
    }

    isAllSelected(): boolean {
        let flag: boolean = true;
        this.candidateActions.forEach(action => {
            this.candidateResources.forEach(resource => {
                if (
                    this.isCandidate(resource, action) &&
                    !hasPermission(this.permissionsModel, { resource, action })
                ) {
                    flag = false;
                }
            });
        });
        return flag;
    }

    selectAllOrUnselectAll() {
        if (this.isAllSelected()) {
            this.permissionsModel = [];
        } else {
            this.permissionsModel = [];
            this.candidateActions.forEach(action => {
                this.candidateResources.forEach(resource => {
                    if (this.isCandidate(resource, action)) {
                        this.permissionsModel.push({ resource, action });
                    }
                });
            });
        }
        this.permissionsModelChange.emit(this.permissionsModel);
    }

    convertKey(key: string): string {
        return convertKey(key);
    }
    protected readonly PermissionSelectPanelModes = PermissionSelectPanelModes;
}

export enum PermissionSelectPanelModes {
    DROPDOWN,
    MODAL,
    NORMAL,
}
