import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ClarityModule } from '@clr/angular';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { OidcOnboardService } from './oidc-onboard.service';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { Router, ActivatedRoute } from '@angular/router';
import { of } from 'rxjs';
import { OidcOnboardComponent } from './oidc-onboard.component';

describe('OidcOnboardComponent', () => {
  let component: OidcOnboardComponent;
  let fixture: ComponentFixture<OidcOnboardComponent>;
  let fakeOidcOnboardService = null;
  let fakeRouter = null;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [OidcOnboardComponent],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
      imports: [
        ClarityModule,
        FormsModule,
        ReactiveFormsModule,
        TranslateModule.forRoot()
      ],
      providers: [
        TranslateService,
        { provide: OidcOnboardService, useValue: fakeOidcOnboardService },
        { provide: Router, useValue: fakeRouter },
        {
          provide: ActivatedRoute, useValue: {
            queryParams: of({
              view: 'abc',
              objectId: 'ddd',
              actionUid: 'ddd',
              targets: '',
              locale: ''
            })
          }
        }
      ]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(OidcOnboardComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
