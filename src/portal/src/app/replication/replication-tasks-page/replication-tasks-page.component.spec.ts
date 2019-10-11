import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ReplicationTasksPageComponent } from './replication-tasks-page.component';
import { ActivatedRoute, Router } from '@angular/router';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { of } from 'rxjs';

describe('ReplicationTasksPageComponent', () => {
  let component: ReplicationTasksPageComponent;
  let fixture: ComponentFixture<ReplicationTasksPageComponent>;
  let mockActivatedRoute = {
    snapshot: {
      params: 1
    }
  };
  let mockRouter = null;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ReplicationTasksPageComponent],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
      imports: [
        TranslateModule.forRoot()
      ],
      providers: [
        {
          provide: ActivatedRoute, useValue: mockActivatedRoute
        },
        {
          provide: Router, useValue: mockRouter
        },
        TranslateService
      ]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ReplicationTasksPageComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
