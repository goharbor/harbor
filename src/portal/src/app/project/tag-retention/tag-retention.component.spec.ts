import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { TagRetentionComponent } from './tag-retention.component';

xdescribe('TagRetentionComponent', () => {
    let component: TagRetentionComponent;
    let fixture: ComponentFixture<TagRetentionComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [TagRetentionComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(TagRetentionComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
