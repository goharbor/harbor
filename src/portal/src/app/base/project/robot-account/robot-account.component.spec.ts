import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { of, Subscription } from 'rxjs';
import { ActivatedRoute, RouterModule } from '@angular/router';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { RobotAccountComponent } from './robot-account.component';
import { UserPermissionService } from '../../../shared/services';
import { OperationService } from '../../../shared/components/operation/operation.service';
import { RobotService } from '../../../../../ng-swagger-gen/services/robot.service';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { Robot } from '../../../../../ng-swagger-gen/models/robot';
import { delay } from 'rxjs/operators';
import {
    Action,
    PermissionsKinds,
    Resource,
} from '../../left-side-nav/system-robot-accounts/system-robot-util';
import { ConfirmationDialogService } from '../../global-confirmation-dialog/confirmation-dialog.service';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CommonModule } from '@angular/common';
import { ClarityModule } from '@clr/angular';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { HarborDatetimePipe } from '../../../shared/pipes/harbor-datetime.pipe';

describe('RobotAccountComponent', () => {
    let component: RobotAccountComponent;
    let fixture: ComponentFixture<RobotAccountComponent>;
    const robot1: Robot = {
        id: 1,
        name: 'robot1',
        level: PermissionsKinds.PROJECT,
        disable: false,
        expires_at: (new Date().getTime() + 100000) % 1000,
        description: 'for test',
        secret: 'tthf54hfth4545dfgd5g454grd54gd54g',
        permissions: [
            {
                kind: PermissionsKinds.PROJECT,
                namespace: 'project1',
                access: [
                    {
                        resource: Resource.ARTIFACT,
                        action: Action.PUSH,
                    },
                ],
            },
        ],
    };
    const robot2: Robot = {
        id: 2,
        name: 'robot2',
        level: PermissionsKinds.PROJECT,
        disable: false,
        expires_at: (new Date().getTime() + 100000) % 1000,
        description: 'for test',
        secret: 'fsdf454654654fs6dfe',
        permissions: [
            {
                kind: PermissionsKinds.PROJECT,
                namespace: 'project2',
                access: [
                    {
                        resource: Resource.ARTIFACT,
                        action: Action.PUSH,
                    },
                ],
            },
        ],
    };
    const robot3: Robot = {
        id: 3,
        name: 'robot3',
        level: PermissionsKinds.PROJECT,
        disable: false,
        expires_at: (new Date().getTime() + 100000) % 1000,
        description: 'for test',
        secret: 'fsdg48454fse84',
        permissions: [
            {
                kind: PermissionsKinds.PROJECT,
                namespace: 'project3',
                access: [
                    {
                        resource: Resource.ARTIFACT,
                        action: Action.PUSH,
                    },
                ],
            },
        ],
    };
    const mockUserPermissionService = {
        getPermission() {
            return of(true);
        },
    };
    const fakedRobotService = {
        ListRobotResponse() {
            const res: HttpResponse<Array<Robot>> = new HttpResponse<
                Array<Robot>
            >({
                headers: new HttpHeaders({ 'x-total-count': '3' }),
                body: [robot1, robot2, robot3],
            });
            return of(res).pipe(delay(0));
        },
    };
    const fakedMessageHandlerService = {
        showSuccess() {},
        error() {},
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            schemas: [NO_ERRORS_SCHEMA],
            imports: [
                TranslateModule.forRoot(),
                CommonModule,
                ClarityModule,
                HttpClientTestingModule,
                RouterTestingModule,
                BrowserAnimationsModule,
            ],
            providers: [
                TranslateService,
                {
                    provide: ActivatedRoute,
                    useValue: {
                        snapshot: {
                            parent: {
                                parent: {
                                    params: { id: 1 },
                                    data: null,
                                },
                            },
                        },
                    },
                },
                {
                    provide: MessageHandlerService,
                    useValue: fakedMessageHandlerService,
                },
                ConfirmationDialogService,
                OperationService,
                {
                    provide: UserPermissionService,
                    useValue: mockUserPermissionService,
                },
                { provide: RobotService, useValue: fakedRobotService },
            ],
            declarations: [RobotAccountComponent, HarborDatetimePipe],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(RobotAccountComponent);
        component = fixture.componentInstance;
        component.searchSub = new Subscription();
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should render project robot list', async () => {
        fixture.autoDetectChanges();
        await fixture.whenStable();
        const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(rows.length).toEqual(3);
    });
});
