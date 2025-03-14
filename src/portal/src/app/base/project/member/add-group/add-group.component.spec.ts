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
import { of } from 'rxjs';
import { AppConfigService } from '../../../../services/app-config.service';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { AddGroupComponent } from './add-group.component';
import { MemberService } from 'ng-swagger-gen/services/member.service';

describe('AddHttpAuthGroupComponent', () => {
    let component: AddGroupComponent;
    let fixture: ComponentFixture<AddGroupComponent>;
    let fakeAppConfigService = {
        isLdapMode: function () {
            return true;
        },
    };

    let fakeMemberService = {
        listProjectMembers: function () {
            return of(null);
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [AddGroupComponent],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            providers: [
                TranslateService,
                { provide: AppConfigService, useValue: fakeAppConfigService },
                { provide: MemberService, useValue: fakeMemberService },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(AddGroupComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
