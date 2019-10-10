import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { TagDetailPageComponent } from './tag-detail-page.component';

xdescribe('TagDetailPageComponent', () => {
    let component: TagDetailPageComponent;
    let fixture: ComponentFixture<TagDetailPageComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [TagDetailPageComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(TagDetailPageComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
