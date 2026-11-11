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
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { OidcOnboardService } from './oidc-onboard.service';
import { Router, ActivatedRoute } from '@angular/router';
import { of } from 'rxjs';
import { OidcOnboardComponent } from './oidc-onboard.component';
import { SharedTestingModule } from '../shared/shared.module';

describe('OidcOnboardComponent', () => {
    let component: OidcOnboardComponent;
    let fixture: ComponentFixture<OidcOnboardComponent>;
    let fakeOidcOnboardService = null;
    let fakeRouter = null;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [OidcOnboardComponent],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            providers: [
                {
                    provide: OidcOnboardService,
                    useValue: fakeOidcOnboardService,
                },
                { provide: Router, useValue: fakeRouter },
                {
                    provide: ActivatedRoute,
                    useValue: {
                        queryParams: of({
                            view: 'abc',
                            objectId: 'ddd',
                            actionUid: 'ddd',
                            targets: '',
                            locale: '',
                        }),
                    },
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(OidcOnboardComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
