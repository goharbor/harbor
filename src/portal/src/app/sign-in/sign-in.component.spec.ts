import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { SignInComponent } from './sign-in.component';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { RouterTestingModule } from '@angular/router/testing';
import { AppConfigService } from '../services/app-config.service';
import { SessionService } from '../shared/session.service';
import { CookieService } from 'ngx-cookie';
import { SkinableConfig } from "../services/skinable-config.service";
import { CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA } from '@angular/core';
import { ClarityModule } from "@clr/angular";
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { of } from "rxjs";

describe('SignInComponent', () => {
    let component: SignInComponent;
    let fixture: ComponentFixture<SignInComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                TranslateModule.forRoot(),
                RouterTestingModule,
                ClarityModule,
                FormsModule,
                ReactiveFormsModule
            ],
            declarations: [SignInComponent],
            providers: [
                TranslateService,
                { provide: SessionService, useValue: null },
                {
                    provide: AppConfigService, useValue: {
                        load: function () {
                            return of({

                            });
                        }
                    }
                },
                {
                    provide: CookieService, useValue: {
                        get: function (key) {
                            return key;
                        }
                    }
                },
                {
                    provide: SkinableConfig, useValue: {
                        getSkinConfig: function () {
                            return {
                                loginBgImg: "abc",
                                appTitle: "Harbor"
                            };
                        }
                    }
                }
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(SignInComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
