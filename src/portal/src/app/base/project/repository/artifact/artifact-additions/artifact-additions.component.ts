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
import { ClrTabs } from '@clr/angular';

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

    @Input() currentTabLinkId: string = 'vulnerability';
    activeTab: string = null;

    @ViewChild('additionsTab') tabs: ClrTabs;
    constructor(private ref: ChangeDetectorRef) {}

    ngOnInit(): void {
        this.activeTab = this.tab;
    }

    ngAfterViewChecked() {
        if (this.activeTab) {
            this.currentTabLinkId = this.activeTab;
            this.activeTab = null;
        }
        this.ref.detectChanges();
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
    getSbom(): AdditionLink {
        if (this.additionLinks && this.additionLinks[ADDITIONS.SBOMS]) {
            return this.additionLinks[ADDITIONS.SBOMS];
        }
        return {};
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

    actionTab(tab: string): void {
        this.currentTabLinkId = tab;
    }
}
