import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { TagRepositoryComponent } from './tag-repository.component';

xdescribe('TagRepositoryComponent', () => {
    let component: TagRepositoryComponent;
    let fixture: ComponentFixture<TagRepositoryComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [TagRepositoryComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(TagRepositoryComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
