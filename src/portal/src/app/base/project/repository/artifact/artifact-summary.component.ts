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
import { Component, EventEmitter, OnInit, Output } from '@angular/core';
import { Artifact } from '../../../../../../ng-swagger-gen/models/artifact';
import { Label } from '../../../../../../ng-swagger-gen/models/label';
import { ActivatedRoute, Router } from '@angular/router';
import { Project } from '../../project';
import { artifactDefault } from './artifact';
import { SafeUrl } from '@angular/platform-browser';
import { ArtifactService } from './artifact.service';
import {
    EventService,
    HarborEvent,
} from '../../../../services/event-service/event.service';
import { ModelImportService } from '../model-import/model-import.service';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import { finalize } from 'rxjs/operators';

interface HuggingFaceArtifactAttrs {
    repo_id: string;
    revision?: string;
    commit_sha?: string;
}

@Component({
    selector: 'artifact-summary',
    templateUrl: './artifact-summary.component.html',
    styleUrls: ['./artifact-summary.component.scss'],
    providers: [],
    standalone: false,
})
export class ArtifactSummaryComponent implements OnInit {
    tagId: string;
    artifactDigest: string;
    sbomDigest?: string;
    activeTab?: string;
    repositoryName: string;
    projectId: string | number;
    referArtifactNameArray: string[] = [];

    labels: Label;
    artifact: Artifact;
    @Output()
    backEvt: EventEmitter<any> = new EventEmitter<any>();
    projectName: string;
    isProxyCacheProject: boolean = false;
    loading: boolean = false;
    syncingFromHuggingFace = false;

    constructor(
        private route: ActivatedRoute,
        private router: Router,
        private frontEndArtifactService: ArtifactService,
        private event: EventService,
        private modelImportService: ModelImportService,
        private errorHandler: ErrorHandler
    ) {}

    goBack(): void {
        this.router.navigate([
            'harbor',
            'projects',
            this.projectId,
            'repositories',
            this.repositoryName,
        ]);
    }

    goBackRep(): void {
        this.router.navigate([
            'harbor',
            'projects',
            this.projectId,
            'repositories',
        ]);
    }

    goBackPro(): void {
        this.router.navigate(['harbor', 'projects']);
    }
    jumpDigest(index: number) {
        const arr: string[] = this.referArtifactNameArray.slice(0, index + 1);
        if (arr && arr.length) {
            this.router.navigate([
                'harbor',
                'projects',
                this.projectId,
                'repositories',
                this.repositoryName,
                'artifacts-tab',
                'depth',
                arr.join('-'),
            ]);
        } else {
            this.router.navigate([
                'harbor',
                'projects',
                this.projectId,
                'repositories',
                this.repositoryName,
            ]);
        }
    }

    ngOnInit(): void {
        let depth = this.route.snapshot.params['depth'];
        if (depth) {
            this.referArtifactNameArray = depth.split('-');
        }
        this.repositoryName = this.route.snapshot.params['repo'];
        this.artifactDigest = this.route.snapshot.params['digest'];
        this.projectId = this.route.snapshot.parent.params['id'];
        this.sbomDigest = this.route.snapshot.queryParams['sbomDigest'];
        this.activeTab = this.route.snapshot.queryParams['tab'];
        if (this.repositoryName && this.artifactDigest) {
            const resolverData = this.route.snapshot.data;
            if (resolverData) {
                const pro: Project = <Project>(
                    resolverData['artifactResolver'][1]
                );
                this.projectName = pro.name;
                if (pro.registry_id) {
                    this.isProxyCacheProject = true;
                }
                this.artifact = <Artifact>resolverData['artifactResolver'][0];
                this.getIconFromBackEnd();
            }
        }
        // scroll to the top for harbor container HTML element
        this.event.publish(HarborEvent.SCROLL_TO_POSITION, 0);
    }
    onBack(): void {
        this.backEvt.emit(this.repositoryName);
    }
    showDefaultIcon(event: any) {
        if (event && event.target) {
            event.target.src = artifactDefault;
        }
    }
    getIcon(icon: string): SafeUrl {
        return this.frontEndArtifactService.getIcon(icon);
    }
    getIconFromBackEnd() {
        if (this.artifact && this.artifact.icon) {
            this.frontEndArtifactService.getIconsFromBackEnd([this.artifact]);
        }
    }

    canSyncFromHuggingFace(): boolean {
        const hf = this.huggingFaceAttrs();
        return (
            this.artifact?.type === 'CNAI' &&
            !!hf?.repo_id
        );
    }

    syncFromHuggingFace(): void {
        const hf = this.huggingFaceAttrs();
        if (!this.canSyncFromHuggingFace() || !hf) {
            return;
        }
        const tag = this.firstTagName() || 'latest';
        this.syncingFromHuggingFace = true;
        this.modelImportService
            .runImport(this.projectName, {
                target_repository: `${this.projectName}/${this.repositoryName}`,
                target_tag: tag,
                hf_repo_id: hf.repo_id,
                hf_revision: hf.revision || hf.commit_sha || 'main',
                last_synced_commit: hf.commit_sha,
            })
            .pipe(
                finalize(() => (this.syncingFromHuggingFace = false))
            )
            .subscribe({
                next: () => this.event.publish(HarborEvent.SCROLL_TO_POSITION, 0),
                error: err => this.errorHandler.error(err),
            });
    }

    private firstTagName(): string {
        const tags = (this.artifact as any)?.tags;
        if (tags && tags.length) {
            return tags[0].name;
        }
        return '';
    }

    private huggingFaceAttrs(): HuggingFaceArtifactAttrs | undefined {
        return (
            this.artifact?.extra_attrs as
                | { huggingface?: HuggingFaceArtifactAttrs }
                | undefined
        )?.huggingface;
    }
}
