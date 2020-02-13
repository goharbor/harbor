import { Component, Input, OnInit } from '@angular/core';
import { AdditionLink } from "../../../../../../ng-swagger-gen/models/addition-link";
import { ErrorHandler } from "../../../../utils/error-handler";
import { AdditionsService } from "../additions.service";
import {
  ScanningResultService,
  VulnerabilityItem
} from "../../../../services";
import { ClrDatagridComparatorInterface, ClrLoadingState } from "@clr/angular";
import { SEVERITY_LEVEL_MAP, VULNERABILITY_SEVERITY } from "../../../../utils/utils";
import { ChannelService } from "../../../../services/channel.service";
import { finalize } from "rxjs/operators";

@Component({
  selector: 'hbr-artifact-vulnerabilities',
  templateUrl: './artifact-vulnerabilities.component.html',
  styleUrls: ['./artifact-vulnerabilities.component.scss']
})
export class ArtifactVulnerabilitiesComponent implements OnInit {
  @Input()
  vulnerabilitiesLink: AdditionLink;

  scanningResults: VulnerabilityItem[] = [];
  loading: boolean = false;
  shouldShowLoading: boolean = true;
  hasEnabledScanner: boolean = false;
  scanBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
  severitySort: ClrDatagridComparatorInterface<VulnerabilityItem>;

  constructor(
    private errorHandler: ErrorHandler,
    private additionsService: AdditionsService,
    private scanningService: ScanningResultService,
    private channel: ChannelService,
  ) {
    const that = this;
    this.severitySort = {
      compare(a: VulnerabilityItem, b: VulnerabilityItem): number {
        return that.getLevel(a) - that.getLevel(b);
      }
    };
  }

  ngOnInit() {
    this.getVulnerabilities();
  }

  getVulnerabilities() {
    if (this.vulnerabilitiesLink
      && !this.vulnerabilitiesLink.absolute
      && this.vulnerabilitiesLink.href) {
      // only show loading for one time
      if (this.shouldShowLoading) {
        this.loading = true;
        this.shouldShowLoading = false;
      }
      this.additionsService.getDetailByLink(this.vulnerabilitiesLink.href)
        .pipe(finalize(() => this.loading = false))
        .subscribe(
        res => {
          this.scanningResults = res;
        }, error => {
          this.errorHandler.error(error);
        }
      );
    }
  }

  getLevel(v: VulnerabilityItem): number {
    if (v && v.severity && SEVERITY_LEVEL_MAP[v.severity]) {
      return SEVERITY_LEVEL_MAP[v.severity];
    }
    return 0;
  }
  refresh(): void {
    this.getVulnerabilities();
  }

  severityText(severity: string): string {
    switch (severity) {
      case VULNERABILITY_SEVERITY.CRITICAL:
        return 'VULNERABILITY.SEVERITY.CRITICAL';
      case VULNERABILITY_SEVERITY.HIGH:
        return 'VULNERABILITY.SEVERITY.HIGH';
      case VULNERABILITY_SEVERITY.MEDIUM:
        return 'VULNERABILITY.SEVERITY.MEDIUM';
      case VULNERABILITY_SEVERITY.LOW:
        return 'VULNERABILITY.SEVERITY.LOW';
      case VULNERABILITY_SEVERITY.NEGLIGIBLE:
        return 'VULNERABILITY.SEVERITY.NEGLIGIBLE';
      case VULNERABILITY_SEVERITY.UNKNOWN:
        return 'VULNERABILITY.SEVERITY.UNKNOWN';
      default:
        return 'UNKNOWN';
    }
  }
}
