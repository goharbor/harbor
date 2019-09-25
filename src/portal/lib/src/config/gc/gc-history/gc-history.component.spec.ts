import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { SharedModule } from '../../../shared/shared.module';
import { GcRepoService } from "../gc.service";
import { of } from 'rxjs';
import { GcViewModelFactory } from "../gc.viewmodel.factory";
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ErrorHandler } from '../../../error-handler';
import { GcHistoryComponent } from './gc-history.component';

describe('GcHistoryComponent', () => {
    let component: GcHistoryComponent;
    let fixture: ComponentFixture<GcHistoryComponent>;
    let fakeGcRepoService = {
        getJobs: function () {
            return of([]);
        }
    };
    let fakeGcViewModelFactory = {
        createJobViewModel: function (data) {
            return data;
        }
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [GcHistoryComponent],
            imports: [
                SharedModule,
                TranslateModule.forRoot()
            ],
            providers: [
                ErrorHandler,
                TranslateService,
                { provide: GcRepoService, useValue: fakeGcRepoService },
                { provide: GcViewModelFactory, useValue: fakeGcViewModelFactory }
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(GcHistoryComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
