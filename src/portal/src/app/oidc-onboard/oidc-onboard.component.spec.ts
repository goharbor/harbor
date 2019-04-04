import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { OidcOnboardComponent } from './oidc-onboard.component';

describe('OidcOnboardComponent', () => {
  let component: OidcOnboardComponent;
  let fixture: ComponentFixture<OidcOnboardComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ OidcOnboardComponent ]
    })
    .compileComponents();
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
