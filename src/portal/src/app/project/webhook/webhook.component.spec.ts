import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { WebhookComponent } from './webhook.component';

xdescribe('WebhookComponent', () => {
    let component: WebhookComponent;
    let fixture: ComponentFixture<WebhookComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [WebhookComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(WebhookComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
