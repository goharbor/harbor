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
import {
    ComponentFixture,
    fakeAsync,
    TestBed,
    tick,
} from '@angular/core/testing';
import { OperationComponent } from './operation.component';
import { OperationService } from './operation.service';
import { OperateInfo } from './operate';
import { SharedTestingModule } from '../../shared.module';

describe('OperationComponent', () => {
    let component: OperationComponent;
    let fixture: ComponentFixture<OperationComponent>;
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
        });
    });
    beforeEach(() => {
        fixture = TestBed.createComponent(OperationComponent);
        component = fixture.componentInstance;
        component.animationState = 'out';
        fixture.detectChanges();
    });
    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should automatically close', fakeAsync(async () => {
        component.animationState = 'in';
        fixture.detectChanges();
        // wait animation finishing
        tick(1000);
        await fixture.whenStable();
        const container: HTMLDivElement =
            fixture.nativeElement.querySelector('.operDiv');
        container.dispatchEvent(new Event('mouseleave'));
        fixture.detectChanges();
        // wait animation finishing
        tick(10000);
        await fixture.whenStable();
        const right: string = getComputedStyle(
            fixture.nativeElement.querySelector('.operDiv')
        ).right;
        expect(right).toEqual('-325px');
    }));
    it("should show '500+' after pushing 60 new operateInfos", fakeAsync(() => {
        const operationService: OperationService =
            TestBed.inject(OperationService);
        for (let i = 0; i < 520; i++) {
            let operateInfo = new OperateInfo();
            if (i > 19) {
                operateInfo.state = 'progressing';
            }
            if (i > 39) {
                operateInfo.state = 'failure';
            }
            tick(50000);
            operationService.publishInfo(operateInfo);
        }
        fixture.detectChanges();
        const toolBar: HTMLAnchorElement =
            fixture.nativeElement.querySelector('.toolBar');
        expect(toolBar.textContent).toContain('500+');
    }));
    it('check toggleTitle function', () => {
        const errorSpan: HTMLSpanElement = document.createElement('span');
        errorSpan.style.display = 'none';
        component.toggleTitle(errorSpan);
        expect(errorSpan.style.display).toEqual('block');
        component.toggleTitle(errorSpan);
        expect(errorSpan.style.display).toEqual('none');
    });
    it('check calculateTime function', () => {
        expect(
            component.calculateTime(
                1000,
                'less than 1 minute',
                ' minute(s) ago',
                ' hour(s) ago',
                ' day(s) ago'
            )
        ).toEqual('less than 1 minute');
        expect(
            component.calculateTime(
                61000,
                'less than 1 minute',
                ' minute(s) ago',
                ' hour(s) ago',
                ' day(s) ago'
            )
        ).toEqual('1 minute(s) ago');
        expect(
            component.calculateTime(
                3601000,
                'less than 1 minute',
                ' minute(s) ago',
                ' hour(s) ago',
                ' day(s) ago'
            )
        ).toEqual('1 hour(s) ago');
        expect(
            component.calculateTime(
                24 * 3601000,
                'less than 1 minute',
                ' minute(s) ago',
                ' hour(s) ago',
                ' day(s) ago'
            )
        ).toEqual('1 day(s) ago');
    });
});
