import { ComponentFixture, TestBed } from '@angular/core/testing';
import { AppConfigService } from '../../../services/app-config.service';
import { SkinableConfig } from '../../../services/skinable-config.service';
import { AboutDialogComponent } from './about-dialog.component';
import { SharedTestingModule } from '../../shared.module';

describe('AboutDialogComponent', () => {
    let component: AboutDialogComponent;
    let fixture: ComponentFixture<AboutDialogComponent>;
    let fakeAppConfigService = {
        getConfig: function () {
            return {
                harbor_version: '1.10',
            };
        },
    };
    let fakeSkinableConfig = {
        getSkinConfig: function () {
            return {
                headerBgColor: {
                    darkMode: '',
                    lightMode: '',
                },
                loginBgImg: '',
                loginTitle: '',
                product: {
                    name: '',
                    logo: '',
                    introduction: '',
                },
            };
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [AboutDialogComponent],
            imports: [SharedTestingModule],
            providers: [
                { provide: AppConfigService, useValue: fakeAppConfigService },
                { provide: SkinableConfig, useValue: fakeSkinableConfig },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(AboutDialogComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
