import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { TagFeatureIntegrationComponent } from './tag-feature-integration.component';
import { RouterModule } from '@angular/router';
import { TranslateModule } from '@ngx-translate/core';
import { RouterTestingModule } from '@angular/router/testing';

describe('TagFeatureIntegrationComponent', () => {
  let component: TagFeatureIntegrationComponent;
  let fixture: ComponentFixture<TagFeatureIntegrationComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ TagFeatureIntegrationComponent ],
      imports: [ RouterModule, TranslateModule.forRoot(), RouterTestingModule ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(TagFeatureIntegrationComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
