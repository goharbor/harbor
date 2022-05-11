import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { SessionService } from '../../../shared/services/session.service';
import { GcPageComponent } from './gc-page.component';
import { SharedTestingModule } from '../../../shared/shared.module';

describe('GcPageComponent', () => {
    let component: GcPageComponent;
    let fixture: ComponentFixture<GcPageComponent>;
    let fakeSessionService = {
        getCurrentUser: function () {
            return { has_admin_role: true };
        },
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [GcPageComponent],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            providers: [
                { provide: SessionService, useValue: fakeSessionService },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(GcPageComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
