import { ComponentFixture, TestBed } from '@angular/core/testing';
import { InterrogationServicesComponent } from './interrogation-services.component';
import { SharedTestingModule } from '../../../shared/shared.module';
import { TranslateService } from '@ngx-translate/core';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';

describe('InterrogationServicesComponent', () => {
    let component: InterrogationServicesComponent;
    let fixture: ComponentFixture<InterrogationServicesComponent>;
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [InterrogationServicesComponent],
            providers: [TranslateService],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(InterrogationServicesComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
