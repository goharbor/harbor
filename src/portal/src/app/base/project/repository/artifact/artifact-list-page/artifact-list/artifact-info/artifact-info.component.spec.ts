import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of } from 'rxjs';
import { ArtifactInfoComponent } from './artifact-info.component';
import { SharedTestingModule } from 'src/app/shared/shared.module';
import { RepositoryService } from 'ng-swagger-gen/services/repository.service';

describe('ArtifactInfoComponent', () => {
    let compRepo: ArtifactInfoComponent;
    let fixture: ComponentFixture<ArtifactInfoComponent>;
    let FakedRepositoryService = {
        updateRepository: () => of(null),
        getRepository: () => of({ description: '' }),
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ArtifactInfoComponent],
            providers: [
                {
                    provide: RepositoryService,
                    useValue: FakedRepositoryService,
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ArtifactInfoComponent);
        compRepo = fixture.componentInstance;
        fixture.detectChanges();
    });
    it('should create', () => {
        expect(compRepo).toBeTruthy();
    });
});
