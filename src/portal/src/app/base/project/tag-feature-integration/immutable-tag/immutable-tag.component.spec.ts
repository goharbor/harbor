import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ImmutableTagComponent } from './immutable-tag.component';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ImmutableTagService } from './immutable-tag.service';
import { ActivatedRoute } from '@angular/router';
import { of, throwError } from 'rxjs';
import { clone } from '../../../../shared/units/utils';
import { InlineAlertComponent } from '../../../../shared/components/inline-alert/inline-alert.component';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { AddImmutableRuleComponent } from './add-rule/add-immutable-rule.component';

describe('ImmutableTagComponent', () => {
    let component: ImmutableTagComponent;
    let fixture: ComponentFixture<ImmutableTagComponent>;
    const mockRules = [
        {
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
        },
        {
            id: 2,
            project_id: 1,
            disabled: false,
            priority: 0,
            action: 'immutable',
            template: 'immutable_template',
            tag_selectors: [
                {
                    kind: 'doublestar',
                    decoration: 'matches',
                    pattern: '44',
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
        },
        {
            id: 3,
            project_id: 1,
            disabled: false,
            priority: 0,
            action: 'immutable',
            template: 'immutable_template',
            tag_selectors: [
                {
                    kind: 'doublestar',
                    decoration: 'matches',
                    pattern: '555',
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
        },
        {
            id: 4,
            project_id: 1,
            disabled: false,
            priority: 0,
            action: 'immutable',
            template: 'immutable_template',
            tag_selectors: [
                {
                    kind: 'doublestar',
                    decoration: 'matches',
                    pattern: 'fff**',
                },
            ],
            scope_selectors: {
                repository: [
                    {
                        kind: 'doublestar',
                        decoration: 'repoMatches',
                        pattern: '**ggg',
                    },
                ],
            },
        },
    ];
    const fakedImmutableTagService = {
        getI18nKey() {
            return 'test';
        },
        getRetentionMetadata() {
            return throwError(() => {
                return { error: { message: 'error' } };
            });
        },
        getRules(projectId) {
            if (projectId) {
                return of(mockRules);
            }
            return throwError(() => {
                return 'error';
            });
        },
        updateRule() {
            return of(null);
        },
        deleteRule() {
            return of(null);
        },
        createRule(projectId, cloneRuleNoId) {
            if (projectId) {
                return of(mockRules);
            }
            return throwError(() => {
                return { error: { message: 'error' } };
            });
        },
        getProjectInfo() {
            return of(null);
        },
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [
                ImmutableTagComponent,
                InlineAlertComponent,
                AddImmutableRuleComponent,
            ],
            schemas: [NO_ERRORS_SCHEMA],
            imports: [SharedTestingModule],
            providers: [
                {
                    provide: ImmutableTagService,
                    useValue: fakedImmutableTagService,
                },
                {
                    provide: ActivatedRoute,
                    useValue: {
                        paramMap: of({ get: key => 'value' }),
                        snapshot: {
                            parent: {
                                parent: {
                                    parent: {
                                        params: { id: 1 },
                                    },
                                },
                            },
                            data: 1,
                        },
                    },
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ImmutableTagComponent);
        component = fixture.componentInstance;
        component.projectId = 1;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should show some rules in page', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        let elRep: HTMLLIElement[] =
            fixture.nativeElement.querySelectorAll('.rule');
        expect(elRep).toBeTruthy();
        expect(elRep.length).toEqual(4);
    });
    it('should show error in list rule', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        component.projectId = 0;
        component.getRules();
    });
    it('should  toggle disable and enable', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        let elRep: HTMLButtonElement =
            fixture.nativeElement.querySelector('#action0');
        elRep.dispatchEvent(new Event('click'));
        elRep.click();
        fixture.detectChanges();
        await fixture.whenStable();
        let elRepDisable: HTMLButtonElement =
            fixture.nativeElement.querySelector('#disable-btn0');
        expect(elRepDisable).toBeTruthy();
        elRepDisable.dispatchEvent(new Event('click'));
        elRepDisable.click();
        mockRules[0].disabled = true;
        fixture.detectChanges();
        await fixture.whenStable();
        let elRepDisableIcon: HTMLButtonElement =
            fixture.nativeElement.querySelector('#disable-icon0');
        expect(elRepDisableIcon).toBeTruthy();
    });
    it('should be deleted', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        let elRep: HTMLButtonElement =
            fixture.nativeElement.querySelector('#action0');
        elRep.dispatchEvent(new Event('click'));
        elRep.click();
        fixture.detectChanges();
        await fixture.whenStable();
        let elRepDisable: HTMLButtonElement =
            fixture.nativeElement.querySelector('#delete-btn3');
        expect(elRepDisable).toBeTruthy();
        elRepDisable.dispatchEvent(new Event('click'));
        elRepDisable.click();
        fixture.detectChanges();
        await fixture.whenStable();
        fixture.detectChanges();
        await fixture.whenStable();
        let elRepRule: HTMLLIElement[] =
            fixture.nativeElement.querySelectorAll('.rule');
        expect(elRepRule.length).toEqual(4);
    });

    it('should be add rule', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        component.clickAdd(clone(mockRules[0]));
        fixture.detectChanges();
        await fixture.whenStable();
        let elRepRule: HTMLLIElement[] =
            fixture.nativeElement.querySelectorAll('.rule');
        expect(elRepRule.length).toEqual(4);
    });
    it('should be add rule error', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        component.projectId = 0;
        const rule = clone(mockRules[0]);
        rule.id = null;
        component.clickAdd(rule);
        fixture.detectChanges();
        await fixture.whenStable();
        component.projectId = 1;
        fixture.detectChanges();
        await fixture.whenStable();
        let elRepRule: HTMLLIElement[] =
            fixture.nativeElement.querySelectorAll('.rule');
        expect(elRepRule.length).toEqual(4);
    });
    it('should be edit rule ', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        component.clickAdd(clone(mockRules[0]));
        mockRules[0].tag_selectors[0].pattern = 'rep';
        fixture.detectChanges();
        await fixture.whenStable();
        let elRepRule: HTMLLIElement = fixture.nativeElement.querySelector(
            '#tag-selectors-patten0'
        );
        expect(elRepRule.textContent).toEqual('rep');
        mockRules[0].tag_selectors[0].pattern = '**';
    });
    it('should be edit rule with no add', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        component.addRuleComponent.isAdd = false;
        component.clickAdd(clone(mockRules[0]));
        mockRules[0].tag_selectors[0].pattern = 'rep';
        fixture.detectChanges();
        await fixture.whenStable();
        let elRepRule: HTMLLIElement = fixture.nativeElement.querySelector(
            '#tag-selectors-patten0'
        );
        expect(elRepRule.textContent).toEqual('rep');
        mockRules[0].tag_selectors[0].pattern = '**';
        component.addRuleComponent.isAdd = true;
    });
});
