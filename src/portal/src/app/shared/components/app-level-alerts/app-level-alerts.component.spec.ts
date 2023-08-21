import { ComponentFixture, TestBed } from '@angular/core/testing';
import { AppLevelAlertsComponent } from './app-level-alerts.component';
import { SharedTestingModule } from '../../shared.module';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { Scanner } from '../../../base/left-side-nav/interrogation-services/scanner/scanner';
import { ScannerService } from 'ng-swagger-gen/services/scanner.service';
import { SessionService } from 'src/app/shared/services/session.service';
import { AppConfigService } from '../../../services/app-config.service';

describe('AppLevelAlertsComponent', () => {
    let component: AppLevelAlertsComponent;
    let fixture: ComponentFixture<AppLevelAlertsComponent>;

    const fakeScannerService = {
        listScannersResponse() {
            const response: HttpResponse<Array<Scanner>> = new HttpResponse<
                Array<Scanner>
            >({
                headers: new HttpHeaders({
                    'x-total-count': '1',
                }),
                body: [
                    {
                        name: 'test',
                        is_default: true,
                    },
                ],
            });
            return of(response).pipe(delay(0));
        },
        listScanners() {
            return of([
                {
                    name: 'test',
                    is_default: true,
                },
            ]).pipe(delay(0));
        },
    };

    const fakeSessionService = {
        getCurrentUser: function () {
            return { has_admin_role: true };
        },
    };
    const MockedAppConfigService = {
        getConfig() {
            return { read_only: false };
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [AppLevelAlertsComponent],
            providers: [
                {
                    provide: ScannerService,
                    useValue: fakeScannerService,
                },
                { provide: SessionService, useValue: fakeSessionService },
                { provide: AppConfigService, useValue: MockedAppConfigService },
            ],
        }).compileComponents();

        fixture = TestBed.createComponent(AppLevelAlertsComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should show scanner alert', async () => {
        await fixture.whenStable();
        fixture.detectChanges();
        const compiled = fixture.nativeElement;
        expect(compiled.querySelector('.alerts.alert-info')).toBeTruthy();
    });
});
