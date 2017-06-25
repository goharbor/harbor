import { Component, Input, Output, EventEmitter, OnInit } from '@angular/core';

import { TAG_DETAIL_STYLES } from './tag-detail.component.css';
import { TAG_DETAIL_HTML } from './tag-detail.component.html';

import { TagService, Tag, VulnerabilitySeverity } from '../service/index';
import { toPromise } from '../utils';
import { ErrorHandler } from '../error-handler/index';

@Component({
    selector: 'hbr-tag-detail',
    styles: [TAG_DETAIL_STYLES],
    template: TAG_DETAIL_HTML,

    providers: []
})
export class TagDetailComponent implements OnInit {
    _highCount: number = 0;
    _mediumCount: number = 0;
    _lowCount: number = 0;
    _unknownCount: number = 0;

    @Input() tagId: string;
    @Input() repositoryId: string;
    tagDetails: Tag = {
        name: "--",
        author: "--",
        created: new Date(),
        architecture: "--",
        os: "--",
        docker_version: "--",
        digest: "--"
    };

    @Output() backEvt: EventEmitter<any> = new EventEmitter<any>();

    constructor(
        private tagService: TagService,
        private errorHandler: ErrorHandler) { }

    ngOnInit(): void {
        if (this.repositoryId && this.tagId) {
            toPromise<Tag>(this.tagService.getTag(this.repositoryId, this.tagId))
                .then(response => {
                    this.tagDetails = response;
                    if (this.tagDetails &&
                        this.tagDetails.scan_overview &&
                        this.tagDetails.scan_overview.components &&
                        this.tagDetails.scan_overview.components.summary) {
                        this.tagDetails.scan_overview.components.summary.forEach(item => {
                            switch (item.severity) {
                                case VulnerabilitySeverity.UNKNOWN:
                                    this._unknownCount += item.count;
                                    break;
                                case VulnerabilitySeverity.LOW:
                                    this._lowCount += item.count;
                                    break;
                                case VulnerabilitySeverity.MEDIUM:
                                    this._mediumCount += item.count;
                                    break;
                                case VulnerabilitySeverity.HIGH:
                                    this._highCount += item.count;
                                    break;
                                default:
                                    break;
                            }
                        });
                    }
                })
                .catch(error => this.errorHandler.error(error))
        }
    }

    onBack(): void {
        this.backEvt.emit(this.tagId);
    }

    getPackageText(count: number): string {
        return count > 1 ? "VULNERABILITY.PACKAGES" : "VULNERABILITY.PACKAGE";
    }

    public get author(): string {
        return this.tagDetails && this.tagDetails.author ? this.tagDetails.author : 'TAG.ANONYMITY';
    }

    public get highCount(): number {
        return this._highCount;
    }

    public get mediumCount(): number {
        return this._mediumCount;
    }

    public get lowCount(): number {
        return this._lowCount;
    }

    public get unknownCount(): number {
        return this._unknownCount;
    }

    public get scanCompletedDatetime(): Date {
        return this.tagDetails && this.tagDetails.scan_overview ?
            this.tagDetails.scan_overview.update_time : new Date();
    }

    public get suffixForHigh(): string {
        return this.highCount > 1 ? "VULNERABILITY.PLURAL" : "VULNERABILITY.SINGULAR";
    }

    public get suffixForMedium(): string {
        return this.mediumCount > 1 ? "VULNERABILITY.PLURAL" : "VULNERABILITY.SINGULAR";
    }

    public get suffixForLow(): string {
        return this.lowCount > 1 ? "VULNERABILITY.PLURAL" : "VULNERABILITY.SINGULAR";
    }

    public get suffixForUnknown(): string {
        return this.unknownCount > 1 ? "VULNERABILITY.PLURAL" : "VULNERABILITY.SINGULAR";
    }
}
