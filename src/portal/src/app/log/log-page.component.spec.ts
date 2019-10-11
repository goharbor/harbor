import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { LogPageComponent } from './log-page.component';

xdescribe('LogPageComponent', () => {
    let component: LogPageComponent;
    let fixture: ComponentFixture<LogPageComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [LogPageComponent]
        })
            .compileComponents();
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
