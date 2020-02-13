import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ArtifactAdditionsComponent } from './artifact-additions.component';
import { AdditionLinks } from "../../../../../ng-swagger-gen/models/addition-links";
import { HarborLibraryModule } from "../../../harbor-library.module";
import { IServiceConfig, SERVICE_CONFIG } from "../../../entities/service.config";

describe('ArtifactAdditionsComponent', () => {
  const mockedAdditionLinks: AdditionLinks = {
   vulnerabilities: {
     absolute: false,
     href: "api/v2/test"
   }
  };
  const config: IServiceConfig = {
    baseEndpoint: "/api/v2"
  };
  let component: ArtifactAdditionsComponent;
  let fixture: ComponentFixture<ArtifactAdditionsComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        HarborLibraryModule
      ],
      providers: [
        { provide: SERVICE_CONFIG, useValue: config },
      ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ArtifactAdditionsComponent);
    component = fixture.componentInstance;
    component.additionLinks = mockedAdditionLinks;
    fixture.detectChanges();
  });

  it('should create and render vulnerabilities tab', async () => {
    expect(component).toBeTruthy();
    await fixture.whenStable();
    const tabButton: HTMLButtonElement = fixture.nativeElement.querySelector('#vulnerability');
    expect(tabButton).toBeTruthy();
  });
});
