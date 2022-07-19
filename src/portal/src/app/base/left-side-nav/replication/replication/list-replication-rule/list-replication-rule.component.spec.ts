import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ConfirmationDialogComponent } from '../../../../../shared/components/confirmation-dialog';
import { ListReplicationRuleComponent } from './list-replication-rule.component';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import { OperationService } from '../../../../../shared/components/operation/operation.service';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { SharedTestingModule } from '../../../../../shared/shared.module';
import { ReplicationPolicy } from '../../../../../../../ng-swagger-gen/models/replication-policy';
import { ReplicationService } from 'ng-swagger-gen/services/replication.service';

describe('ListReplicationRuleComponent (inline template)', () => {
    let mockRules: ReplicationPolicy[] = [
        {
            id: 1,
            name: 'sync_01',
            description: '',
            filters: null,
            trigger: { type: 'Manual', trigger_settings: null },
            deletion: false,
            dest_namespace: '',
            src_registry: { id: 3 },
            enabled: true,
            override: true,
            speed: -1,
        },
        {
            id: 2,
            name: 'sync_02',
            description: '',
            filters: null,
            trigger: { type: 'Manual', trigger_settings: null },
            deletion: false,
            dest_namespace: '',
            dest_registry: { id: 3 },
            enabled: true,
            override: true,
            speed: -1,
        },
    ];

    let fixture: ComponentFixture<ListReplicationRuleComponent>;

    let comp: ListReplicationRuleComponent;
    const fakedReplicationService = {
        updateReplicationPolicy() {
            return of(true).pipe(delay(0));
        },
        deleteReplicationPolicy() {
            return of(true).pipe(delay(0));
        },
        listReplicationPoliciesResponse() {
            return of(
                new HttpResponse({
                    body: mockRules,
                    headers: new HttpHeaders({
                        'x-total-count': '2',
                    }),
                })
            ).pipe(delay(0));
        },
    };
    const fakedOperationService = {
        publishInfo() {
            return undefined;
        },
    };
    const fakedErrorHandler = {
        info() {
            return undefined;
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [
                ListReplicationRuleComponent,
                ConfirmationDialogComponent,
            ],
            providers: [
                { provide: ErrorHandler, useValue: fakedErrorHandler },
                {
                    provide: ReplicationService,
                    useValue: fakedReplicationService,
                },
                { provide: OperationService, useValue: fakedOperationService },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ListReplicationRuleComponent);
        comp = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('Should load and render data', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        fixture.detectChanges();
        const el = fixture.nativeElement.querySelector('clr-dg-cell');
        expect(el).toBeTruthy();
        fixture.detectChanges();
        expect(el.textContent.trim()).toEqual('sync_01');
    });
    it('should disable rule', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        comp.selectedRow = comp.rules[0];
        comp.selectedRow.enabled = true;
        fixture.detectChanges();
        await fixture.whenStable();
        const action: HTMLElement =
            fixture.nativeElement.querySelector('#rule-action');
        action.click();
        fixture.detectChanges();
        await fixture.whenStable();
        const disable: HTMLElement =
            fixture.nativeElement.querySelector('#rule-disable');
        disable.click();
        fixture.detectChanges();
        await fixture.whenStable();
        const button: HTMLElement = fixture.nativeElement.querySelector(
            '#dialog-action-disable'
        );
        button.click();
        fixture.detectChanges();
        await fixture.whenStable();
        const body: HTMLElement =
            fixture.nativeElement.querySelector('.modal-body');
        expect(body).toBeFalsy();
    });
    it('should enable rule', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        comp.selectedRow = comp.rules[0];
        comp.selectedRow.enabled = false;
        fixture.detectChanges();
        await fixture.whenStable();
        const action: HTMLElement =
            fixture.nativeElement.querySelector('#rule-action');
        action.click();
        fixture.detectChanges();
        await fixture.whenStable();
        const enable: HTMLElement =
            fixture.nativeElement.querySelector('#rule-enable');
        enable.click();
        fixture.detectChanges();
        await fixture.whenStable();
        const button: HTMLElement = fixture.nativeElement.querySelector(
            '#dialog-action-enable'
        );
        button.click();
        fixture.detectChanges();
        await fixture.whenStable();
        const body: HTMLElement =
            fixture.nativeElement.querySelector('.modal-body');
        expect(body).toBeFalsy();
    });
});
