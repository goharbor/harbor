import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { LicenseComponent } from './license.component';

fdescribe('LicenseComponent', () => {
  let component: LicenseComponent;
  let fixture: ComponentFixture<LicenseComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [LicenseComponent],
      imports: [
        HttpClientTestingModule
      ]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(LicenseComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
