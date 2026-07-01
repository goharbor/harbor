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
import { ArtifactAdditionsComponent } from './artifact-additions.component';
import { DockerfileComponent } from './dockerfile/dockerfile.component';
import { AdditionLinks } from '../../../../../../../ng-swagger-gen/models/addition-links';
import { ADDITIONS } from './models';
import { CURRENT_BASE_HREF } from '../../../../../shared/units/utils';
import { SharedTestingModule } from '../../../../../shared/shared.module';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ArtifactListPageService } from '../artifact-list-page/artifact-list-page.service';
import { ClrLoadingState } from '@clr/angular';

describe('ArtifactAdditionsComponent', () => {
    const mockedAdditionLinks: AdditionLinks = {
        vulnerabilities: {
            absolute: false,
            href: CURRENT_BASE_HREF + '/test',
        },
    };
    const mockedArtifactListPageService = {
        hasScannerSupportSBOM(): boolean {
            return true;
        },
        hasEnabledScanner(): boolean {
            return true;
        },
        getScanBtnState(): ClrLoadingState {
            return ClrLoadingState.SUCCESS;
        },
        init() {},
    };
    let component: ArtifactAdditionsComponent;
    let fixture: ComponentFixture<ArtifactAdditionsComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ArtifactAdditionsComponent, DockerfileComponent],
            schemas: [NO_ERRORS_SCHEMA],
            providers: [
                {
                    provide: ArtifactListPageService,
                    useValue: mockedArtifactListPageService,
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ArtifactAdditionsComponent);
        component = fixture.componentInstance;
        component.additionLinks = mockedAdditionLinks;
        component.tab = 'vulnerability';
        fixture.detectChanges();
    });

    it('should create and render vulnerabilities tab', async () => {
        expect(component).toBeTruthy();
        await fixture.whenStable();
        const tabButton: HTMLButtonElement =
            fixture.nativeElement.querySelector('#vulnerability');
        expect(tabButton).toBeTruthy();
    });

    it('should return null for dockerfile when not available', () => {
        const dockerfileLink = component.getDockerfile();
        expect(dockerfileLink).toBeNull();
    });

    it('should return dockerfile link when available', () => {
        component.additionLinks = {
            ...mockedAdditionLinks,
            dockerfile: { absolute: false, href: '/test-dockerfile' },
        };
        const dockerfileLink = component.getDockerfile();
        expect(dockerfileLink).toBeTruthy();
        expect(dockerfileLink.href).toBe('/test-dockerfile');
    });

    it('should return false for hasBuildHistory when not available', () => {
        const hasBuildHistory = component.hasBuildHistory();
        expect(hasBuildHistory).toBe(false);
    });

    it('should return true for hasBuildHistory when available', () => {
        component.additionLinks = {
            ...mockedAdditionLinks,
            build_history: { absolute: false, href: '/test-history' },
        };
        const hasBuildHistory = component.hasBuildHistory();
        expect(hasBuildHistory).toBe(true);
    });

    it('should call actionTab on onDockerfileSwitchTab', () => {
        spyOn(component, 'actionTab');
        component.onDockerfileSwitchTab('build-history');
        expect(component.actionTab).toHaveBeenCalledWith('build-history');
    });
});
