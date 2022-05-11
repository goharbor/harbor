import { ComponentFixture, TestBed } from '@angular/core/testing';
import { EditProjectQuotasComponent } from './edit-project-quotas.component';
import { EditQuotaQuotaInterface } from '../../../../../shared/services';
import { SharedTestingModule } from '../../../../../shared/shared.module';

describe('EditProjectQuotasComponent', () => {
    let component: EditProjectQuotasComponent;
    let fixture: ComponentFixture<EditProjectQuotasComponent>;
    const mockedEditQuota: EditQuotaQuotaInterface = {
        editQuota: 'Edit Default Project Quotas',
        setQuota: 'Set the default project quotas when creating new projects',
        storageQuota: 'Default storage consumption',
        quotaHardLimitValue: { storageLimit: -1, storageUnit: 'Byte' },
        isSystemDefaultQuota: true,
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [EditProjectQuotasComponent],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(EditProjectQuotasComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });
    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
