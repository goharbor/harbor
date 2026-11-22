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
import {
    ComponentFixture,
    TestBed,
    fakeAsync,
    tick,
} from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { delay } from 'rxjs/operators';
import { DockerfileOptimizationComponent } from './dockerfile-optimization.component';
import { SharedTestingModule } from '../../../../../../shared/shared.module';
import { DockerfileService } from '../../../../../../../../ng-swagger-gen/services/dockerfile.service';
import { DockerfileOptimization } from '../../../../../../../../ng-swagger-gen/models/dockerfile-optimization';
import { ArtifactService } from '../../../../../../../../ng-swagger-gen/services/artifact.service';
import { Tag } from '../../../../../../../../ng-swagger-gen/models/tag';
import { AppConfigService } from '../../../../../../services/app-config.service';

describe('DockerfileOptimizationComponent', () => {
    let component: DockerfileOptimizationComponent;
    let fixture: ComponentFixture<DockerfileOptimizationComponent>;
    let fakedDockerfileService: {
        getDockerfileOptimization: jasmine.Spy;
        getDockerfileOptimizeAvailability: jasmine.Spy;
        optimizeDockerfile: jasmine.Spy;
    };
    let fakedArtifactService: {
        listTags: jasmine.Spy;
    };

    const pendingRecord: DockerfileOptimization = { status: 'Pending' };
    const runningRecord: DockerfileOptimization = { status: 'Running' };
    const successRecord: DockerfileOptimization = {
        status: 'Success',
        dockerfile: 'FROM scratch\n',
        optimized_dockerfile: 'FROM alpine:3.21\n',
    };
    const errorRecord: DockerfileOptimization = {
        status: 'Error',
        error: 'no attestation found',
    };

    const oneTag: Tag[] = [{ name: 'v1' }];

    function createComponent(): void {
        fixture = TestBed.createComponent(DockerfileOptimizationComponent);
        component = fixture.componentInstance;
        component.projectName = 'library';
        component.repoName = 'library/photon';
        component.digest = 'sha256:artifact';
    }

    beforeEach(() => {
        fakedDockerfileService = {
            getDockerfileOptimization: jasmine
                .createSpy('getDockerfileOptimization')
                .and.returnValue(throwError({ status: 404 })),
            getDockerfileOptimizeAvailability: jasmine
                .createSpy('getDockerfileOptimizeAvailability')
                .and.returnValue(of({ available: true })),
            optimizeDockerfile: jasmine.createSpy('optimizeDockerfile'),
        };
        fakedArtifactService = {
            listTags: jasmine.createSpy('listTags').and.returnValue(of([])),
        };

        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [DockerfileOptimizationComponent],
            providers: [
                {
                    provide: DockerfileService,
                    useValue: fakedDockerfileService,
                },
                {
                    provide: ArtifactService,
                    useValue: fakedArtifactService,
                },
                {
                    provide: AppConfigService,
                    useValue: {
                        getConfig: () => ({
                            registry_url: 'demo.harbor.io',
                        }),
                    },
                },
            ],
        });
    });

    afterEach(() => {
        if (component) {
            component.ngOnDestroy();
        }
    });

    it('should create and show the button when no cached record exists (404)', () => {
        createComponent();
        fixture.detectChanges();

        expect(component).toBeTruthy();
        expect(component.showButton).toBeTrue();
        expect(component.loadingCached).toBeFalse();
        expect(component.errorMessage).toBeNull();
    });

    it('should hide the button and show the reason when no optimizer is available', () => {
        fakedDockerfileService.getDockerfileOptimizeAvailability.and.returnValue(
            of({
                available: false,
                reason: 'no optimizer adapter is configured',
            })
        );
        createComponent();
        fixture.detectChanges();

        expect(component.showButton).toBeFalse();
        expect(component.unavailableReason).toBe(
            'no optimizer adapter is configured'
        );
    });

    it('should fall back to an empty-reason message when unavailable with no reason given', () => {
        fakedDockerfileService.getDockerfileOptimizeAvailability.and.returnValue(
            of({ available: false })
        );
        createComponent();
        fixture.detectChanges();

        expect(component.showButton).toBeFalse();
        expect(component.unavailableReason).toBe(
            'Dockerfile optimization is not available'
        );
    });

    it('should still show the button when the availability check itself fails', () => {
        fakedDockerfileService.getDockerfileOptimizeAvailability.and.returnValue(
            throwError({ status: 500 })
        );
        createComponent();
        fixture.detectChanges();

        expect(component.showButton).toBeTrue();
        expect(component.unavailableReason).toBeNull();
    });

    it('should hide an already-shown button once a slower availability check reports unavailable', fakeAsync(() => {
        // Cached lookup (404) resolves synchronously above, but availability
        // resolves on a later tick -- the button must be retracted once it does.
        fakedDockerfileService.getDockerfileOptimizeAvailability.and.returnValue(
            of({
                available: false,
                reason: 'optimizer Sleeko is deactivated',
            }).pipe(delay(10))
        );
        createComponent();
        fixture.detectChanges();

        expect(component.showButton).toBeTrue();

        tick(10);
        expect(component.showButton).toBeFalse();
        expect(component.unavailableReason).toBe(
            'optimizer Sleeko is deactivated'
        );
    }));

    it('should surface a non-404 error from the initial cached lookup', () => {
        fakedDockerfileService.getDockerfileOptimization.and.returnValue(
            throwError({ error: { errors: [{ message: 'boom' }] } })
        );
        createComponent();
        fixture.detectChanges();

        expect(component.errorMessage).toBe('boom');
        expect(component.showButton).toBeFalse();
    });

    it('should render a cached Success record without polling', () => {
        fakedDockerfileService.getDockerfileOptimization.and.returnValues(
            of(successRecord)
        );
        createComponent();
        fixture.detectChanges();

        expect(component.result).toEqual(successRecord);
        expect(component.inProgress).toBeFalse();
        expect(component.showButton).toBeFalse();
        expect(
            fakedDockerfileService.getDockerfileOptimization
        ).toHaveBeenCalledTimes(1);
    });

    it('should render a cached Error record and offer the retry button', () => {
        fakedDockerfileService.getDockerfileOptimization.and.returnValues(
            of(errorRecord)
        );
        createComponent();
        fixture.detectChanges();

        expect(component.errorMessage).toBe('no attestation found');
        expect(component.showButton).toBeTrue();
        expect(component.result).toBeNull();
    });

    it('optimize() should start a fresh optimization and render the result', () => {
        fakedDockerfileService.optimizeDockerfile.and.returnValue(
            of(successRecord)
        );
        createComponent();
        fixture.detectChanges(); // triggers the initial (404) cached lookup

        component.optimize();

        expect(component.result).toEqual(successRecord);
        expect(component.loading).toBeFalse();
        expect(component.showButton).toBeFalse();
    });

    it('optimize() should surface an error message on failure', () => {
        fakedDockerfileService.optimizeDockerfile.and.returnValue(
            throwError({
                error: { errors: [{ message: 'adapter unreachable' }] },
            })
        );
        createComponent();
        fixture.detectChanges();

        component.optimize();

        expect(component.errorMessage).toBe('adapter unreachable');
        expect(component.loading).toBeFalse();
    });

    it('should poll while Pending/Running and stop once a terminal status arrives', fakeAsync(() => {
        fakedDockerfileService.optimizeDockerfile.and.returnValue(
            of(pendingRecord)
        );
        createComponent();
        fixture.detectChanges();

        component.optimize();
        expect(component.inProgress).toBeTrue();

        // First poll tick: still running.
        fakedDockerfileService.getDockerfileOptimization.and.returnValue(
            of(runningRecord)
        );
        tick(3000);
        expect(component.inProgress).toBeTrue();
        expect(component.result).toBeNull();

        // Second poll tick: terminal Success.
        fakedDockerfileService.getDockerfileOptimization.and.returnValue(
            of(successRecord)
        );
        tick(3000);
        expect(component.inProgress).toBeFalse();
        expect(component.result).toEqual(successRecord);

        // Polling must have stopped: no further calls even after more time passes.
        const callsSoFar =
            fakedDockerfileService.getDockerfileOptimization.calls.count();
        tick(3000);
        expect(
            fakedDockerfileService.getDockerfileOptimization.calls.count()
        ).toBe(callsSoFar);

        component.ngOnDestroy();
        tick(10000);
    }));

    it('should ignore transient polling errors and keep polling', fakeAsync(() => {
        fakedDockerfileService.optimizeDockerfile.and.returnValue(
            of(pendingRecord)
        );
        createComponent();
        fixture.detectChanges();
        component.optimize();

        fakedDockerfileService.getDockerfileOptimization.and.returnValue(
            throwError({ status: 500 })
        );
        tick(3000);
        expect(component.inProgress).toBeTrue();
        expect(component.errorMessage).toBeNull();

        fakedDockerfileService.getDockerfileOptimization.and.returnValue(
            of(successRecord)
        );
        tick(3000);
        expect(component.inProgress).toBeFalse();
        expect(component.result).toEqual(successRecord);

        component.ngOnDestroy();
        tick(10000);
    }));

    it('should give up and show a timeout message after MAX_POLLS', fakeAsync(() => {
        fakedDockerfileService.optimizeDockerfile.and.returnValue(
            of(pendingRecord)
        );
        createComponent();
        fixture.detectChanges();
        component.optimize();

        fakedDockerfileService.getDockerfileOptimization.and.returnValue(
            of(runningRecord)
        );
        // MAX_POLLS is 200; the 201st tick trips the timeout guard.
        tick(3000 * 201);

        expect(component.inProgress).toBeFalse();
        expect(component.errorMessage).toBe(
            'Optimization is taking too long; please retry later'
        );
        expect(component.showButton).toBeTrue();

        component.ngOnDestroy();
    }));

    it('ngOnDestroy should unsubscribe from an in-flight poll', fakeAsync(() => {
        fakedDockerfileService.optimizeDockerfile.and.returnValue(
            of(pendingRecord)
        );
        createComponent();
        fixture.detectChanges();
        component.optimize();

        fakedDockerfileService.getDockerfileOptimization.and.returnValue(
            of(runningRecord)
        );
        component.ngOnDestroy();

        const callsBefore =
            fakedDockerfileService.getDockerfileOptimization.calls.count();
        tick(10000);
        expect(
            fakedDockerfileService.getDockerfileOptimization.calls.count()
        ).toBe(callsBefore);
    }));

    it('should seed editedDockerfile from a Success record', () => {
        fakedDockerfileService.getDockerfileOptimization.and.returnValue(
            of(successRecord)
        );
        createComponent();
        fixture.detectChanges();

        expect(component.editedDockerfile).toBe(
            successRecord.optimized_dockerfile
        );
    });

    it('should leave tags empty if the listTags lookup fails', () => {
        fakedArtifactService.listTags.and.returnValue(
            throwError({ status: 500 })
        );
        createComponent();
        fixture.detectChanges();

        expect(component.tags).toEqual([]);
        expect(component.hasSourceTag).toBeFalse();
    });

    describe('sourceTag / hasSourceTag / buildCommand', () => {
        it('should all be empty/false before any tags have loaded (default state)', () => {
            // listTags resolves synchronously via `of()` in these tests, but the
            // component's `tags` field starts as [] regardless of the response,
            // so this also covers the "still loading" window in real usage.
            fakedArtifactService.listTags.and.returnValue(of([]));
            createComponent();
            fixture.detectChanges();

            expect(component.tags).toEqual([]);
            expect(component.sourceTag).toBe('');
            expect(component.hasSourceTag).toBeFalse();
            expect(component.buildCommand).toBe('');
        });

        it('should treat a tag with an empty-string name as no tag', () => {
            fakedArtifactService.listTags.and.returnValue(of([{ name: '' }]));
            createComponent();
            fixture.detectChanges();

            expect(component.sourceTag).toBe('');
            expect(component.hasSourceTag).toBeFalse();
            expect(component.buildCommand).toBe('');
        });

        it('should be empty when the artifact has a tag but registryUrl is not configured', () => {
            fakedArtifactService.listTags.and.returnValue(of(oneTag));
            TestBed.overrideProvider(AppConfigService, {
                useValue: { getConfig: () => ({ registry_url: '' }) },
            });
            createComponent();
            fixture.detectChanges();

            expect(component.hasSourceTag).toBeTrue();
            expect(component.registryUrl).toBe('');
            expect(component.buildCommand).toBe('');
        });

        it('should build a buildx command with provenance/attest flags for a single-tag artifact', () => {
            fakedArtifactService.listTags.and.returnValue(of(oneTag));
            createComponent();
            fixture.detectChanges();

            expect(component.sourceTag).toBe('v1');
            expect(component.hasSourceTag).toBeTrue();

            const cmd = component.buildCommand;
            const expectedRef = 'demo.harbor.io/library/library/photon:v1-opt';
            expect(cmd).toBe(
                'docker buildx build --push --provenance=mode=max ' +
                    '--attest type=provenance,mode=max ' +
                    `-t ${expectedRef} -f Dockerfile .`
            );
            // Structural checks independent of exact flag ordering/wording,
            // so this still catches regressions if the string is reworded.
            expect(cmd).toMatch(/^docker buildx build\b/);
            expect(cmd).toContain('--push');
            expect(cmd).toContain('--provenance=mode=max');
            expect(cmd).toContain('--attest type=provenance,mode=max');
            expect(cmd).toContain(`-t ${expectedRef}`);
            expect(cmd.endsWith('-f Dockerfile .')).toBeTrue();
            // Plain "docker build" can't emit provenance; a naive edit that
            // reverts to it should fail this.
            expect(cmd).not.toMatch(/^docker build\s/);
        });

        it('should use the first tag when the artifact has multiple tags', () => {
            fakedArtifactService.listTags.and.returnValue(
                of([{ name: 'v2' }, { name: 'v1' }, { name: 'latest' }])
            );
            createComponent();
            fixture.detectChanges();

            expect(component.sourceTag).toBe('v2');
            expect(component.buildCommand).toContain(
                'library/library/photon:v2-opt'
            );
            expect(component.buildCommand).not.toContain('v1-opt');
            expect(component.buildCommand).not.toContain('latest-opt');
        });

        it('should preserve a nested repository name verbatim in the target ref', () => {
            fakedArtifactService.listTags.and.returnValue(of(oneTag));
            createComponent();
            component.repoName = 'library/nested/photon';
            fixture.detectChanges();

            expect(component.buildCommand).toContain(
                '-t demo.harbor.io/library/library/nested/photon:v1-opt'
            );
        });
    });

    it('downloadDockerfile should trigger a browser download of the edited text', () => {
        fakedDockerfileService.getDockerfileOptimization.and.returnValue(
            of(successRecord)
        );
        createComponent();
        fixture.detectChanges();
        component.editedDockerfile = 'FROM alpine:3.21\nRUN echo hi\n';

        const createObjectURLSpy = spyOn(
            URL,
            'createObjectURL'
        ).and.returnValue('blob:fake-url');
        const revokeObjectURLSpy = spyOn(URL, 'revokeObjectURL');

        component.downloadDockerfile();

        expect(createObjectURLSpy).toHaveBeenCalledTimes(1);
        const blobArg = createObjectURLSpy.calls.mostRecent().args[0] as Blob;
        expect(blobArg.type).toBe('application/octet-stream');
        expect(revokeObjectURLSpy).toHaveBeenCalledWith('blob:fake-url');
    });
});
