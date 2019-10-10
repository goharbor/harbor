import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ConfigurationEmailComponent } from './config-email.component';

xdescribe('ConfigurationEmailComponent', () => {
    let component: ConfigurationEmailComponent;
    let fixture: ComponentFixture<ConfigurationEmailComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [ConfigurationEmailComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ConfigurationEmailComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
