import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { AddWebhookComponent } from './add-webhook.component';

xdescribe('AddWebhookComponent', () => {
    let component: AddWebhookComponent;
    let fixture: ComponentFixture<AddWebhookComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [AddWebhookComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(AddWebhookComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
