import { ComponentFixture, TestBed } from '@angular/core/testing';
import { GroupComponent } from './group.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { SessionService } from '../../../shared/services/session.service';
import { of } from 'rxjs';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { AppConfigService } from '../../../services/app-config.service';
import { OperationService } from '../../../shared/components/operation/operation.service';
import { ConfirmationDialogService } from '../../global-confirmation-dialog/confirmation-dialog.service';
import { UsergroupService } from '../../../../../ng-swagger-gen/services/usergroup.service';
import { SharedTestingModule } from '../../../shared/shared.module';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { delay } from 'rxjs/operators';
import { UserGroup } from '../../../../../ng-swagger-gen/models/user-group';

describe('GroupComponent', () => {
    let component: GroupComponent;
    let fixture: ComponentFixture<GroupComponent>;
    let fakeMessageHandlerService = null;
    let fakeOperationService = null;
    let fakeGroupService = {
        listUserGroupsResponse: function () {
            const res: HttpResponse<Array<UserGroup>> = new HttpResponse<
                Array<UserGroup>
            >({
                headers: new HttpHeaders({ 'x-total-count': '3' }),
                body: [
                    {
                        group_name: '',
                    },
                    {
                        group_name: 'abc',
                    },
                ],
            });
            return of(res).pipe(delay(0));
        },
    };
    let fakeConfirmationDialogService = {
        confirmationConfirm$: of({
            state: 1,
            source: 2,
        }),
    };
    let fakeSessionService = {
        currentUser: {
            has_admin_role: true,
        },
    };
    let fakeAppConfigService = {
        isLdapMode: function () {
            return true;
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [GroupComponent],
            imports: [SharedTestingModule],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            providers: [
                {
                    provide: MessageHandlerService,
                    useValue: fakeMessageHandlerService,
                },
                { provide: OperationService, useValue: fakeOperationService },
                { provide: UsergroupService, useValue: fakeGroupService },
                {
                    provide: ConfirmationDialogService,
                    useValue: fakeConfirmationDialogService,
                },
                { provide: SessionService, useValue: fakeSessionService },
                { provide: AppConfigService, useValue: fakeAppConfigService },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(GroupComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
