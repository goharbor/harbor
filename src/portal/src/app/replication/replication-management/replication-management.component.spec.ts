import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ReplicationManagementComponent } from './replication-management.component';

xdescribe('ReplicationManagementComponent', () => {
    let component: ReplicationManagementComponent;
    let fixture: ComponentFixture<ReplicationManagementComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [ReplicationManagementComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ReplicationManagementComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
