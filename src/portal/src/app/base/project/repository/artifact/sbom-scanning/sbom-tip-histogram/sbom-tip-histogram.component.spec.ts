import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ClarityModule } from '@clr/angular';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { ActivatedRoute, Router } from '@angular/router';
import { SbomTipHistogramComponent } from './sbom-tip-histogram.component';
import { of } from 'rxjs';
import { Project } from '../../../../../../../app/base/project/project';
import { Artifact } from 'ng-swagger-gen/models';

describe('SbomTipHistogramComponent', () => {
    let component: SbomTipHistogramComponent;
    let fixture: ComponentFixture<SbomTipHistogramComponent>;
    const mockRouter = {
        navigate: () => {},
    };
    const mockedArtifact: Artifact = {
        id: 123,
        type: 'IMAGE',
    };
    const mockActivatedRoute = {
        RouterparamMap: of({ get: key => 'value' }),
        snapshot: {
            params: {
                repo: 'test',
                digest: 'ABC',
                subscribe: () => {
                    return of(null);
                },
            },
            parent: {
                params: {
                    id: 1,
                },
            },
            data: {
                artifactResolver: [mockedArtifact, new Project()],
            },
        },
        data: of({
            projectResolver: {
                ismember: true,
                role_name: 'maintainer',
            },
        }),
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [
                BrowserAnimationsModule,
                ClarityModule,
                TranslateModule.forRoot(),
            ],
            providers: [
                TranslateService,
                { provide: Router, useValue: mockRouter },
                { provide: ActivatedRoute, useValue: mockActivatedRoute },
            ],
            declarations: [SbomTipHistogramComponent],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(SbomTipHistogramComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        fixture.whenStable().then(() => {
            expect(component).toBeTruthy();
            expect(component.isLimitedSuccess()).toBeFalsy();
            expect(component.noSbom).toBeTruthy();
            expect(component.isThemeLight()).toBeFalsy();
            expect(component.duration()).toBe('0');
            expect(component.completePercent).toBe('0%');
        });
    });
});
