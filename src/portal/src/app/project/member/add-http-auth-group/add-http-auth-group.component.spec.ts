import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { of } from "rxjs";
import { MemberService } from '../member.service';
import { AppConfigService } from "../../../services/app-config.service";
import { ClarityModule } from '@clr/angular';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { AddHttpAuthGroupComponent } from './add-http-auth-group.component';
import { HarborLibraryModule } from "../../../../lib/harbor-library.module";

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

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [AddHttpAuthGroupComponent],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
      imports: [
        ClarityModule,
        HarborLibraryModule,
        TranslateModule.forRoot()
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
