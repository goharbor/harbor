import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ArtifactAdditionsComponent } from './artifact-additions.component';
import { AdditionLinks } from "../../../../../../ng-swagger-gen/models/addition-links";
import { IServiceConfig, SERVICE_CONFIG } from "../../../../../lib/entities/service.config";
import { ProjectModule } from "../../../project.module";
import { CURRENT_BASE_HREF } from "../../../../../lib/utils/utils";


describe('ArtifactAdditionsComponent', () => {
  const mockedAdditionLinks: AdditionLinks = {
   vulnerabilities: {
     absolute: false,
     href: CURRENT_BASE_HREF + "/test"
   }
  };
  const config: IServiceConfig = {
    baseEndpoint: CURRENT_BASE_HREF
  };
  let component: ArtifactAdditionsComponent;
  let fixture: ComponentFixture<ArtifactAdditionsComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        ProjectModule
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
