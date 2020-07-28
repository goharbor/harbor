import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { P2pProviderComponent } from './p2p-provider.component';
import { NO_ERRORS_SCHEMA } from '@angular/core';

describe('P2pProviderComponent', () => {
  let component: P2pProviderComponent;
  let fixture: ComponentFixture<P2pProviderComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ P2pProviderComponent ],
      schemas: [
        NO_ERRORS_SCHEMA
      ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(P2pProviderComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
