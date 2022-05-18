import { ComponentFixture, TestBed } from '@angular/core/testing';
import { SharedTestingModule } from 'src/app/shared/shared.module';
import { SubAccessoriesComponent } from './sub-accessories.component';
import { Accessory } from '../../../../../../../../../../ng-swagger-gen/models/accessory';
import { AccessoryType } from '../../../../artifact';
import { ArtifactService as NewArtifactService } from '../../../../../../../../../../ng-swagger-gen/services/artifact.service';
import { of } from 'rxjs';
import {
    ArtifactDefaultService,
    ArtifactService,
} from '../../../../artifact.service';

describe('SubAccessoriesComponent', () => {
    const mockedAccessories: Accessory[] = [
        {
            id: 1,
            artifact_id: 1,
            digest: 'sha256:test',
            type: AccessoryType.COSIGN,
            size: 1024,
        },
        {
            id: 2,
            artifact_id: 2,
            digest: 'sha256:test2',
            type: AccessoryType.COSIGN,
            size: 1024,
        },
        {
            id: 3,
            artifact_id: 3,
            digest: 'sha256:test3',
            type: AccessoryType.COSIGN,
            size: 1024,
        },
        {
            id: 4,
            artifact_id: 4,
            digest: 'sha256:test4',
            type: AccessoryType.COSIGN,
            size: 1024,
        },
        {
            id: 5,
            artifact_id: 5,
            digest: 'sha256:test5',
            type: AccessoryType.COSIGN,
            size: 1024,
        },
    ];

    const page2: Accessory[] = [
        {
            id: 6,
            artifact_id: 6,
            digest: 'sha256:test6',
            type: AccessoryType.COSIGN,
            size: 1024,
        },
    ];

    const mockedArtifactService = {
        listAccessories() {
            return of(page2);
        },
    };

    let component: SubAccessoriesComponent;
    let fixture: ComponentFixture<SubAccessoriesComponent>;
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [SubAccessoriesComponent],
            providers: [
                {
                    provide: NewArtifactService,
                    useValue: mockedArtifactService,
                },
                { provide: ArtifactService, useClass: ArtifactDefaultService },
            ],
        });
    });
    beforeEach(() => {
        fixture = TestBed.createComponent(SubAccessoriesComponent);
        component = fixture.componentInstance;
        component.accessories = mockedAccessories;
        component.total = 6;
        fixture.autoDetectChanges();
    });
    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should render rows', async () => {
        await fixture.whenStable();
        const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(rows.length).toEqual(5);
    });

    it('should render next page', async () => {
        await fixture.whenStable();
        const nextPageButton: HTMLButtonElement =
            fixture.nativeElement.querySelector('.pagination-next');
        nextPageButton.click();
        fixture.detectChanges();
        await fixture.whenStable();
        const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(rows.length).toEqual(1);
    });
});
