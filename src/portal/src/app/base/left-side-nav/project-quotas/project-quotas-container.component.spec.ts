import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ProjectQuotasContainerComponent } from './project-quotas-container.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { SessionService } from '../../../shared/services/session.service';
import { SessionUser } from '../../../shared/entities/session-user';
import { ConfigurationService } from '../../../services/config.service';
import { of } from 'rxjs';
import { Configuration } from '../config/config';
import { SharedTestingModule } from '../../../shared/shared.module';

describe('ProjectQuotasContainerComponent', () => {
    let component: ProjectQuotasContainerComponent;
    let fixture: ComponentFixture<ProjectQuotasContainerComponent>;
    const mockedUser: SessionUser = {
        user_id: 1,
        username: 'admin',
        email: 'harbor@vmware.com',
        realname: 'admin',
        has_admin_role: true,
        comment: 'no comment',
    };
    let mockedConfig: Configuration = new Configuration();
    mockedConfig.count_per_project.value = 10;
    const fakedSessionService = {
        getCurrentUser() {
            return mockedUser;
        },
    };
    const fakedConfigurationService = {
        getConfiguration() {
            return of(mockedConfig);
        },
    };
    const fakedMessageHandlerService = {
        handleError() {
            return;
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ProjectQuotasContainerComponent],
            providers: [
                {
                    provide: MessageHandlerService,
                    useValue: fakedMessageHandlerService,
                },
                { provide: SessionService, useValue: fakedSessionService },
                {
                    provide: ConfigurationService,
                    useValue: fakedConfigurationService,
                },
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ProjectQuotasContainerComponent);
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
