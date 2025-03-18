// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ArtifactSummaryComponent } from './artifact-summary.component';
import { of } from 'rxjs';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { Artifact } from '../../../../../../ng-swagger-gen/models/artifact';
import { ProjectService } from '../../../../shared/services';
import { ActivatedRoute, Router } from '@angular/router';
import { AppConfigService } from '../../../../services/app-config.service';
import { Project } from '../../project';
import { ArtifactService } from './artifact.service';
import { SharedTestingModule } from '../../../../shared/shared.module';

describe('ArtifactSummaryComponent', () => {
    const mockedArtifact: Artifact = {
        id: 123,
        type: 'IMAGE',
    };

    const fakedProjectService = {
        getProject() {
            return of({ name: 'test' });
        },
    };

    const fakedArtifactDefaultService = {
        getIconsFromBackEnd() {
            return undefined;
        },
        getIcon() {
            return undefined;
        },
    };
    const mockedSbomDigest =
        'sha256:51a41cec9de9d62ee60e206f5a8a615a028a65653e45539990867417cb486285';
    let component: ArtifactSummaryComponent;
    let fixture: ComponentFixture<ArtifactSummaryComponent>;
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
            queryParams: {
                sbomDigest: mockedSbomDigest,
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
    const fakedAppConfigService = {
        getConfig: () => {
            return { with_admiral: false };
        },
    };
    const mockRouter = {
        navigate: () => {},
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ArtifactSummaryComponent],
            schemas: [NO_ERRORS_SCHEMA],
            providers: [
                { provide: AppConfigService, useValue: fakedAppConfigService },
                { provide: Router, useValue: mockRouter },
                { provide: ActivatedRoute, useValue: mockActivatedRoute },
                { provide: ProjectService, useValue: fakedProjectService },
                {
                    provide: ArtifactService,
                    useValue: fakedArtifactDefaultService,
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ArtifactSummaryComponent);
        component = fixture.componentInstance;
        component.repositoryName = 'demo';
        component.artifactDigest = 'sha: acf4234f';
        component.sbomDigest = mockedSbomDigest;
        fixture.detectChanges();
    });

    it('should create and get artifactDetails', async () => {
        expect(component).toBeTruthy();
        await fixture.whenStable();
        expect(component.artifact.type).toEqual('IMAGE');
    });
});
