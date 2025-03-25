// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
