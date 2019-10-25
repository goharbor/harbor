import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { RepositoryPageComponent } from './repository-page.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { of } from 'rxjs';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { ActivatedRoute, Router } from '@angular/router';
import { SessionService } from '../shared/session.service';
describe('RepositoryPageComponent', () => {
    let component: RepositoryPageComponent;
    let fixture: ComponentFixture<RepositoryPageComponent>;
    const mockActivatedRoute = {
        RouterparamMap: of({ get: (key) => 'value' }),
        snapshot: {
            parent: {
                params: { id: 1 },
                data: {
                    projectResolver: {
                        ismember: true,
                        name: 'library',
                    }
                }
            }
        }
    };
    const mockSessionService = {
        getCurrentUser: () => { }
    };
    beforeEach(async(() => {
        TestBed.configureTestingModule({
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            imports: [
                BrowserAnimationsModule,
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule,
                RouterTestingModule,
                NoopAnimationsModule,
                HttpClientTestingModule
            ],
            declarations: [RepositoryPageComponent],
            providers: [
                TranslateService,
                { provide: ActivatedRoute, useValue: mockActivatedRoute },
                { provide: SessionService, useValue: mockSessionService },

            ]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(RepositoryPageComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
