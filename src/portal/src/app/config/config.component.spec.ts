import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ConfigurationComponent } from './config.component';

xdescribe('ConfigurationComponent', () => {
    let component: ConfigurationComponent;
    let fixture: ComponentFixture<ConfigurationComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [ConfigurationComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ConfigurationComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
