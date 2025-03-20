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
import { TranslateModule } from '@ngx-translate/core';
import { ClarityModule } from '@clr/angular';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { DistributionSetupModalComponent } from './distribution-setup-modal.component';
import { PreheatService } from '../../../../../../ng-swagger-gen/services/preheat.service';
import { Instance } from '../../../../../../ng-swagger-gen/models/instance';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';

describe('DistributionSetupModalComponent', () => {
    let component: DistributionSetupModalComponent;
    let fixture: ComponentFixture<DistributionSetupModalComponent>;

    const instance1: Instance = {
        name: 'Test1',
        default: true,
        enabled: true,
        description: 'Test1',
        endpoint: 'http://test.com',
        id: 1,
        setup_timestamp: new Date().getTime(),
        auth_mode: 'NONE',
        vendor: 'kraken',
        status: 'Healthy',
    };
    const fakedPreheatService = {
        ListInstances() {
            return of([]).pipe(delay(0));
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [
                ClarityModule,
                TranslateModule,
                SharedTestingModule,
                HttpClientTestingModule,
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            providers: [
                { provide: PreheatService, useValue: fakedPreheatService },
            ],
            declarations: [DistributionSetupModalComponent],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(DistributionSetupModalComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should show "name is required"', async () => {
        fixture.autoDetectChanges();
        component._open();
        await fixture.whenStable();
        const nameInput = fixture.nativeElement.querySelector('#name');
        nameInput.value = '';
        nameInput.dispatchEvent(new Event('input'));
        nameInput.blur();
        nameInput.dispatchEvent(new Event('blur'));
        let el = fixture.nativeElement.querySelector('clr-control-error');
        expect(el).toBeTruthy();
    });

    it('should show "endpoint is required"', async () => {
        fixture.autoDetectChanges();
        component._open();
        await fixture.whenStable();
        const endpointInput = fixture.nativeElement.querySelector('#endpoint');
        endpointInput.value = 'svn://test.com';
        endpointInput.dispatchEvent(new Event('input'));
        endpointInput.blur();
        endpointInput.dispatchEvent(new Event('blur'));
        let el = fixture.nativeElement.querySelector('clr-control-error');
        expect(el).toBeTruthy();
    });

    it('should be edit model', async () => {
        fixture.autoDetectChanges();
        component.openSetupModal(true, instance1);
        await fixture.whenStable();
        const nameInput = fixture.nativeElement.querySelector('#name');
        expect(nameInput.value).toEqual('Test1');
    });

    it('should be valid', async () => {
        fixture.autoDetectChanges();
        component._open();
        await fixture.whenStable();
        component.model.vendor = 'kraken';
        const nameInput = fixture.nativeElement.querySelector('#name');
        nameInput.value = 'test';
        nameInput.dispatchEvent(new Event('input'));
        const endpointInput = fixture.nativeElement.querySelector('#endpoint');
        endpointInput.value = 'https://test.com';
        endpointInput.dispatchEvent(new Event('input'));
        await fixture.whenStable();
        expect(component.isValid).toBeTruthy();
    });
});
