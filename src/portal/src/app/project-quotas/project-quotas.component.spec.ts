import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ProjectQuotasComponent } from './project-quotas.component';
import { TranslateModule } from "@ngx-translate/core";
import { CUSTOM_ELEMENTS_SCHEMA } from "@angular/core";
import { MessageHandlerService } from "../shared/message-handler/message-handler.service";
import { SessionService } from "../shared/session.service";
import { SessionUser } from "../shared/session-user";
import { ConfigurationService } from "../config/config.service";
import { of } from "rxjs";
import { Configuration } from "../../lib/components/config/config";

describe('ProjectQuotasComponent', () => {
  let component: ProjectQuotasComponent;
  let fixture: ComponentFixture<ProjectQuotasComponent>;
  const mockedUser: SessionUser = {
    user_id: 1,
    username: 'admin',
    email: 'harbor@vmware.com',
    realname: 'admin',
    has_admin_role: true,
    comment: 'no comment'
  };
  let mockedConfig: Configuration = new Configuration();
  mockedConfig.count_per_project.value = 10;
  const fakedSessionService = {
    getCurrentUser() {
      return mockedUser;
    }
  };
  const fakedConfigurationService = {
    getConfiguration() {
      return of(mockedConfig);
    }
  };
  const fakedMessageHandlerService = {
    handleError() {
      return;
    }
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        TranslateModule.forRoot()
      ],
      declarations: [ ProjectQuotasComponent ],
      providers: [
        MessageHandlerService,
        {provide: MessageHandlerService, useValue: fakedMessageHandlerService},
        {provide: SessionService, useValue: fakedSessionService},
        {provide: ConfigurationService, useValue: fakedConfigurationService}
      ],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ProjectQuotasComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should get config', () => {
    expect(component.allConfig.count_per_project.value).toEqual(10);
  });
});
