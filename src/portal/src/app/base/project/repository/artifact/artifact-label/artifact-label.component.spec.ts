import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ArtifactLabelComponent } from './artifact-label.component';
import { ClarityModule } from '@clr/angular';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import {
    TranslateFakeLoader,
    TranslateLoader,
    TranslateModule,
    TranslateService,
} from '@ngx-translate/core';

describe('ArtifactLabelComponent', () => {
    let component: ArtifactLabelComponent;
    let fixture: ComponentFixture<ArtifactLabelComponent>;

    const mockedExtraAttrs = {
        extra_attrs: {
            config: {
                architecture: 'transformer',
                format: 'tensorflow',
                parameterSize: 50000000000,
                precision: 'int8',
                puantization: 'gptq',
            },
            descriptor: {
                createdAt: '2025-02-21T15:42:00.309773+08:00',
                family: 'qwen2',
                name: 'Qwen2.5-0.5B',
            },
        },
        type: 'MODEL',
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
            declarations: [ArtifactLabelComponent],
            providers: [TranslateService],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ArtifactLabelComponent);
        component = fixture.componentInstance;
        component.artifactDetails = mockedExtraAttrs;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should render all properties', async () => {
        component.artifactExtraAttrs = mockedExtraAttrs.extra_attrs;
        component.ngOnInit();
        fixture.detectChanges();
        await fixture.whenStable();

        const contentRows = fixture.nativeElement.getElementsByTagName('img');
        expect(contentRows.length).toBeGreaterThan(1);
    });
});
