import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { ClarityModule } from '@clr/angular';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { FormsModule } from '@angular/forms';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { AddGroupComponent } from './add-group.component';
import { GroupService } from "../../../left-side-nav/group/group.service";
import { MemberService } from "../member.service";
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { OperationService } from "../../../../shared/components/operation/operation.service";

describe('AddGroupComponent', () => {
  let component: AddGroupComponent;
  let fixture: ComponentFixture<AddGroupComponent>;
  let fakeMessageHandlerService = null;
  let fakeOperationService = null;
  let fakeGroupService = null;
  let fakeMemberService = null;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [AddGroupComponent],
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
        { provide: MemberService, useValue: fakeMemberService }
      ]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(AddGroupComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
