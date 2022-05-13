import { finalize } from 'rxjs/operators';
import {
    Component,
    OnInit,
    EventEmitter,
    Output,
    ViewChild,
} from '@angular/core';
import { NgForm } from '@angular/forms';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { SessionService } from '../../../../shared/services/session.service';
import { AppConfigService } from '../../../../services/app-config.service';
import { GroupType } from '../../../../shared/entities/shared.const';
import { UserGroup } from 'ng-swagger-gen/models/user-group';
import { UsergroupService } from '../../../../../../ng-swagger-gen/services/usergroup.service';

@Component({
    selector: 'hbr-add-group-modal',
    templateUrl: './add-group-modal.component.html',
    styleUrls: ['./add-group-modal.component.scss'],
})
export class AddGroupModalComponent implements OnInit {
    opened = false;
    mode = 'create';
    dnTooltip = 'TOOLTIP.ITEM_REQUIRED';

    group: UserGroup;
    @ViewChild('groupForm', { static: true })
    groupForm: NgForm;

    submitted = false;

    @Output() dataChange = new EventEmitter();

    isLdapMode: boolean;
    isHttpAuthMode: boolean;
    isOidcMode: boolean;
    constructor(
        private session: SessionService,
        private msgHandler: MessageHandlerService,
        private appConfigService: AppConfigService,
        private groupService: UsergroupService
    ) {}

    ngOnInit() {
        if (this.appConfigService.isLdapMode()) {
            this.isLdapMode = true;
        }
        if (this.appConfigService.isHttpAuthMode()) {
            this.isHttpAuthMode = true;
        }
        if (this.appConfigService.isOidcMode()) {
            this.isOidcMode = true;
        }
        this.group = {
            group_type: this.isLdapMode
                ? GroupType.LDAP_TYPE
                : this.isHttpAuthMode
                ? GroupType.HTTP_TYPE
                : GroupType.OIDC_TYPE,
        };
    }

    public get isFormValid(): boolean {
        return this.groupForm.valid;
    }

    public open(group?: UserGroup, editMode: boolean = false): void {
        this.resetGroup();
        if (editMode) {
            this.mode = 'edit';
            Object.assign(this.group, group);
        } else {
            this.mode = 'create';
        }
        this.opened = true;
    }

    public close(): void {
        this.opened = false;
        this.resetGroup();
    }

    save(): void {
        if (this.mode === 'create') {
            this.createGroup();
        } else {
            this.editGroup();
        }
    }

    createGroup() {
        let groupCopy = Object.assign({}, this.group);
        this.groupService
            .createUserGroup({
                usergroup: groupCopy,
            })
            .pipe(finalize(() => this.close()))
            .subscribe(
                res => {
                    this.msgHandler.showSuccess('GROUP.ADD_GROUP_SUCCESS');
                    this.dataChange.emit();
                },
                error => this.msgHandler.handleError(error)
            );
    }

    editGroup() {
        let groupCopy = Object.assign({}, this.group);
        this.groupService
            .updateUserGroup({
                groupId: groupCopy.id,
                usergroup: groupCopy,
            })
            .pipe(finalize(() => this.close()))
            .subscribe(
                res => {
                    this.msgHandler.showSuccess('GROUP.EDIT_GROUP_SUCCESS');
                    this.dataChange.emit();
                },
                error => this.msgHandler.handleError(error)
            );
    }

    resetGroup() {
        this.group = {
            group_type: this.isLdapMode
                ? GroupType.LDAP_TYPE
                : this.isHttpAuthMode
                ? GroupType.HTTP_TYPE
                : GroupType.OIDC_TYPE,
        };
        this.groupForm.reset();
    }
}
