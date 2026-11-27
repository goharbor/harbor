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
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { SharedTestingModule } from '../../../shared/shared.module';
import { ModelSyncPageComponent } from './model-sync-page.component';
import {
    formatFileFilters,
    isExecutionInProgress,
    isTaskInProgress,
    parseFileFilters,
} from './model-sync';

describe('ModelSyncPageComponent', () => {
    let component: ModelSyncPageComponent;
    let fixture: ComponentFixture<ModelSyncPageComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [ModelSyncPageComponent],
        }).compileComponents();
        fixture = TestBed.createComponent(ModelSyncPageComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});

describe('model sync helpers', () => {
    it('should parse and format file filters', () => {
        expect(parseFileFilters('')).toEqual([]);
        expect(parseFileFilters(' *.safetensors, *.json\nREADME.md ')).toEqual([
            '*.safetensors',
            '*.json',
            'README.md',
        ]);
        expect(formatFileFilters(['a', 'b'])).toEqual('a, b');
        expect(formatFileFilters(null)).toEqual('');
    });

    it('should map statuses', () => {
        expect(isExecutionInProgress({ status: 'InProgress' })).toBeTrue();
        expect(isExecutionInProgress({ status: 'Succeed' })).toBeFalse();
        expect(isExecutionInProgress(null)).toBeFalse();
        expect(isTaskInProgress({ status: 'Pending' })).toBeTrue();
        expect(isTaskInProgress({ status: 'Failed' })).toBeFalse();
    });
});
