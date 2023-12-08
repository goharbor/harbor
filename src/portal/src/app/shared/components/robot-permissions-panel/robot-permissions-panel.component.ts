import {
    AfterViewInit,
    Component,
    DoCheck,
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

@Component({
    selector: 'robot-permissions-panel',
    templateUrl: './robot-permissions-panel.component.html',
    styleUrls: ['./robot-permissions-panel.component.scss'],
})
export class RobotPermissionsPanelComponent implements OnChanges, DoCheck {
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

    @ViewChild('dropdownMenu')
    dropdownMenu: any;

    @ViewChild('dropdown')
    dropdown: ElementRef;

    // to avoid ng check error, getTransform() should always return 'unset' before dropdownMenu appears
    dropdownMenuAppeared: boolean = false;
    getTransform(): string {
        if (
            this.dropdownMenuAppeared &&
            this.dropdownMenu?.el &&
            this.dropdown
        ) {
            const width = this.dropdownMenu.el.nativeElement.offsetWidth;
            const height = this.dropdownMenu.el.nativeElement.offsetHeight;
            const bcr = this.dropdown.nativeElement.getBoundingClientRect();
            return `translateX(${bcr.x - width}px) translateY(${
                bcr.y - height / 2
            }px)`;
        }
        return 'unset';
    }

    ngOnChanges(changes: SimpleChanges) {
        if (changes && changes['candidatePermissions']) {
            this.initCandidates();
        }
    }
    ngDoCheck() {
        this.dropdownMenuAppeared = !!this.dropdownMenu;
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
