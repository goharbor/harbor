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
import { Component, Input, OnDestroy, OnInit } from '@angular/core';
import { Subscription, of, timer } from 'rxjs';
import { catchError, finalize, switchMap } from 'rxjs/operators';
import { TranslateService } from '@ngx-translate/core';
import { DockerfileService } from '../../../../../../../../ng-swagger-gen/services/dockerfile.service';
import { DockerfileOptimization } from '../../../../../../../../ng-swagger-gen/models/dockerfile-optimization';
import { ArtifactService } from '../../../../../../../../ng-swagger-gen/services/artifact.service';
import { Tag } from '../../../../../../../../ng-swagger-gen/models/tag';
import {
    dbEncodeURIComponent,
    downloadFile,
} from '../../../../../../shared/units/utils';
import { AppConfigService } from '../../../../../../services/app-config.service';
import { MessageHandlerService } from '../../../../../../shared/services/message-handler.service';

const POLL_INTERVAL_MS = 3000;
// Stop polling after ~10 minutes; the backend job has its own timeout that marks
// the record errored, so a stuck poll means something is genuinely wrong.
const MAX_POLLS = 200;

const STATUS_PENDING = 'Pending';
const STATUS_RUNNING = 'Running';
const STATUS_SUCCESS = 'Success';
const STATUS_ERROR = 'Error';

@Component({
    selector: 'hbr-dockerfile-optimization',
    templateUrl: './dockerfile-optimization.component.html',
    styleUrls: ['./dockerfile-optimization.component.scss'],
    standalone: false,
})
export class DockerfileOptimizationComponent implements OnInit, OnDestroy {
    @Input() projectName: string;
    @Input() repoName: string;
    @Input() digest: string;

    loading = false;
    loadingCached = false;
    loadingAvailability = false;
    inProgress = false;
    result: DockerfileOptimization = null;
    errorMessage: string = null;
    showButton = false;
    // Set once the availability check resolves; a configured-but-unreachable
    // optimizer surfaces here instead of only after a failed click.
    unavailableReason: string = null;

    // The optimized Dockerfile as edited by the user; seeded from the result
    // whenever a fresh Success record arrives (see handleRecord). Harbor can't
    // build this server-side (no access to the original build context), so it's
    // offered as a download + a copy-ready build/push command instead.
    editedDockerfile = '';
    registryUrl = '';
    // Fetched independently rather than taken from a parent-supplied artifact:
    // the artifact detail page resolves its Artifact with withTag: false, so
    // .tags is never populated there.
    tags: Tag[] = [];

    private pollSubscription: Subscription;
    private pollCount = 0;

    constructor(
        private dockerfileService: DockerfileService,
        private artifactService: ArtifactService,
        private appConfigService: AppConfigService,
        private msgHandler: MessageHandlerService,
        private translate: TranslateService
    ) {}

    ngOnInit(): void {
        this.registryUrl = this.appConfigService.getConfig()?.registry_url;
        this.loadingCached = true;
        this.loadingAvailability = true;

        // Runs in parallel with the cached-record lookup below; whichever
        // resolves last wins, so each handler re-applies showButton rather
        // than assuming the other has already run.
        this.dockerfileService
            .getDockerfileOptimizeAvailability({
                projectName: this.projectName,
                repositoryName: dbEncodeURIComponent(this.repoName),
                reference: this.digest,
            })
            .pipe(finalize(() => (this.loadingAvailability = false)))
            .subscribe({
                next: res => {
                    this.unavailableReason = res.available
                        ? null
                        : res.reason ||
                          'Dockerfile optimization is not available';
                    if (this.unavailableReason) {
                        this.showButton = false;
                    }
                },
                // Availability is a best-effort hint; on failure fall back to
                // showing the button and surfacing errors on click as before.
                error: () => (this.unavailableReason = null),
            });

        this.dockerfileService
            .getDockerfileOptimization({
                projectName: this.projectName,
                repositoryName: dbEncodeURIComponent(this.repoName),
                reference: this.digest,
            })
            .pipe(finalize(() => (this.loadingCached = false)))
            .subscribe({
                next: res => {
                    this.handleRecord(res);
                },
                error: err => {
                    if (err?.status === 404) {
                        this.showButton = !this.unavailableReason;
                    } else {
                        this.errorMessage =
                            err?.error?.errors?.[0]?.message ||
                            err?.message ||
                            'Failed to check for cached optimization';
                    }
                },
            });

        this.artifactService
            .listTags({
                projectName: this.projectName,
                repositoryName: dbEncodeURIComponent(this.repoName),
                reference: this.digest,
            })
            .subscribe({
                next: tags => (this.tags = tags || []),
                // No tags is a legitimate state (e.g. untagged/GC'd artifacts);
                // the Download/Build section just stays hidden via hasSourceTag.
                error: () => (this.tags = []),
            });
    }

    ngOnDestroy(): void {
        this.stopPolling();
    }

    optimize(): void {
        this.errorMessage = null;
        this.result = null;
        this.loading = true;
        this.dockerfileService
            .optimizeDockerfile({
                projectName: this.projectName,
                repositoryName: dbEncodeURIComponent(this.repoName),
                reference: this.digest,
            })
            .pipe(finalize(() => (this.loading = false)))
            .subscribe({
                next: res => {
                    this.showButton = false;
                    this.handleRecord(res);
                },
                error: err => {
                    this.errorMessage =
                        err?.error?.errors?.[0]?.message ||
                        err?.message ||
                        'Failed to start Dockerfile optimization';
                },
            });
    }

    // handleRecord interprets the record status: terminal states render, pending
    // and running states enter/continue the polling loop.
    private handleRecord(rec: DockerfileOptimization): void {
        if (rec.status === STATUS_PENDING || rec.status === STATUS_RUNNING) {
            this.inProgress = true;
            this.startPolling();
            return;
        }

        this.inProgress = false;
        this.stopPolling();

        if (rec.status === STATUS_ERROR) {
            this.errorMessage = rec.error || 'Optimization failed';
            this.showButton = !this.unavailableReason;
            return;
        }

        // Success, or a legacy record without a status field
        this.result = rec;
        this.editedDockerfile = rec.optimized_dockerfile || '';
        this.showButton = false;
    }

    // The tag this artifact was pushed under; the built-and-pushed image
    // reuses it with an "-opt" suffix. Falls back to the first tag if the
    // artifact carries more than one.
    get sourceTag(): string {
        return this.tags?.[0]?.name || '';
    }

    get hasSourceTag(): boolean {
        return !!this.sourceTag;
    }

    get buildCommand(): string {
        if (!this.hasSourceTag || !this.registryUrl) {
            return '';
        }
        const ref = `${this.registryUrl}/${this.projectName}/${this.repoName}:${this.sourceTag}-opt`;
        // Plain "docker build" can't emit provenance; buildx/BuildKit is
        // required. Attaching provenance keeps the pushed "-opt" image
        // eligible for the same Dockerfile-extraction path this feature
        // itself relies on, instead of falling back to history reconstruction.
        return (
            `docker buildx build --push --provenance=mode=max ` +
            `--attest type=provenance,mode=max -t ${ref} -f Dockerfile .`
        );
    }

    downloadDockerfile(): void {
        downloadFile({
            // Chrome appends an extension (e.g. ".txt") to extension-less
            // download filenames when the blob has a recognized text MIME
            // type; octet-stream avoids that sniffing so the file is saved
            // as plain "Dockerfile".
            data: new Blob([this.editedDockerfile], {
                type: 'application/octet-stream',
            }),
            filename: 'Dockerfile',
        });
    }

    onCpSuccess(): void {
        this.translate
            .get('REPOSITORY.COPY_SUCCESS', { param: this.buildCommand })
            .subscribe((res: string) => {
                this.msgHandler.showSuccess(res);
            });
    }

    onCpError(): void {
        this.msgHandler.showError('REPOSITORY.COPY_ERROR', {});
    }

    private startPolling(): void {
        if (this.pollSubscription) {
            return;
        }
        this.pollCount = 0;
        this.pollSubscription = timer(POLL_INTERVAL_MS, POLL_INTERVAL_MS)
            .pipe(
                switchMap(() =>
                    this.dockerfileService
                        .getDockerfileOptimization({
                            projectName: this.projectName,
                            repositoryName: dbEncodeURIComponent(this.repoName),
                            reference: this.digest,
                        })
                        // An error here must not reach the outer subscribe(): with
                        // switchMap, an inner-observable error terminates the whole
                        // subscription (no further timer ticks), so a single transient
                        // failure would otherwise kill polling forever.
                        .pipe(catchError(() => of(null)))
                )
            )
            .subscribe({
                next: res => {
                    this.pollCount++;
                    if (this.pollCount > MAX_POLLS) {
                        this.stopPolling();
                        this.inProgress = false;
                        this.errorMessage =
                            'Optimization is taking too long; please retry later';
                        this.showButton = !this.unavailableReason;
                        return;
                    }
                    // res is null when this tick's request failed; the next tick retries.
                    if (res) {
                        this.handleRecord(res);
                    }
                },
            });
    }

    private stopPolling(): void {
        if (this.pollSubscription) {
            this.pollSubscription.unsubscribe();
            this.pollSubscription = null;
        }
    }
}
