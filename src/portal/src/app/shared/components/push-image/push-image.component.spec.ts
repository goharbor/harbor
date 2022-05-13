import { ComponentFixture, TestBed } from '@angular/core/testing';
import { PushImageButtonComponent } from './push-image.component';
import { CopyInputComponent } from './copy-input.component';
import { InlineAlertComponent } from '../inline-alert/inline-alert.component';
import { ErrorHandler } from '../../units/error-handler';
import { SharedTestingModule } from '../../shared.module';
import { Component } from '@angular/core';

describe('PushImageButtonComponent (inline template)', () => {
    let component: TestHostComponent;
    let fixture: ComponentFixture<TestHostComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [
                InlineAlertComponent,
                CopyInputComponent,
                PushImageButtonComponent,
                TestHostComponent,
            ],
            providers: [ErrorHandler],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(TestHostComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should be created', () => {
        expect(component).toBeTruthy();
    });

    it('should open the drop-down panel', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        let el: HTMLElement = fixture.nativeElement.querySelector('button');
        expect(el).not.toBeNull();
        el.click();
        fixture.detectChanges();
        await fixture.whenStable();
        let copyInputs: HTMLInputElement[] =
            fixture.nativeElement.querySelectorAll('.command-input');
        expect(copyInputs.length).toEqual(5);
        expect(copyInputs[0].value.trim()).toEqual(
            `docker tag SOURCE_IMAGE[:TAG] https://testing.harbor.com/testing/REPOSITORY[:TAG]`
        );
        expect(copyInputs[1].value.trim()).toEqual(
            `docker push https://testing.harbor.com/testing/REPOSITORY[:TAG]`
        );
    });
});

// mock a TestHostComponent for PushImageButtonComponent
@Component({
    template: ` <hbr-push-image-button
        [projectName]="'testing'"
        [registryUrl]="'https://testing.harbor.com'">
    </hbr-push-image-button>`,
})
class TestHostComponent {}
