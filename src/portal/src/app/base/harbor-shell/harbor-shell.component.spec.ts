import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { HarborShellComponent } from './harbor-shell.component';

xdescribe('HarborShellComponent', () => {
    let component: HarborShellComponent;
    let fixture: ComponentFixture<HarborShellComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [HarborShellComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(HarborShellComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
