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

    beforeEach(() => {
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
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should rightly display default repositories and tag', () => {
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            fixture.detectChanges();
            let elRep: HTMLInputElement =
                fixture.nativeElement.querySelector('#scope-input');
            expect(elRep).toBeTruthy();
            expect(elRep.value.trim()).toEqual('**');
            let elTag: HTMLInputElement =
                fixture.nativeElement.querySelector('#tag-input');
            expect(elTag).toBeTruthy();
            expect(elTag.value.trim()).toEqual('**');
        });
    });
    it('should rightly close', () => {
        fixture.detectChanges();
        let elRep: HTMLButtonElement =
            fixture.nativeElement.querySelector('#close-btn');
        elRep.dispatchEvent(new Event('click'));
        elRep.click();
        fixture.whenStable().then(() => {
            fixture.detectChanges();
            expect(component.addRuleOpened).toEqual(false);
        });
    });
    it('should be validating repeat rule ', () => {
        fixture.detectChanges();
        component.rules = [mockRule];
        const elRep: HTMLButtonElement =
            fixture.nativeElement.querySelector('#add-edit-btn');
        elRep.dispatchEvent(new Event('click'));
        elRep.click();
        fixture.whenStable().then(() => {
            fixture.detectChanges();
            const elRep1: HTMLSpanElement =
                fixture.nativeElement.querySelector('.alert-text');
            expect(elRep1).toBeTruthy();
        });
    });
});
