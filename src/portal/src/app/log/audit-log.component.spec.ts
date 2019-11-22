import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { AuditLogComponent } from './audit-log.component';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { AuditLogService } from './audit-log.service';
import { MessageHandlerService } from '../shared/message-handler/message-handler.service';
import { ActivatedRoute, Router } from '@angular/router';
import { of } from 'rxjs';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
import { RouterTestingModule } from '@angular/router/testing';
import { HarborLibraryModule } from '@harbor/ui';
import { delay } from 'rxjs/operators';

describe('AuditLogComponent', () => {
    let component: AuditLogComponent;
    let fixture: ComponentFixture<AuditLogComponent>;
    const mockMessageHandlerService = {
        handleError: () => {}
    };
    const mockAuditLogService = {
        listAuditLogs: () => {
            return of({
                headers: new Map().set('x-total-count', 0),
                body: []
            }).pipe(delay(0));
        },
    };
    const mockActivatedRoute = {
        data: of({
            auditLogResolver: ""
        }).pipe(delay(0)),
        snapshot: {
            parent: {
                params: {
                    id: 1
                }
            }
        }
    };
    const mockRouter = null;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            imports: [
                BrowserAnimationsModule,
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule,
                RouterTestingModule,
                NoopAnimationsModule,
                HarborLibraryModule
            ],
            declarations: [AuditLogComponent],
            providers: [
                TranslateService,
                { provide: ActivatedRoute, useValue: mockActivatedRoute },
                { provide: Router, useValue: mockRouter },
                { provide: AuditLogService, useValue: mockAuditLogService },
                { provide: MessageHandlerService, useValue: mockMessageHandlerService },

            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(AuditLogComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
