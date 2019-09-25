import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { ClarityModule } from '@clr/angular';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { UserService } from "../user.service";
import { SharedModule } from '../../shared/shared.module';
import { ChangePasswordComponent } from './change-password.component';

describe('ChangePasswordComponent', () => {
    let component: ChangePasswordComponent;
    let fixture: ComponentFixture<ChangePasswordComponent>;
    let fakeUserService = null;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [ChangePasswordComponent],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            imports: [
                ClarityModule,
                SharedModule,
                TranslateModule.forRoot()
            ],
            providers: [
                { provide: UserService, useValue: fakeUserService }
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ChangePasswordComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
