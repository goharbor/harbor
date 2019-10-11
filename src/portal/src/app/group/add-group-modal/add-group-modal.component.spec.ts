import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ClarityModule } from '@clr/angular';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CUSTOM_ELEMENTS_SCHEMA, ChangeDetectorRef } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { GroupService } from "../group.service";
import { MessageHandlerService } from "./../../shared/message-handler/message-handler.service";
import { SessionService } from "./../../shared/session.service";
import { UserGroup } from "./../group";
import { AppConfigService } from "../../app-config.service";
import { AddGroupModalComponent } from './add-group-modal.component';

describe('AddGroupModalComponent', () => {
  let component: AddGroupModalComponent;
  let fixture: ComponentFixture<AddGroupModalComponent>;
  let fakeSessionService = {
    getCurrentUser: function () {
      return { has_admin_role: true };
    }
  };
  let fakeGroupService = null;
  let fakeAppConfigService = {
    isLdapMode: function () {
      return true;
    },
    isHttpAuthMode: function () {
      return false;
    },
    isOidcMode: function () {
      return false;
    }
  };
  let fakeMessageHandlerService = null;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [AddGroupModalComponent],
      imports: [
        ClarityModule,
        FormsModule,
        TranslateModule.forRoot()
      ],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
      providers: [
        ChangeDetectorRef,
        { provide: MessageHandlerService, useValue: fakeMessageHandlerService },
        { provide: SessionService, useValue: fakeSessionService },
        { provide: AppConfigService, useValue: fakeAppConfigService },
        { provide: GroupService, useValue: fakeGroupService },
      ]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(AddGroupModalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
