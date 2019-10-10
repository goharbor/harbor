import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { StatisticsComponent } from './statistics.component';

describe('StatisticsComponent', () => {
    let component: StatisticsComponent;
    let fixture: ComponentFixture<StatisticsComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [StatisticsComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(StatisticsComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
