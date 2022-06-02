import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ArtifactCommonPropertiesComponent } from './artifact-common-properties.component';
import { ClarityModule } from '@clr/angular';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import {
    TranslateFakeLoader,
    TranslateLoader,
    TranslateModule,
    TranslateService,
} from '@ngx-translate/core';
import { ExtraAttrs } from '../../../../../../../ng-swagger-gen/models/extra-attrs';

describe('ArtifactCommonPropertiesComponent', () => {
    let component: ArtifactCommonPropertiesComponent;
    let fixture: ComponentFixture<ArtifactCommonPropertiesComponent>;
    const mockedExtraAttrs: ExtraAttrs = {
        architecture: 'amd64',
        author: '',
        created: '2019-11-11T09:42:44.892055836Z',
        os: 'linux',
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [
                ClarityModule,
                BrowserAnimationsModule,
                TranslateModule.forRoot({
                    loader: {
                        provide: TranslateLoader,
                        useClass: TranslateFakeLoader,
                    },
                }),
            ],
            declarations: [ArtifactCommonPropertiesComponent],
            providers: [TranslateService],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ArtifactCommonPropertiesComponent);
        component = fixture.componentInstance;
        component.artifactDetails = {};
        component.artifactDetails.extra_attrs = mockedExtraAttrs;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should render all properties', async () => {
        component.commonProperties = mockedExtraAttrs;
        fixture.detectChanges();
        await fixture.whenStable();
        const contentRows =
            fixture.nativeElement.getElementsByTagName('clr-stack-content');
        expect(contentRows.length).toEqual(4);
    });
});
