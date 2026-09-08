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
import { CUSTOM_ELEMENTS_SCHEMA, EventEmitter } from '@angular/core';
import { ImmutableTagService } from '../immutable-tag.service';
import { ImmutableRetentionRule } from '../../tag-retention/retention';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import { InlineAlertComponent } from '../../../../../shared/components/inline-alert/inline-alert.component';
import { AddImmutableRuleComponent } from './add-immutable-rule.component';
import { SharedTestingModule } from '../../../../../shared/shared.module';

describe('AddRuleComponent', () => {
    let component: AddImmutableRuleComponent;
    let fixture: ComponentFixture<AddImmutableRuleComponent>;
    let mockRule = {
        id: 1,
        project_id: 1,
        disabled: false,
        priority: 0,
        action: 'immutable',
        template: 'immutable_template',
        tag_selectors: [
            {
                kind: 'doublestar',
                decoration: 'matches',
                pattern: '**',
            },
        ],
        scope_selectors: {
            repository: [
                {
                    kind: 'doublestar',
                    decoration: 'repoMatches',
                    pattern: '**',
                },
            ],
        },
    };
    const mockErrorHandler = {
        handleErrorPopupUnauthorized: () => {},
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [AddImmutableRuleComponent, InlineAlertComponent],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            providers: [
                ImmutableTagService,
                ErrorHandler,
                {
                    provide: ErrorHandler,
                    useValue: mockErrorHandler,
                },
            ],
        }).compileComponents();
    });

    beforeEach(async () => {
        fixture = TestBed.createComponent(AddImmutableRuleComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
        component.addRuleOpened = true;
        component.repoSelect =
            mockRule.scope_selectors.repository[0].decoration;
        component.repositories =
            mockRule.scope_selectors.repository[0].pattern.replace(/[{}]/g, '');
        component.tagsSelect = mockRule.tag_selectors[0].decoration;
        component.tagsInput = mockRule.tag_selectors[0].pattern.replace(
            /[{}]/g,
            ''
        );
        component.clickAdd = new EventEmitter<ImmutableRetentionRule>();
        component.rules = [];
        component.isAdd = true;
        component.open();
        fixture.detectChanges();
        await fixture.whenStable();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should rightly display default repositories and tag', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        fixture.detectChanges();
        const elRep: HTMLInputElement =
            fixture.nativeElement.querySelector('#scope-input');
        expect(elRep).toBeTruthy();
        expect(elRep.value.trim()).toEqual('**');
        const elTag: HTMLInputElement =
            fixture.nativeElement.querySelector('#tag-input');
        expect(elTag).toBeTruthy();
        expect(elTag.value.trim()).toEqual('**');
    });
    it('should rightly close', async () => {
        fixture.detectChanges();
        const elRep: HTMLButtonElement =
            fixture.nativeElement.querySelector('#close-btn');
        elRep.dispatchEvent(new Event('click'));
        elRep.click();
        await fixture.whenStable();
        fixture.detectChanges();
        expect(component.addRuleOpened).toEqual(false);
    });
    it('should be validating repeat rule ', async () => {
        fixture.detectChanges();
        component.rules = [mockRule];
        const elRep: HTMLButtonElement =
            fixture.nativeElement.querySelector('#add-edit-btn');
        elRep.dispatchEvent(new Event('click'));
        elRep.click();
        await fixture.whenStable();
        fixture.detectChanges();
        const elRep1: HTMLSpanElement =
            fixture.nativeElement.querySelector('.alert-text');
        expect(elRep1).toBeTruthy();
    });
});
