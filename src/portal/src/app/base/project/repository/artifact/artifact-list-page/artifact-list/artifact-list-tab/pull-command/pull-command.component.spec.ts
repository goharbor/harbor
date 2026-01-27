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
import { PullCommandComponent } from './pull-command.component';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { SharedTestingModule } from '../../../../../../../../shared/shared.module';
import { ArtifactType, Clients } from '../../../../artifact'; // Import the necessary type
import { ArtifactFront } from '../../../../artifact';

describe('PullCommandComponent', () => {
    let component: PullCommandComponent;
    let fixture: ComponentFixture<PullCommandComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [PullCommandComponent],
            imports: [SharedTestingModule],
        }).compileComponents();

        fixture = TestBed.createComponent(PullCommandComponent);
        component = fixture.componentInstance;

        component.artifact = {
            type: ArtifactType.CHART,
            tagNumber: 1,
            digest: 'sha256@digest',
            tags: [{ name: '1.0.0' }],
        };

        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    const artifactTestCases = [
        {
            type: ArtifactType.IMAGE,
            method: (component: any, artifact: any) =>
                component.getPullCommandForRuntimeByTag(artifact),
        },
        {
            type: ArtifactType.CNAB,
            method: (component: any, artifact: any) =>
                component.getPullCommandForCNABByTag(artifact),
        },
        {
            type: ArtifactType.CHART,
            method: (component: any, artifact: any) =>
                component.getPullCommandForChart(artifact),
        },
    ];

    artifactTestCases.forEach(({ type, method }) => {
        it(`should not display pull command modal when tag is undefined for artifact type: ${type}`, async () => {
            // Arrange: mock artifact with missing tag data
            component.artifact = {
                type,
                tagNumber: undefined,
                digest: 'sha256@digest',
                tags: undefined,
            };

            const pullCommand = method(component, component.artifact);
            expect(pullCommand.length).toBe(0);

            fixture.detectChanges();
            await fixture.whenStable();

            const modal = fixture.nativeElement.querySelector(
                '#pullCommandForChart'
            );
            expect(modal).toBeFalsy();
        });
    });

    artifactTestCases.forEach(({ type, method }) => {
        it(`should not display pull command modal when no tag for artifact type: ${type}`, async () => {
            component.artifact = {
                type,
                tagNumber: 0,
                digest: 'sha256@digest',
                tags: [],
            };

            const pullCommand = method(component, component.artifact);
            expect(pullCommand.length).toBe(0);

            fixture.detectChanges();
            await fixture.whenStable();

            const modal = fixture.nativeElement.querySelector(
                '#pullCommandForChart'
            );
            expect(modal).toBeFalsy();
        });
    });

    it('should display when pull command for chart is available', async () => {
        // Mock the artifact input with a valid value
        component.artifact = {
            type: ArtifactType.CHART,
            tagNumber: 1,
            digest: 'sha256@digest',
            tags: [{ name: '1.0.0' }],
        };
        component.getPullCommandForChart(component.artifact);
        expect(
            component.getPullCommandForChart(component.artifact).length
        ).toBeGreaterThan(0);
        fixture.detectChanges();
        await fixture.whenStable();
        const modal =
            fixture.nativeElement.querySelector(`#pullCommandForChart`);
        expect(modal).toBeTruthy();
    });

    it('should display when pull command for digest is available', async () => {
        // Mock the artifact input with a valid value
        component.artifact = {
            type: ArtifactType.IMAGE,
        };
        component.getPullCommandForRuntimeByDigest(component.artifact);
        fixture.detectChanges();
        await fixture.whenStable();
        const modal = fixture.nativeElement.querySelector(
            `#pullCommandForDigest`
        );
        expect(modal).toBeTruthy();
    });

    it('should display when pull command for CNAB is available', async () => {
        // Mock the artifact input with a valid value
        component.artifact = {
            type: ArtifactType.CNAB,
        };
        component.getPullCommandForCNAB(component.artifact);
        fixture.detectChanges();
        await fixture.whenStable();
        const modal =
            fixture.nativeElement.querySelector(`#pullCommandForCNAB`);
        expect(modal).toBeTruthy();
    });
});
