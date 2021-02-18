import { waitForAsync, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateService } from '@ngx-translate/core';
import { TagRetentionComponent } from './tag-retention.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { of } from 'rxjs';
import { ActivatedRoute } from '@angular/router';
import { AddRuleComponent } from "./add-rule/add-rule.component";
import { TagRetentionService } from "./tag-retention.service";
import { RuleMetadate, Retention } from './retention';
import { delay } from 'rxjs/operators';
import { ErrorHandler } from "../../../../shared/units/error-handler";
import { SharedTestingModule } from "../../../../shared/shared.module";

describe('TagRetentionComponent', () => {
    let component: TagRetentionComponent;
    let fixture: ComponentFixture<TagRetentionComponent>;
    const mockTagRetentionService = {
        createRetention: () => of(null).pipe(delay(0)),
        updateRetention: () => of(null).pipe(delay(0)),
        runNowTrigger: () => of(null).pipe(delay(0)),
        whatIfRunTrigger: () => of(null).pipe(delay(0)),
        AbortRun: () => of(null).pipe(delay(0)),
        seeLog: () => of(null).pipe(delay(0)),
        getExecutionHistory: () => of({
            body: []
        }).pipe(delay(0)),
        getRunNowList: () => of({
            body: []
        }).pipe(delay(0)),
        getProjectInfo: () => of({
            metadata: {
                retention_id: 1
            }
        }).pipe(delay(0)),
        getRetentionMetadata: () => of(new RuleMetadate()).pipe(delay(0)),
        getRetention: () => of(new Retention()).pipe(delay(0)),
    };
    const mockActivatedRoute = {
        snapshot: {
            parent: {
                parent: {
                    parent: {
                        params: { id: 1 },
                        data: {
                            projectResolver: {
                                metadata: {
                                    retention_id: 1
                                }
                            }
                        }
                    }
                }
            }
        }
    };
    const mockErrorHandler = {
        error: () => { }
    };
    beforeEach(waitForAsync(() => {
        TestBed.configureTestingModule({
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            imports: [
                SharedTestingModule
            ],
            declarations: [TagRetentionComponent, AddRuleComponent],
            providers: [
                TranslateService,
                { provide: TagRetentionService, useValue: mockTagRetentionService },
                { provide: ActivatedRoute, useValue: mockActivatedRoute },
                { provide: ErrorHandler, useValue: mockErrorHandler }

            ]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(TagRetentionComponent);
        component = fixture.componentInstance;
        component.loadingHistories = false;
        component.loadingRule = false;
        component.loadingHistories = false;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
