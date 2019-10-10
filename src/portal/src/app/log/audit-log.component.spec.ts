import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { AuditLogComponent } from './audit-log.component';

xdescribe('AuditLogComponent', () => {
    let component: AuditLogComponent;
    let fixture: ComponentFixture<AuditLogComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [AuditLogComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(AuditLogComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
