import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TotalReplicationPageComponent } from './total-replication-page.component';
import { Router, ActivatedRoute } from '@angular/router';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { SessionService } from '../../../shared/services/session.service';
import { AppConfigService } from '../../../services/app-config.service';
import { SharedTestingModule } from '../../../shared/shared.module';

describe('TotalReplicationPageComponent', () => {
    let component: TotalReplicationPageComponent;
    let fixture: ComponentFixture<TotalReplicationPageComponent>;
    const mockSessionService = {
        getCurrentUser: () => {},
    };
    const mockAppConfigService = {
        getConfig: () => {
            return {
                project_creation_restriction: '',
                with_chartmuseum: '',
            };
        },
    };
    const mockRouter = {
        navigate: () => {},
        events: {
            subscribe: () => {},
        },
    };
    const mockActivatedRoute = null;
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [TotalReplicationPageComponent],
            providers: [
                { provide: SessionService, useValue: mockSessionService },
                { provide: AppConfigService, useValue: mockAppConfigService },
                { provide: Router, useValue: mockRouter },
                { provide: ActivatedRoute, useValue: mockActivatedRoute },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(TotalReplicationPageComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
