import { waitForAsync, ComponentFixture, TestBed } from '@angular/core/testing';
import { ClarityModule } from '@clr/angular';
import { AddRobotComponent } from './add-robot.component';
import { FormsModule } from '@angular/forms';
import { of } from "rxjs";
import { MessageHandlerService } from "../../../../shared/services/message-handler.service";
import { TranslateModule } from '@ngx-translate/core';
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { delay } from "rxjs/operators";
import { RobotService } from "../../../../../../ng-swagger-gen/services/robot.service";
import { OperationService } from "../../../../shared/components/operation/operation.service";
import { NO_ERRORS_SCHEMA } from "@angular/core";
import { SharedTestingModule } from "../../../../shared/shared.module";

describe('AddRobotComponent', () => {
  let component: AddRobotComponent;
  let fixture: ComponentFixture<AddRobotComponent>;
  const fakedRobotService = {
    ListRobot() {
      return of([]).pipe(delay(0));
    }
  };
  const fakedMessageHandlerService = {
    showSuccess() {
    },
    error() {
    }
  };
  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [AddRobotComponent],
      imports: [
        BrowserAnimationsModule,
        ClarityModule,
        TranslateModule.forRoot(),
        FormsModule,
        SharedTestingModule
      ],
      providers: [
        OperationService,
        { provide: RobotService, useValue: fakedRobotService },
        { provide: MessageHandlerService, useValue: fakedMessageHandlerService },
      ],
      schemas: [
        NO_ERRORS_SCHEMA
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
