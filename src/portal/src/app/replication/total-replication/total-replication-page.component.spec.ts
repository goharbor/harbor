import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { TotalReplicationPageComponent } from './total-replication-page.component';

xdescribe('TotalReplicationPageComponent', () => {
    let component: TotalReplicationPageComponent;
    let fixture: ComponentFixture<TotalReplicationPageComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [TotalReplicationPageComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(TotalReplicationPageComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
