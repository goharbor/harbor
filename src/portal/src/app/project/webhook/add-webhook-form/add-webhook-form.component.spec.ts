import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { AddWebhookFormComponent } from './add-webhook-form.component';

xdescribe('AddWebhookFormComponent', () => {
    let component: AddWebhookFormComponent;
    let fixture: ComponentFixture<AddWebhookFormComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [AddWebhookFormComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(AddWebhookFormComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
