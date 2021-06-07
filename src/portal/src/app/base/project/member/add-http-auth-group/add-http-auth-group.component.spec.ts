import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { of } from "rxjs";
import { MemberService } from '../member.service';
import { AppConfigService } from "../../../../services/app-config.service";
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';
import { AddHttpAuthGroupComponent } from './add-http-auth-group.component';
import { SharedTestingModule } from "../../../../shared/shared.module";

describe('AddHttpAuthGroupComponent', () => {
  let component: AddHttpAuthGroupComponent;
  let fixture: ComponentFixture<AddHttpAuthGroupComponent>;
  let fakeAppConfigService = {
    isHttpAuthMode: function () {
      return true;
    }
  };

  let fakeMemberService = {addGroupMember: function() {
    return of(null);
  }};

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [AddHttpAuthGroupComponent],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
      imports: [
        SharedTestingModule
      ],
      providers: [
        TranslateService,
        { provide: AppConfigService, useValue: fakeAppConfigService },
        { provide: MemberService, useValue: fakeMemberService }
      ],
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(AddHttpAuthGroupComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
