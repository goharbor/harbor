import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { GcComponent } from './gc.component';
import { SERVICE_CONFIG, IServiceConfig } from '../../service.config';
import { GcApiRepository, GcApiDefaultRepository} from './gc.api.repository';
import { GcRepoService } from './gc.service';
import { SharedModule } from "../../shared/shared.module";
import { ErrorHandler } from '../../error-handler/error-handler';
import { GcViewModelFactory } from './gc.viewmodel.factory';
import { GcUtility } from './gc.utility';

describe('GcComponent', () => {
  let component: GcComponent;
  let fixture: ComponentFixture<GcComponent>;
  let config: IServiceConfig = {
    systemInfoEndpoint: "/api/system/gc"
  };
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      declarations: [ GcComponent ],
      providers: [
        { provide: GcApiRepository, useClass: GcApiDefaultRepository },
        { provide: SERVICE_CONFIG, useValue: config },
        GcRepoService,
        ErrorHandler,
        GcViewModelFactory,
        GcUtility
      ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(GcComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
