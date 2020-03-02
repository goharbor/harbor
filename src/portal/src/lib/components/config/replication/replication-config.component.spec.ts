import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ReplicationConfigComponent } from "./replication-config.component";
import { HarborLibraryModule } from "../../../harbor-library.module";
import { IServiceConfig, SERVICE_CONFIG } from "../../../entities/service.config";
import { Configuration } from "../config";
import { CURRENT_BASE_HREF } from "../../../utils/utils";
describe('ReplicationConfigComponent', () => {
  let component: ReplicationConfigComponent;
  let fixture: ComponentFixture<ReplicationConfigComponent>;
  const config: IServiceConfig = {
    baseEndpoint: CURRENT_BASE_HREF + "/testing"
  };
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
          HarborLibraryModule
      ],
       providers: [
           { provide: SERVICE_CONFIG, useValue: config }
       ]
    });
  });
  beforeEach(() => {
    fixture = TestBed.createComponent(ReplicationConfigComponent);
    component = fixture.componentInstance;
    component.config = new Configuration();
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
