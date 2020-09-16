import { ComponentFixture, TestBed, async } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { NoopAnimationsModule } from "@angular/platform-browser/animations";

import { DebugElement } from '@angular/core';

import { SharedModule } from '../../utils/shared/shared.module';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';

import { ListReplicationRuleComponent } from './list-replication-rule.component';
import { ReplicationRule } from '../../services/interface';

import { ErrorHandler } from '../../utils/error-handler/error-handler';
import { SERVICE_CONFIG, IServiceConfig } from '../../entities/service.config';
import { ReplicationService } from '../../services/replication.service';
import { OperationService } from "../operation/operation.service";
import { of } from 'rxjs';
import { CURRENT_BASE_HREF } from "../../utils/utils";
import { delay } from "rxjs/operators";

describe('ListReplicationRuleComponent (inline template)', () => {

    let mockRules: ReplicationRule[] = [
        {
            "id": 1,
            "name": "sync_01",
            "description": "",
            "filters": null,
            "trigger": {"type": "Manual", "trigger_settings": null},
            "error_job_count": 2,
            "deletion": false,
            "src_namespaces": ["name1", "name2"],
            "src_registry": {id: 3},
            "enabled": true,
            "override": true
        },
        {
            "id": 2,
            "name": "sync_02",
            "description": "",
            "filters": null,
            "trigger": {"type": "Manual", "trigger_settings": null},
            "error_job_count": 2,
            "deletion": false,
            "src_namespaces": ["name1", "name2"],
            "dest_registry": {id: 3},
            "enabled": true,
            "override": true
        },
    ];

    let fixture: ComponentFixture<ListReplicationRuleComponent>;

    let comp: ListReplicationRuleComponent;

    let replicationService: ReplicationService;

    let spyRules: jasmine.Spy;

    let config: IServiceConfig = {
        replicationRuleEndpoint: CURRENT_BASE_HREF + '/policies/replication/testing'
    };
    const fakedReplicationService = {
        getReplicationRules() {
            return of(mockRules).pipe(delay(0));
        },
        updateReplicationRule() {
            return of(true).pipe(delay(0));
        }
    };
    const fakedOperationService = {
        publishInfo() {
           return undefined;
        }
    };
    const fakedErrorHandler = {
        info() {
            return undefined;
        }
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                SharedModule,
                NoopAnimationsModule
            ],
            declarations: [
                ListReplicationRuleComponent,
                ConfirmationDialogComponent
            ],
            providers: [
                {provide: ErrorHandler, useValue: fakedErrorHandler},
                {provide: SERVICE_CONFIG, useValue: config},
                {provide: ReplicationService, useValue: fakedReplicationService},
                {provide: OperationService, useValue: fakedOperationService}
            ]
        });
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ListReplicationRuleComponent);
        comp = fixture.componentInstance;
        replicationService = fixture.debugElement.injector.get(ReplicationService);
        spyRules = spyOn(replicationService, 'getReplicationRules').and.returnValues(of(mockRules));

        fixture.detectChanges();
    });

    it('Should load and render data', async(() => {
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            fixture.detectChanges();
            let de: DebugElement = fixture.debugElement.query(By.css('datagrid-cell'));
            expect(de).toBeTruthy();
            fixture.detectChanges();
            let el: HTMLElement = de.nativeElement;
            expect(el.textContent.trim()).toEqual('sync_01');
        });
    }));
    it('should disable rule',   () => {
        fixture.detectChanges();
        comp.selectedRow = comp.rules[0];
        comp.selectedRow.enabled = true;
        fixture.detectChanges();
        const action: HTMLElement = fixture.nativeElement.querySelector("#rule-action");
        action.click();
        fixture.detectChanges();
        const disable: HTMLElement = fixture.nativeElement.querySelector("#rule-disable");
        disable.click();
        fixture.detectChanges();
        const button: HTMLElement = fixture.nativeElement.querySelector("#dialog-action-disable");
        button.click();
        fixture.detectChanges();
        const body: HTMLElement = fixture.nativeElement.querySelector(".modal-body");
        expect(body).toBeFalsy();
    });
    it('should enable rule',   () => {
        fixture.detectChanges();
        comp.selectedRow = comp.rules[0];
        comp.selectedRow.enabled = false;
        fixture.detectChanges();
        const action: HTMLElement = fixture.nativeElement.querySelector("#rule-action");
        action.click();
        fixture.detectChanges();
        const enable: HTMLElement = fixture.nativeElement.querySelector("#rule-enable");
        enable.click();
        fixture.detectChanges();
        const button: HTMLElement = fixture.nativeElement.querySelector("#dialog-action-enable");
        button.click();
        fixture.detectChanges();
        const body: HTMLElement = fixture.nativeElement.querySelector(".modal-body");
        expect(body).toBeFalsy();
    });
});
