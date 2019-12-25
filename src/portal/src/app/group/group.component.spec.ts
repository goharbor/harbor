import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { GroupComponent } from './group.component';
import { ClarityModule } from '@clr/angular';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { FormsModule } from '@angular/forms';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { SessionService } from "./../shared/session.service";
import { GroupService } from "./group.service";
import { of } from "rxjs";
import { ConfirmationDialogService } from "./../shared/confirmation-dialog/confirmation-dialog.service";
import { MessageHandlerService } from '../shared/message-handler/message-handler.service';
import { AppConfigService } from '../app-config.service';
import { OperationService } from "../../lib/components/operation/operation.service";

describe('GroupComponent', () => {
  let component: GroupComponent;
  let fixture: ComponentFixture<GroupComponent>;
  let fakeMessageHandlerService = null;
  let fakeOperationService = null;
  let fakeGroupService = {
    getUserGroups: function () {
      return of([{
        group_name: ''
      }, {
        group_name: 'abc'
      }]);
    }
  };
  let fakeConfirmationDialogService = {
    confirmationConfirm$: of({
      state: 1,
      source: 2
    })
  };
  let fakeSessionService = {
    currentUser: {
      sysadmin_flag: true,
      admin_role_in_auth: true,
    }
  };
  let fakeAppConfigService = {
    isLdapMode: function () {
      return true;
    }
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [GroupComponent],
      imports: [
        ClarityModule,
        FormsModule,
        TranslateModule.forRoot()
      ],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
      providers: [
        TranslateService,
        { provide: MessageHandlerService, useValue: fakeMessageHandlerService },
        { provide: OperationService, useValue: fakeOperationService },
        { provide: GroupService, useValue: fakeGroupService },
        { provide: ConfirmationDialogService, useValue: fakeConfirmationDialogService },
        { provide: SessionService, useValue: fakeSessionService },
        { provide: AppConfigService, useValue: fakeAppConfigService }
      ]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(GroupComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
