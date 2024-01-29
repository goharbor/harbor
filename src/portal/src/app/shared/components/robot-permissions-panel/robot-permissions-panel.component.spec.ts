import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Component, ViewChild } from '@angular/core';
import { SharedTestingModule } from '../../shared.module';
import {
    PermissionSelectPanelModes,
    RobotPermissionsPanelComponent,
} from './robot-permissions-panel.component';

describe('RobotPermissionsPanelComponent', () => {
    let component: TestHostComponent;
    let fixture: ComponentFixture<TestHostComponent>;
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [TestHostComponent, RobotPermissionsPanelComponent],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(TestHostComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should render right mode', async () => {
        component.robotPermissionsPanelComponent.modalOpen = true;
        fixture.detectChanges();
        await fixture.whenStable();
        const table = fixture.nativeElement.querySelector('table');
        expect(table).toBeTruthy();
        component.mode = PermissionSelectPanelModes.DROPDOWN;
        fixture.detectChanges();
        await fixture.whenStable();
        const clrDropdown = fixture.nativeElement.querySelector('clr-dropdown');
        expect(clrDropdown).toBeTruthy();
        component.mode = PermissionSelectPanelModes.MODAL;
        fixture.detectChanges();
        await fixture.whenStable();
        const modal = fixture.nativeElement.querySelector('clr-modal');
        expect(modal).toBeTruthy();
    });
});

// mock a TestHostComponent for RobotPermissionsPanelComponent
@Component({
    template: `
        <ng-container *ngIf="mode === PermissionSelectPanelModes.MODAL">
            <robot-permissions-panel [mode]="mode">
                <div>modal</div>
            </robot-permissions-panel>
        </ng-container>
        <ng-container *ngIf="mode === PermissionSelectPanelModes.DROPDOWN">
            <robot-permissions-panel [mode]="mode">
                <div>dropDown</div>
            </robot-permissions-panel>
        </ng-container>
        <ng-container *ngIf="mode === PermissionSelectPanelModes.NORMAL">
            <robot-permissions-panel [mode]="mode"> </robot-permissions-panel>
        </ng-container>
    `,
})
class TestHostComponent {
    @ViewChild(RobotPermissionsPanelComponent)
    robotPermissionsPanelComponent: RobotPermissionsPanelComponent;
    mode = PermissionSelectPanelModes.NORMAL;
    protected readonly PermissionSelectPanelModes = PermissionSelectPanelModes;
}
