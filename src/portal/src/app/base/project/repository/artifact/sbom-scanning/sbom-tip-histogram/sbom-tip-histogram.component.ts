import { Component, Input } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';
import { ActivatedRoute, Router } from '@angular/router';
import { ScannerVo, SbomSummary } from '../../../../../../shared/services';
import { SBOM_SCAN_STATUS } from '../../../../../../shared/units/utils';
import {
    UN_LOGGED_PARAM,
    YES,
} from '../../../../../../account/sign-in/sign-in.service';
import { HAS_STYLE_MODE, StyleMode } from '../../../../../../services/theme';
import { ScanTypes } from '../../../../../../shared/entities/shared.const';

const MIN = 60;
const MIN_STR = 'min ';
const SEC_STR = 'sec';
const SUCCESS_PCT: number = 100;

@Component({
    selector: 'hbr-sbom-tip-histogram',
    templateUrl: './sbom-tip-histogram.component.html',
    styleUrls: ['./sbom-tip-histogram.component.scss'],
})
export class SbomTipHistogramComponent {
    @Input() scanner: ScannerVo;
    @Input() sbomSummary: SbomSummary = {
        scan_status: SBOM_SCAN_STATUS.NOT_GENERATED_SBOM,
    };
    @Input() artifactDigest: string = '';
    @Input() sbomDigest: string = '';
    constructor(
        private translate: TranslateService,
        private activatedRoute: ActivatedRoute,
        private router: Router
    ) {}

    duration(): string {
        if (this.sbomSummary && this.sbomSummary.duration) {
            let str = '';
            const min = Math.floor(this.sbomSummary.duration / MIN);
            if (min) {
                str += min + ' ' + MIN_STR;
            }
            const sec = this.sbomSummary.duration % MIN;
            if (sec) {
                str += sec + ' ' + SEC_STR;
            }
            return str;
        }
        return '0';
    }

    public get completePercent(): string {
        return this.sbomSummary.scan_status === SBOM_SCAN_STATUS.SUCCESS
            ? `100%`
            : '0%';
    }
    isLimitedSuccess(): boolean {
        return (
            this.sbomSummary && this.sbomSummary.complete_percent < SUCCESS_PCT
        );
    }
    get completeTimestamp(): Date {
        return this.sbomSummary && this.sbomSummary.end_time
            ? this.sbomSummary.end_time
            : new Date();
    }

    get noSbom(): boolean {
        return (
            this.sbomSummary.scan_status === SBOM_SCAN_STATUS.NOT_GENERATED_SBOM
        );
    }

    isThemeLight() {
        return localStorage.getItem(HAS_STYLE_MODE) === StyleMode.LIGHT;
    }

    getScannerInfo(): string {
        if (this.scanner) {
            if (this.scanner.name && this.scanner.version) {
                return `${this.scanner.name}@${this.scanner.version}`;
            }
            if (this.scanner.name && !this.scanner.version) {
                return `${this.scanner.name}`;
            }
        }
        return '';
    }

    goIntoArtifactSbomSummaryPage(): void {
        const relativeRouterLink: string[] = ['artifacts', this.artifactDigest];
        if (this.activatedRoute.snapshot.queryParams[UN_LOGGED_PARAM] === YES) {
            this.router.navigate(relativeRouterLink, {
                relativeTo: this.activatedRoute,
                queryParams: {
                    [UN_LOGGED_PARAM]: YES,
                    sbomDigest: this.sbomDigest ?? '',
                    tab: ScanTypes.SBOM,
                },
            });
        } else {
            this.router.navigate(relativeRouterLink, {
                relativeTo: this.activatedRoute,
                queryParams: {
                    sbomDigest: this.sbomDigest ?? '',
                    tab: ScanTypes.SBOM,
                },
            });
        }
    }
}
