import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ArtifactVulnerabilitiesComponent } from './artifact-vulnerabilities.component';
import { NO_ERRORS_SCHEMA } from "@angular/core";
import { ClarityModule } from "@clr/angular";
import { AdditionsService } from "../additions.service";
import { of } from "rxjs";
import { TranslateFakeLoader, TranslateLoader, TranslateModule } from "@ngx-translate/core";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { VulnerabilityItem } from "../../../../../../lib/services";
import { AdditionLink } from "../../../../../../../ng-swagger-gen/models/addition-link";
import { ErrorHandler } from "../../../../../../lib/utils/error-handler";


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
  const mockedLink: AdditionLink = {
    absolute: false,
    href: '/test'
  };
  const fakedAdditionsService = {
    getDetailByLink() {
      return of(mockedVulnerabilities);
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
        {provide: AdditionsService, useValue: fakedAdditionsService}
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
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
  it('should get vulnerability list and render', async () => {
    component.vulnerabilitiesLink = mockedLink;
    component.ngOnInit();
    fixture.detectChanges();
    await fixture.whenStable();
    const rows = fixture.nativeElement.getElementsByTagName('clr-dg-row');
    expect(rows.length).toEqual(2);
  });
});
