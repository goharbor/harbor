import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { EditProjectQuotasComponent } from './edit-project-quotas.component';
import { SharedModule } from '../../../shared/shared.module';
import { InlineAlertComponent } from '../../../inline-alert/inline-alert.component';
import { SERVICE_CONFIG, IServiceConfig } from '../../../service.config';
import { RouterModule } from '@angular/router';

describe('EditProjectQuotasComponent', () => {
  let component: EditProjectQuotasComponent;
  let fixture: ComponentFixture<EditProjectQuotasComponent>;
  let config: IServiceConfig = {
    quotaUrl: "/api/quotas/testing"
  };
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule,
        RouterModule.forRoot([])
      ],
      declarations: [ EditProjectQuotasComponent, InlineAlertComponent ],
      providers: [
        { provide: SERVICE_CONFIG, useValue: config },
      ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(EditProjectQuotasComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });
  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
