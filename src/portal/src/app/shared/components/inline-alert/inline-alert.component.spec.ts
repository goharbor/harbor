import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { InlineAlertComponent } from './inline-alert.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { SharedTestingModule } from "../../shared.module";
describe('InlineAlertComponent', () => {
    let component: InlineAlertComponent;
    let fixture: ComponentFixture<InlineAlertComponent>;

    beforeEach(waitForAsync(() => {
        TestBed.configureTestingModule({
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            imports: [
                SharedTestingModule
            ],
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(InlineAlertComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
