import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ReplicationPageComponent } from './replication-page.component';

xdescribe('ReplicationPageComponent', () => {
    let component: ReplicationPageComponent;
    let fixture: ComponentFixture<ReplicationPageComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [ReplicationPageComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ReplicationPageComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
