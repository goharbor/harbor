import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ArtifactAdditionsComponent } from './artifact-additions.component';
import { AdditionLinks } from '../../../../../../../ng-swagger-gen/models/addition-links';
import { CURRENT_BASE_HREF } from '../../../../../shared/units/utils';
import { SharedTestingModule } from '../../../../../shared/shared.module';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ArtifactListPageService } from '../artifact-list-page/artifact-list-page.service';
import { ClrLoadingState } from '@clr/angular';

describe('ArtifactAdditionsComponent', () => {
    const mockedAdditionLinks: AdditionLinks = {
        vulnerabilities: {
            absolute: false,
            href: CURRENT_BASE_HREF + '/test',
        },
    };
    const mockedArtifactListPageService = {
        hasScannerSupportSBOM(): boolean {
            return true;
        },
        hasEnabledScanner(): boolean {
            return true;
        },
        getScanBtnState(): ClrLoadingState {
            return ClrLoadingState.SUCCESS;
        },
        init() {},
    };
    let component: ArtifactAdditionsComponent;
    let fixture: ComponentFixture<ArtifactAdditionsComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ArtifactAdditionsComponent],
            schemas: [NO_ERRORS_SCHEMA],
            providers: [
                {
                    provide: ArtifactListPageService,
                    useValue: mockedArtifactListPageService,
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ArtifactAdditionsComponent);
        component = fixture.componentInstance;
        component.additionLinks = mockedAdditionLinks;
        component.tab = 'vulnerability';
        fixture.detectChanges();
    });

    it('should create and render vulnerabilities tab', async () => {
        expect(component).toBeTruthy();
        await fixture.whenStable();
        const tabButton: HTMLButtonElement =
            fixture.nativeElement.querySelector('#vulnerability');
        expect(tabButton).toBeTruthy();
    });
});
