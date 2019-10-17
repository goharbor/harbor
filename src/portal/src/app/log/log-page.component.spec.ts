import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { LogPageComponent } from './log-page.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';

describe('LogPageComponent', () => {
    let component: LogPageComponent;
    let fixture: ComponentFixture<LogPageComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
            ],
            declarations: [LogPageComponent],
            providers: [
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(LogPageComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
