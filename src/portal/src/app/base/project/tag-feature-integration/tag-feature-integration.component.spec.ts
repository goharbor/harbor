import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TagFeatureIntegrationComponent } from './tag-feature-integration.component';
import { SharedTestingModule } from '../../../shared/shared.module';
import { UserPermissionService } from '../../../shared/services';
import { ActivatedRoute } from '@angular/router';
import { of } from 'rxjs';

describe('TagFeatureIntegrationComponent', () => {
    let component: TagFeatureIntegrationComponent;
    let fixture: ComponentFixture<TagFeatureIntegrationComponent>;

    const mockActivatedRoute = {
        snapshot: {
            parent: {
                parent: {
                    params: { id: 1 },
                },
            },
        },
    };
    const mockUserPermissionService = {
        getPermission() {
            return of(true);
        },
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [TagFeatureIntegrationComponent],
            imports: [SharedTestingModule],
            providers: [
                {
                    provide: UserPermissionService,
                    useValue: mockUserPermissionService,
                },
                {
                    provide: ActivatedRoute,
                    useValue: mockActivatedRoute,
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(TagFeatureIntegrationComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should get project id and permissions', async () => {
        await fixture.whenStable();
        expect(component.projectId).toEqual(1);
        expect(component.hasTagImmutablePermission).toBeTruthy();
        expect(component.hasTagRetentionPermission).toBeTruthy();
    });
});
