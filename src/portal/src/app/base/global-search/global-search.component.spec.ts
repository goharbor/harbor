import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { GlobalSearchComponent } from './global-search.component';
import { SearchTriggerService } from './search-trigger.service';
import { FormsModule } from '@angular/forms';
import { AppConfigService } from '../../app-config.service';
import { SkinableConfig } from "../../skinable-config.service";
import { RouterTestingModule } from '@angular/router/testing';
import { of } from 'rxjs';

describe('GlobalSearchComponent', () => {
    let component: GlobalSearchComponent;
    let fixture: ComponentFixture<GlobalSearchComponent>;
    let fakeSearchTriggerService = {
        searchClearChan$: {
            subscribe: function () {
            }
        }
    };
    let fakeAppConfigService = {
        isIntegrationMode: function () {
            return true;
        }
    };
    let fakeSkinableConfig = {
        getProject: function () {
            return {
                introduction: {}
            };
        }
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                TranslateModule.forRoot(),
                FormsModule,
                RouterTestingModule
            ],
            declarations: [GlobalSearchComponent],
            providers: [
                TranslateService,
                { provide: SearchTriggerService, useValue: fakeSearchTriggerService },
                { provide: AppConfigService, useValue: fakeAppConfigService },
                { provide: SkinableConfig, useValue: fakeSkinableConfig }
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(GlobalSearchComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
