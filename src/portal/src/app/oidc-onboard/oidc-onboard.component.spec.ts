import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { OidcOnboardService } from './oidc-onboard.service';
import { Router, ActivatedRoute } from '@angular/router';
import { of } from 'rxjs';
import { OidcOnboardComponent } from './oidc-onboard.component';
import { SharedTestingModule } from '../shared/shared.module';

describe('OidcOnboardComponent', () => {
    let component: OidcOnboardComponent;
    let fixture: ComponentFixture<OidcOnboardComponent>;
    let fakeOidcOnboardService = null;
    let fakeRouter = null;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [OidcOnboardComponent],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            providers: [
                {
                    provide: OidcOnboardService,
                    useValue: fakeOidcOnboardService,
                },
                { provide: Router, useValue: fakeRouter },
                {
                    provide: ActivatedRoute,
                    useValue: {
                        queryParams: of({
                            view: 'abc',
                            objectId: 'ddd',
                            actionUid: 'ddd',
                            targets: '',
                            locale: '',
                        }),
                    },
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(OidcOnboardComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
