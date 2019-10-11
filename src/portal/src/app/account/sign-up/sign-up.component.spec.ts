import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ClarityModule } from "@clr/angular";
import { SignUpComponent } from './sign-up.component';
import { SessionService } from '../../shared/session.service';
import { UserService } from '../../user/user.service';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { NewUserFormComponent } from '../../shared/new-user-form/new-user-form.component';
import { FormsModule } from '@angular/forms';

describe('SignUpComponent', () => {
    let component: SignUpComponent;
    let fixture: ComponentFixture<SignUpComponent>;
    let fakeSessionService = null;
    let fakeUserService = null;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [SignUpComponent, NewUserFormComponent],
            imports: [
                FormsModule,
                ClarityModule,
                TranslateModule.forRoot()
            ],
            providers: [
                TranslateService,
                { provide: SessionService, useValue: fakeSessionService },
                { provide: UserService, useValue: fakeUserService }
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(SignUpComponent);
        component = fixture.componentInstance;
        component.newUserForm =
            TestBed.createComponent(NewUserFormComponent).componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
