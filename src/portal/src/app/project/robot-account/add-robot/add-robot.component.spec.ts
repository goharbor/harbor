import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ClarityModule } from '@clr/angular';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { AddRobotComponent } from './add-robot.component';
import { FormsModule } from '@angular/forms';
import { RobotService } from "../robot-account.service";
import { of } from "rxjs";
import { ErrorHandler } from "@harbor/ui";
import { MessageHandlerService } from "../../../shared/message-handler/message-handler.service";
import { TranslateModule, TranslateService } from '@ngx-translate/core';

describe('AddRobotComponent', () => {
  let component: AddRobotComponent;
  let fixture: ComponentFixture<AddRobotComponent>;
  let fakeRobotService = {
    listRobotAccount: function () {
      return of([{
        name: "robot$" + 1
      }, {
        name: "abc"
      }]);
    }
  };
  let fakeMessageHandlerService = {
    showSuccess: function() {}
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [AddRobotComponent],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
      imports: [
        ClarityModule,
        TranslateModule.forRoot(),
        FormsModule
      ],
      providers: [
        TranslateService,
        ErrorHandler,
        { provide: MessageHandlerService, useValue: fakeMessageHandlerService },
        { provide: RobotService, useValue: fakeRobotService }
      ]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(AddRobotComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
