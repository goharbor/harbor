import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { HarborLibraryModule } from '../../harbor-library.module';
import { ConfirmationDialogComponent } from './confirmation-dialog.component';
import { IServiceConfig, SERVICE_CONFIG } from '../../entities/service.config';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { ConfirmationTargets } from '../../entities/shared.const';
import { ConfirmationMessage } from './confirmation-message';
import { BatchInfo } from './confirmation-batch-message';


describe('ConfirmationDialogComponent', () => {

    let comp: ConfirmationDialogComponent;
    let fixture: ComponentFixture<ConfirmationDialogComponent>;
    let config: IServiceConfig = {
        configurationEndpoint: '/api/configurations/testing'
    };
    const deletionMessage: ConfirmationMessage = new ConfirmationMessage(
        "MEMBER.DELETION_TITLE",
        "MEMBER.DELETION_SUMMARY",
        "user1",
        {},
        ConfirmationTargets.CONFIG
    );

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                HarborLibraryModule,
                BrowserAnimationsModule
            ],
            providers: [
                {provide: SERVICE_CONFIG, useValue: config}
            ]
        });
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ConfirmationDialogComponent);
        comp = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(comp).toBeTruthy();
    });
    it('should open dialog and show right buttons', async () => {
        let message = deletionMessage;
        let buttons: HTMLElement;
        comp.open(message);
        fixture.detectChanges();
        await fixture.whenStable();
        buttons = fixture.nativeElement.querySelector('.modal-footer button');
        expect(buttons.textContent).toEqual('BUTTON.CANCEL');
        message.buttons = 1;
        comp.open(message);
        fixture.detectChanges();
        await fixture.whenStable();
        buttons = fixture.nativeElement.querySelector('.modal-footer button');
        expect(buttons.textContent).toEqual('BUTTON.NO');
        message.buttons = 3;
        comp.open(message);
        fixture.detectChanges();
        await fixture.whenStable();
        buttons = fixture.nativeElement.querySelector('.modal-footer button');
        expect(buttons.textContent).toEqual('BUTTON.CLOSE');
    });
    it('check cancel and confirm function', async () => {
        let buttons: HTMLElement;
        comp.opened = true;
        comp.message = null;
        fixture.detectChanges();
        await fixture.whenStable();
        comp.cancel();
        fixture.detectChanges();
        await fixture.whenStable();
        buttons = fixture.nativeElement.querySelector('.modal-footer button');
        expect(buttons).toBeFalsy();
        comp.open(deletionMessage);
        fixture.detectChanges();
        await fixture.whenStable();
        comp.confirm();
        fixture.detectChanges();
        await fixture.whenStable();
        buttons = fixture.nativeElement.querySelector('.modal-footer button');
        expect(buttons).toBeFalsy();
    });
    it('check colorChange and toggleErrorTitle functions', () => {
        let batchInfo = new BatchInfo();
        const resultColor1: string = comp.colorChange(batchInfo);
        expect(resultColor1).toEqual("green");
        batchInfo.errorState = true;
        const resultColor2: string = comp.colorChange(batchInfo);
        expect(resultColor2).toEqual("red");
        batchInfo.loading = true;
        const resultColor3: string = comp.colorChange(batchInfo);
        expect(resultColor3).toEqual("#666");
        const errorSpan: HTMLSpanElement = document.createElement('span');
        errorSpan.style.display = "none";
        comp.toggleErrorTitle(errorSpan);
        expect(errorSpan.style.display).toEqual('block');
        comp.toggleErrorTitle(errorSpan);
        expect(errorSpan.style.display).toEqual('none');
    });
});
