import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ArtifactVulnerabilitiesComponent } from './artifact-vulnerabilities.component';
import { NO_ERRORS_SCHEMA } from "@angular/core";
import { ClarityModule } from "@clr/angular";
import { AdditionsService } from "../additions.service";
import { of } from "rxjs";
import { TranslateFakeLoader, TranslateLoader, TranslateModule } from "@ngx-translate/core";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { ScanningResultService, UserPermissionService, VulnerabilityItem } from "../../../../../../lib/services";
import { AdditionLink } from "../../../../../../../ng-swagger-gen/models/addition-link";
import { ErrorHandler } from "../../../../../../lib/utils/error-handler";
import { ChannelService } from "../../../../../../lib/services/channel.service";
import { DEFAULT_SUPPORTED_MIME_TYPE } from "../../../../../../lib/utils/utils";


describe('ArtifactVulnerabilitiesComponent', () => {
  let component: ArtifactVulnerabilitiesComponent;
  let fixture: ComponentFixture<ArtifactVulnerabilitiesComponent>;
  const mockedVulnerabilities: VulnerabilityItem[] = [
    {
      id: '123',
      severity: 'low',
      package: 'test',
      version: '1.0',
      links: ['testLink'],
      fix_version: '1.1.1',
      description: 'just a test'
    },
    {
      id: '456',
      severity: 'high',
      package: 'test',
      version: '1.0',
      links: ['testLink'],
      fix_version: '1.1.1',
      description: 'just a test'
    },
  ];
  let scanOverview = {};
  scanOverview[DEFAULT_SUPPORTED_MIME_TYPE] = {};
  scanOverview[DEFAULT_SUPPORTED_MIME_TYPE].vulnerabilities = mockedVulnerabilities;
  const mockedLink: AdditionLink = {
    absolute: false,
    href: '/test'
  };
  const fakedAdditionsService = {
    getDetailByLink() {
      return of(scanOverview);
    }
  };
  const fakedUserPermissionService = {
    hasProjectPermissions() {
      return of(true);
    }
  };
  const fakedScanningResultService = {
    getProjectScanner() {
      return of(true);
    }
  };
  const fakedChannelService = {
    ArtifactDetail$: {
      subscribe() {
        return null;
      }
    }
  };
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        BrowserAnimationsModule,
        ClarityModule,
        TranslateModule.forRoot({
          loader: {
            provide: TranslateLoader,
            useClass: TranslateFakeLoader,
          }
        })
      ],
      declarations: [ArtifactVulnerabilitiesComponent],
      providers: [
        ErrorHandler,
        {provide: AdditionsService, useValue: fakedAdditionsService},
        {provide: UserPermissionService, useValue: fakedUserPermissionService},
        {provide: ScanningResultService, useValue: fakedScanningResultService},
        {provide: ChannelService, useValue: fakedChannelService},
      ],
      schemas: [
        NO_ERRORS_SCHEMA
      ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ArtifactVulnerabilitiesComponent);
    component = fixture.componentInstance;
    component.hasScanningPermission = true;
    component.hasEnabledScanner = true;
    component.vulnerabilitiesLink = mockedLink;
    component.ngOnInit();
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
  it('should get vulnerability list and render', async () => {
    fixture.detectChanges();
    await fixture.whenStable();
    const rows = fixture.nativeElement.getElementsByTagName('clr-dg-row');
    expect(rows.length).toEqual(2);
  });
});
