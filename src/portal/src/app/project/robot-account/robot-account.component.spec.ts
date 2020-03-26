import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ClarityModule } from '@clr/angular';
import { of } from 'rxjs';
import { ActivatedRoute, Router } from "@angular/router";
import { MessageHandlerService } from "../../shared/message-handler/message-handler.service";
import { RobotService } from "./robot-account.service";
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ConfirmationDialogService } from "../../shared/confirmation-dialog/confirmation-dialog.service";
import { RobotAccountComponent } from './robot-account.component';
import { UserPermissionService } from "../../../lib/services";
import { ErrorHandler } from "../../../lib/utils/error-handler";
import { OperationService } from "../../../lib/components/operation/operation.service";

describe('RobotAccountComponent', () => {
  let component: RobotAccountComponent;
  let fixture: ComponentFixture<RobotAccountComponent>;
  let robotService = {
    listRobotAccount: function () {
      return of([]);
    }
  };
  let mockConfirmationDialogService = null;
  let mockUserPermissionService = {
    getPermission: function () {
      return 1;
    }
  };
  let mockErrorHandler = {
    error: function () { }
  };
  let mockMessageHandlerService = null;
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
      imports: [
        ClarityModule,
        TranslateModule.forRoot()
      ],
      providers: [
        {
          provide: ActivatedRoute, useValue: {
            paramMap: of({ get: (key) => 'value' }),
            snapshot: {
              parent: {
                params: { id: 1 }
              },
              data: 1
            }
          }
        },
        TranslateService,
        { provide: RobotService, useValue: robotService },
        { provide: ConfirmationDialogService, useClass: ConfirmationDialogService },
        { provide: UserPermissionService, useValue: mockUserPermissionService },
        { provide: ErrorHandler, useValue: mockErrorHandler },
        { provide: MessageHandlerService, useValue: mockMessageHandlerService },
        OperationService
      ],
      declarations: [RobotAccountComponent]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(RobotAccountComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
