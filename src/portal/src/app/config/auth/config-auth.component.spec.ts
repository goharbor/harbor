import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ConfigurationAuthComponent } from './config-auth.component';

xdescribe('ConfigurationAuthComponent', () => {
    let component: ConfigurationAuthComponent;
    let fixture: ComponentFixture<ConfigurationAuthComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [ConfigurationAuthComponent]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ConfigurationAuthComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
