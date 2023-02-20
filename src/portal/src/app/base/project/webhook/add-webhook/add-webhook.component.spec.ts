import { ComponentFixture, TestBed } from '@angular/core/testing';
import { AddWebhookComponent } from './add-webhook.component';
import { CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA } from '@angular/core';
import { SharedTestingModule } from '../../../../shared/shared.module';

describe('AddWebhookComponent', () => {
    let component: AddWebhookComponent;
    let fixture: ComponentFixture<AddWebhookComponent>;

    beforeEach(() => {
        TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [AddWebhookComponent],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(AddWebhookComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should open modal and should be edit model', async () => {
        component.isEdit = true;
        component.isOpen = true;
        fixture.detectChanges();
        await fixture.whenStable();
        const body: HTMLElement =
            fixture.nativeElement.querySelector('.modal-body');
        expect(body).toBeTruthy();
        const title: HTMLElement =
            fixture.nativeElement.querySelector('.modal-title');
        expect(title.innerText).toEqual('WEBHOOK.EDIT_WEBHOOK');
    });
});
