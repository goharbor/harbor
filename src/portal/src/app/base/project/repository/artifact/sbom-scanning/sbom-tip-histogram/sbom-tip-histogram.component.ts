import { Component, Input } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { SbomSummary } from '../../../../../../shared/services';
import { SBOM_SCAN_STATUS } from '../../../../../../shared/units/utils';
import { HAS_STYLE_MODE, StyleMode } from '../../../../../../services/theme';
import { ScanTypes } from '../../../../../../shared/entities/shared.const';
import { Scanner } from '../../../../../left-side-nav/interrogation-services/scanner/scanner';
import { Accessory } from 'ng-swagger-gen/models';
import { AccessoryType } from '../../artifact';

const MIN = 60;
const MIN_STR = 'min ';
const SEC_STR = 'sec';

@Component({
    selector: 'hbr-sbom-tip-histogram',
    templateUrl: './sbom-tip-histogram.component.html',
    styleUrls: ['./sbom-tip-histogram.component.scss'],
})
export class SbomTipHistogramComponent {
    @Input() scanner: Scanner;
    @Input() sbomSummary: SbomSummary = {
        scan_status: SBOM_SCAN_STATUS.NOT_GENERATED_SBOM,
    };
    @Input() artifactDigest: string = '';
    @Input() sbomDigest: string = '';
    @Input() accessories: Accessory[] = [];
    constructor(
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

    public getSbomAccessories(): Accessory[] {
        return (
            this.accessories?.filter(
                accessory => accessory.type === AccessoryType.SBOM
            ) ?? []
        );
    }

    public get completePercent(): string {
        return this.sbomSummary.scan_status === SBOM_SCAN_STATUS.SUCCESS
            ? `100%`
            : '0%';
    }

    get completeTimestamp(): Date {
        return this.sbomSummary && this.sbomSummary.end_time
            ? this.sbomSummary.end_time
            : new Date();
    }

    showSbomDetailLink(): boolean {
        return this.sbomDigest && this.getSbomAccessories().length > 0;
    }

    showNoSbom(): boolean {
        return !this.sbomDigest || this.getSbomAccessories().length === 0;
    }

    showTooltip() {
        return (
            !this.sbomSummary ||
            !(
                this.sbomSummary &&
                this.sbomSummary.scan_status !== SBOM_SCAN_STATUS.SUCCESS
            )
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
        this.router.navigate(relativeRouterLink, {
            relativeTo: this.activatedRoute,
            queryParams: {
                sbomDigest: this.sbomDigest ?? '',
                tab: ScanTypes.SBOM,
            },
        });
    }
}
