import { Component, Input, Output, EventEmitter, OnInit } from '@angular/core';

import { TAG_DETAIL_STYLES } from './tag-detail.component.css';
import { TAG_DETAIL_HTML } from './tag-detail.component.html';

import { TagService, Tag } from '../service/index';
import { toPromise } from '../utils';
import { ErrorHandler } from '../error-handler/index';

@Component({
    selector: 'hbr-tag-detail',
    styles: [TAG_DETAIL_STYLES],
    template: TAG_DETAIL_HTML,

    providers: []
})
export class TagDetailComponent implements OnInit {
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
                .then(response => this.tagDetails = response)
                .catch(error => this.errorHandler.error(error))
        }
    }

    onBack(): void {
        this.backEvt.emit(this.tagId);
    }

    public get highCount(): number {
        return this.tagDetails && this.tagDetails.vulnerability ?
            this.tagDetails.vulnerability.package_with_high : 0;
    }

    public get mediumCount(): number {
        return this.tagDetails && this.tagDetails.vulnerability ?
            this.tagDetails.vulnerability.package_with_medium : 0;
    }

    public get lowCount(): number {
        return this.tagDetails && this.tagDetails.vulnerability ?
            this.tagDetails.vulnerability.package_With_low : 0;
    }

    public get unknownCount(): number {
        return this.tagDetails && this.tagDetails.vulnerability ?
            this.tagDetails.vulnerability.package_with_unknown : 0;
    }

    public get scanCompletedDatetime(): Date {
        return this.tagDetails && this.tagDetails.vulnerability ?
            this.tagDetails.vulnerability.complete_timestamp : new Date();
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
