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
import { AddRobotComponent } from './add-robot.component';
import { of } from 'rxjs';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { delay } from 'rxjs/operators';
import { RobotService } from '../../../../../../ng-swagger-gen/services/robot.service';
import { OperationService } from '../../../../shared/components/operation/operation.service';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { SharedTestingModule } from '../../../../shared/shared.module';

describe('AddRobotComponent', () => {
    let component: AddRobotComponent;
    let fixture: ComponentFixture<AddRobotComponent>;
    const fakedRobotService = {
        ListRobot() {
            return of([]).pipe(delay(0));
        },
    };
    const fakedMessageHandlerService = {
        showSuccess() {},
        error() {},
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [AddRobotComponent],
            imports: [SharedTestingModule],
            providers: [
                OperationService,
                { provide: RobotService, useValue: fakedRobotService },
                {
                    provide: MessageHandlerService,
                    useValue: fakedMessageHandlerService,
                },
            ],
            schemas: [NO_ERRORS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(AddRobotComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
