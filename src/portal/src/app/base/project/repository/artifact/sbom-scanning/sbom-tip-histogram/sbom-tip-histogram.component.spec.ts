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
    const mockedScanner = {
        name: 'Trivy',
        vendor: 'vm',
        version: 'v1.2',
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

    it('Test SbomTipHistogramComponent basic functions', () => {
        fixture.whenStable().then(() => {
            expect(component).toBeTruthy();
            expect(component.showNoSbom()).toBeTruthy();
            expect(component.isThemeLight()).toBeFalsy();
            expect(component.duration()).toBe('0');
            expect(component.completePercent).toBe('0%');
        });
    });

    it('Test SbomTipHistogramComponent completeTimestamp', () => {
        fixture.whenStable().then(() => {
            component.sbomSummary.end_time = new Date('2024-04-08 00:01:02');
            expect(component.completeTimestamp).toBe(
                component.sbomSummary.end_time
            );
        });
    });

    it('Test SbomTipHistogramComponent getScannerInfo', () => {
        fixture.whenStable().then(() => {
            expect(component.getScannerInfo()).toBe('');
            component.scanner = mockedScanner;
            expect(component.getScannerInfo()).toBe(
                `${mockedScanner.name}@${mockedScanner.version}`
            );
        });
    });
});
