import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ViewTokenComponent } from './view-token.component';
import { RobotService } from '../../../../../ng-swagger-gen/services/robot.service';
import { OperationService } from '../operation/operation.service';
import { MessageHandlerService } from '../../services/message-handler.service';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { TranslateModule } from '@ngx-translate/core';
import { Robot } from '../../../../../ng-swagger-gen/models/robot';
import {
    Action,
    PermissionsKinds,
    Resource,
} from '../../../base/left-side-nav/system-robot-accounts/system-robot-util';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { SharedTestingModule } from '../../shared.module';

describe('ViewTokenComponent', () => {
    let component: ViewTokenComponent;
    let fixture: ComponentFixture<ViewTokenComponent>;
    const robot1: Robot = {
        id: 1,
        name: 'robot1',
        level: PermissionsKinds.SYSTEM,
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
    const fakedMessageHandlerService = {
        showSuccess() {},
        error() {},
    };
    const fakedRobotService = {
        UpdateRobot() {
            return of(null).pipe(delay(0));
        },
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [
                BrowserAnimationsModule,
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule,
                SharedTestingModule,
            ],
            declarations: [ViewTokenComponent],
            providers: [
                { provide: RobotService, useValue: fakedRobotService },
                OperationService,
                {
                    provide: MessageHandlerService,
                    useValue: fakedMessageHandlerService,
                },
            ],
            schemas: [NO_ERRORS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ViewTokenComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should show invalid secret', async () => {
        await fixture.whenStable();
        component.tokenModalOpened = true;
        component.robot = robot1;
        fixture.detectChanges();
        await fixture.whenStable();
        const newSecretInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#new-token');
        newSecretInput.value = '123';
        newSecretInput.dispatchEvent(new Event('input'));
        fixture.detectChanges();
        await fixture.whenStable();
        const error = fixture.nativeElement.querySelector('clr-control-error');
        expect(error).toBeTruthy();
    });
    it('should show secrets inconsistent', async () => {
        await fixture.whenStable();
        component.tokenModalOpened = true;
        component.robot = robot1;
        fixture.detectChanges();
        await fixture.whenStable();
        const newSecretInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#new-token');
        newSecretInput.value = 'Harbor12345';
        newSecretInput.dispatchEvent(new Event('input'));
        const confirmSecretInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#confirm-token');
        confirmSecretInput.value = 'Harbor123456';
        confirmSecretInput.dispatchEvent(new Event('input'));
        fixture.detectChanges();
        await fixture.whenStable();
        const error = fixture.nativeElement.querySelector('clr-control-error');
        expect(error).toBeTruthy();
    });
});
