// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { RemainingTimeComponent } from './remaining-time.component';
import { Component, ViewChild } from '@angular/core';
import { RobotTimeRemainColor } from '../../../base/left-side-nav/system-robot-accounts/system-robot-util';
import { SharedTestingModule } from '../../shared.module';

describe('RemainingTimeComponent', () => {
    let component: TestHostComponent;
    let fixture: ComponentFixture<TestHostComponent>;
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [TestHostComponent, RemainingTimeComponent],
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
    it('should show green color', () => {
        fixture.detectChanges();
        expect(component?.remainingTimeComponent.color).toEqual(
            RobotTimeRemainColor.GREEN
        );
        expect(component?.remainingTimeComponent.timeRemain).toEqual(
            'ROBOT_ACCOUNT.NEVER_EXPIRED'
        );
    });
    it('should show yellow color', () => {
        component.deltaTime = 0;
        component.expires_at =
            new Date(new Date().getTime() + 1000 * 60 * 60 * 24 * 5).getTime() /
            1000;
        fixture.detectChanges();
        expect(component?.remainingTimeComponent.color).toEqual(
            RobotTimeRemainColor.WARNING
        );
    });
    it('should show red color', () => {
        component.deltaTime = 0;
        component.expires_at =
            new Date(
                new Date().getTime() - 1000 * 60 * 60 * 24 * 31
            ).getTime() / 1000;
        fixture.detectChanges();
        expect(component?.remainingTimeComponent.color).toEqual(
            RobotTimeRemainColor.EXPIRED
        );
        expect(component?.remainingTimeComponent.timeRemain).toEqual(
            'SYSTEM_ROBOT.EXPIRED'
        );
    });
});

// mock a TestHostComponent for ListProjectROComponent
@Component({
    template: ` <app-remaining-time
        [deadline]="expires_at"
        [timeDiff]="deltaTime"></app-remaining-time>`,
})
class TestHostComponent {
    @ViewChild(RemainingTimeComponent)
    remainingTimeComponent: RemainingTimeComponent;
    expires_at: number = -1;
    deltaTime: number = 100;
}
