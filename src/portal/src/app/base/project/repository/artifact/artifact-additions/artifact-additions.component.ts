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
    AfterViewChecked,
    ChangeDetectorRef,
    Component,
    Input,
    OnInit,
    ViewChild,
} from '@angular/core';
import { ADDITIONS } from './models';
import { AdditionLinks } from '../../../../../../../ng-swagger-gen/models/addition-links';
import { AdditionLink } from '../../../../../../../ng-swagger-gen/models/addition-link';
import { Artifact } from '../../../../../../../ng-swagger-gen/models/artifact';
import { ClrLoadingState, ClrTabs } from '@clr/angular';
import { ArtifactListPageService } from '../artifact-list-page/artifact-list-page.service';

@Component({
    selector: 'artifact-additions',
    templateUrl: './artifact-additions.component.html',
    styleUrls: ['./artifact-additions.component.scss'],
})
export class ArtifactAdditionsComponent implements AfterViewChecked, OnInit {
    @Input() artifact: Artifact;
    @Input() additionLinks: AdditionLinks;
    @Input() projectName: string;
    @Input()
    projectId: number;
    @Input()
    repoName: string;
    @Input()
    digest: string;
    @Input()
    sbomDigest: string;
    @Input()
    tab: string;

    @Input() currentTabLinkId: string = '';
    activeTab: string = null;

    @ViewChild('additionsTab') tabs: ClrTabs;
    constructor(
        private ref: ChangeDetectorRef,
        private artifactListPageService: ArtifactListPageService
    ) {}

    ngOnInit(): void {
        this.activeTab = this.tab;

        if (!this.activeTab && this.additionLinks[ADDITIONS.VULNERABILITIES]) {
            this.currentTabLinkId = 'vulnerability';
        } else if (!this.activeTab && this.additionLinks[ADDITIONS.SUMMARY]) {
            this.currentTabLinkId = 'summary-link';
        }

        this.artifactListPageService.init(this.projectId);
    }

    ngAfterViewChecked() {
        if (this.activeTab) {
            this.currentTabLinkId = this.activeTab;
            this.activeTab = null;
        }
        this.ref.detectChanges();
    }

    hasScannerSupportSBOM(): boolean {
        if (this.additionLinks && this.additionLinks[ADDITIONS.SBOMS]) {
            return true;
        }
        return false;
    }

    getVulnerability(): AdditionLink {
        if (
            this.additionLinks &&
            this.additionLinks[ADDITIONS.VULNERABILITIES]
        ) {
            return this.additionLinks[ADDITIONS.VULNERABILITIES];
        }
        return null;
    }

    getBuildHistory(): AdditionLink {
        if (this.additionLinks && this.additionLinks[ADDITIONS.BUILD_HISTORY]) {
            return this.additionLinks[ADDITIONS.BUILD_HISTORY];
        }
        return null;
    }
    getSummary(): AdditionLink {
        if (this.additionLinks && this.additionLinks[ADDITIONS.SUMMARY]) {
            return this.additionLinks[ADDITIONS.SUMMARY];
        }
        return null;
    }
    getDependencies(): AdditionLink {
        if (this.additionLinks && this.additionLinks[ADDITIONS.DEPENDENCIES]) {
            return this.additionLinks[ADDITIONS.DEPENDENCIES];
        }
        return null;
    }
    getValues(): AdditionLink {
        if (this.additionLinks && this.additionLinks[ADDITIONS.VALUES]) {
            return this.additionLinks[ADDITIONS.VALUES];
        }
        return null;
    }

    getFile(): AdditionLink {
        if (this.additionLinks && this.additionLinks[ADDITIONS.FILES]) {
            return this.additionLinks[ADDITIONS.FILES];
        }
        return null;
    }

    getLicense(): AdditionLink {
        if (this.additionLinks && this.additionLinks[ADDITIONS.LICENSE]) {
            return this.additionLinks[ADDITIONS.LICENSE];
        }
        return null;
    }

    actionTab(tab: string): void {
        this.currentTabLinkId = tab;
    }

    getScanBtnState(): ClrLoadingState {
        return this.artifactListPageService.getScanBtnState();
    }

    hasEnabledScanner(): boolean {
        return this.artifactListPageService.hasEnabledScanner();
    }
}
