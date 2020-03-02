import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ClarityModule } from '@clr/angular';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CUSTOM_ELEMENTS_SCHEMA, ChangeDetectorRef } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { SessionService } from '../../shared/session.service';
import { ListChartsComponent } from './list-charts.component';

describe('ListChartsComponent', () => {
  let component: ListChartsComponent;
  let fixture: ComponentFixture<ListChartsComponent>;
  let fakeSessionService = {
    getCurrentUser: function () {
      return "admin";
    }
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ListChartsComponent],
      imports: [
        ClarityModule,
        TranslateModule.forRoot()
      ],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
      providers: [
        {
          provide: ActivatedRoute, useValue: {
            snapshot: {
              parent: {
                params: {
                  id: 1,
                  data: 'chart'
                }
              }
            }
          }
        },
        { provide: Router, useValue: null },
        { provide: SessionService, useValue: fakeSessionService }
      ]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ListChartsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
