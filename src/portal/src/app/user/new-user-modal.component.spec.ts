import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ClarityModule } from '@clr/angular';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { UserService } from './user.service';
import { SessionService } from '../shared/session.service';
import { MessageHandlerService } from '../shared/message-handler/message-handler.service';
import { SharedModule } from '../shared/shared.module';
import { NewUserModalComponent } from './new-user-modal.component';

describe('NewUserModalComponent', () => {
    let component: NewUserModalComponent;
    let fixture: ComponentFixture<NewUserModalComponent>;
    let fakeSessionService = null;
    let fakeUserService = null;
    let fakeMessageHandlerService = {
        handleError: function () { }
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [NewUserModalComponent],
            imports: [
                ClarityModule,
                SharedModule,
                TranslateModule.forRoot()
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            providers: [
                { provide: MessageHandlerService, useValue: fakeMessageHandlerService },
                { provide: UserService, useValue: fakeUserService },
                { provide: SessionService, useValue: fakeSessionService }
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(NewUserModalComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
