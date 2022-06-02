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
