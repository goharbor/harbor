import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from "@angular/core";
import { ClarityModule } from "@clr/angular";
import { AdditionsService } from "../additions.service";
import { of } from "rxjs";
import { TranslateFakeLoader, TranslateLoader, TranslateModule } from "@ngx-translate/core";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { BuildHistoryComponent } from "./build-history.component";
import { ArtifactBuildHistory } from "../models";
import { AdditionLink } from "../../../../../../../ng-swagger-gen/models/addition-link";
import { ErrorHandler } from "../../../../../../lib/utils/error-handler";

describe('BuildHistoryComponent', () => {
  let component: BuildHistoryComponent;
  let fixture: ComponentFixture<BuildHistoryComponent>;
  const mockedLink: AdditionLink = {
    absolute: false,
    href: '/test'
  };
  const mockedHistoryList: ArtifactBuildHistory[] = [
    {
      created: new Date(),
      created_by: 'test command'
    },
    {
      created: new Date(new Date().getTime() + 123456),
      created_by: 'test command'
    },
  ];
  const fakedAdditionsService = {
    getDetailByLink() {
      return of(mockedHistoryList);
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
      declarations: [BuildHistoryComponent],
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
    fixture = TestBed.createComponent(BuildHistoryComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
  it('should get build history list and render', async () => {
    component.buildHistoryLink = mockedLink;
    component.ngOnInit();
    fixture.detectChanges();
    await fixture.whenStable();
    const rows = fixture.nativeElement.getElementsByTagName('clr-dg-row');
    expect(rows.length).toEqual(2);
  });
});
