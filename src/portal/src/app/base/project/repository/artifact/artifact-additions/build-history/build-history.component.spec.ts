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
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { AdditionsService } from '../additions.service';
import { of } from 'rxjs';
import { BuildHistoryComponent } from './build-history.component';
import { ArtifactBuildHistory } from '../models';
import { AdditionLink } from '../../../../../../../../ng-swagger-gen/models/addition-link';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';
import { SharedTestingModule } from '../../../../../../shared/shared.module';

describe('BuildHistoryComponent', () => {
    let component: BuildHistoryComponent;
    let fixture: ComponentFixture<BuildHistoryComponent>;
    const mockedLink: AdditionLink = {
        absolute: false,
        href: '/test',
    };
    const mockedHistoryList: ArtifactBuildHistory[] = [
        {
            created: new Date(),
            created_by: 'test command',
        },
        {
            created: new Date(new Date().getTime() + 123456),
            created_by: 'test command',
        },
    ];
    const fakedAdditionsService = {
        getDetailByLink() {
            return of(mockedHistoryList);
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [BuildHistoryComponent],
            providers: [
                ErrorHandler,
                { provide: AdditionsService, useValue: fakedAdditionsService },
            ],
            schemas: [NO_ERRORS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(BuildHistoryComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should get build history list and render', async () => {
        component.buildHistoryLink = mockedLink;
        component.ngOnInit();
        fixture.detectChanges();
        await fixture.whenStable();
        const rows = fixture.nativeElement.getElementsByTagName('clr-dg-row');
        expect(rows.length).toEqual(2);
    });
});
