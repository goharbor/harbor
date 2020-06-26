import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from "@ngx-translate/core";
import { AppConfigService } from '../../services/app-config.service';
import { SkinableConfig } from "../../services/skinable-config.service";
import { AboutDialogComponent } from './about-dialog.component';
import { ClarityModule } from "@clr/angular";
import { SessionService } from '../../shared/session.service';

describe('AboutDialogComponent', () => {
    let component: AboutDialogComponent;
    let fixture: ComponentFixture<AboutDialogComponent>;
    let fakeAppConfigService = {
        getConfig: function() {
            return {
                harbor_version: '1.10'
            };
        }
    };
    let fakeSkinableConfig = {
        getProject: function () {
            return {
                introduction: {}
            };
        }
    };
    let fakeSessionService = {
        getCurrentUser: function () {
            return {
                username: 'abc',
                has_admin_role: true
            };
        }
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [AboutDialogComponent],
            imports: [
                TranslateModule.forRoot(),
                ClarityModule
            ],
            providers: [
                TranslateService,
                { provide: AppConfigService, useValue: fakeAppConfigService },
                { provide: SkinableConfig, useValue: fakeSkinableConfig },
                { provide: SessionService, useValue: fakeSessionService }
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(AboutDialogComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
