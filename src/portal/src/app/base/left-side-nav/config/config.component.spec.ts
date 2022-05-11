import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ConfigurationComponent } from './config.component';
import { SharedTestingModule } from '../../../shared/shared.module';
import { ConfigService } from './config.service';
import { Configuration } from './config';

describe('ConfigurationComponent', () => {
    let component: ConfigurationComponent;
    let fixture: ComponentFixture<ConfigurationComponent>;
    const fakeConfigService = {
        getConfig() {
            return new Configuration();
        },
        getOriginalConfig() {
            return new Configuration();
        },
        getLoadingConfigStatus() {
            return false;
        },
        updateConfig() {},
        initConfig() {},
    };
    let initSpy: jasmine.Spy;
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            declarations: [ConfigurationComponent],
            providers: [
                { provide: ConfigService, useValue: fakeConfigService },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        initSpy = spyOn(fakeConfigService, 'initConfig').and.returnValue(
            undefined
        );
        fixture = TestBed.createComponent(ConfigurationComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should init config', async () => {
        await fixture.whenStable();
        expect(initSpy.calls.count()).toEqual(1);
    });
});
