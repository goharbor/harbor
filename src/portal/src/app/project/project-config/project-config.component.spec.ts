import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ClarityModule } from '@clr/angular';
import { ActivatedRoute, Router } from "@angular/router";
import { SessionService } from '../../shared/session.service';
import { of } from 'rxjs';
import { ProjectConfigComponent } from './project-config.component';

describe('ProjectConfigComponent', () => {
  let component: ProjectConfigComponent;
  let fixture: ComponentFixture<ProjectConfigComponent>;
  let fakeSessionService = {
    getCurrentUser: function () {
      return { has_admin_role: true };
    }
  };
  let fakeRouter = null;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ProjectConfigComponent],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
      imports: [
        ClarityModule,
        TranslateModule.forRoot()
      ],
      providers: [
        {
          provide: ActivatedRoute, useValue: {
            paramMap: of({ get: (key) => 'value' }),
            snapshot: {
              parent: {
                params: { id: 1, chart: 'chart', version: 1.0 }
              },
              data: {
                projectResolver: {
                  role_name: 'admin'
                }
              }
            }
          }
        },
        { provide: Router, useValue: fakeRouter },
        { provide: SessionService, useValue: fakeSessionService },
        TranslateService
      ]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ProjectConfigComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
