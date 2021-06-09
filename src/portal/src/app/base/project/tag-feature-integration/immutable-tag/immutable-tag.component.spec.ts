import { waitForAsync, ComponentFixture, TestBed } from '@angular/core/testing';
import { ImmutableTagComponent } from './immutable-tag.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ImmutableTagService } from './immutable-tag.service';
import { ActivatedRoute } from '@angular/router';
import { of, throwError } from 'rxjs';
import { ErrorHandler } from "../../../../shared/units/error-handler";
import { clone } from "../../../../shared/units/utils";
import { InlineAlertComponent } from "../../../../shared/components/inline-alert/inline-alert.component";
import { AddImmutableRuleComponent } from "./add-rule/add-immutable-rule.component";
import { SharedTestingModule } from "../../../../shared/shared.module";
import { RuleMetadate } from "../tag-retention/retention";

describe('ImmutableTagComponent', () => {
  let component: ImmutableTagComponent;
  let addRuleComponent: AddImmutableRuleComponent;
  let immutableTagService: ImmutableTagService;
  let errorHandler: ErrorHandler;
  let fixture: ComponentFixture<ImmutableTagComponent>;
  let fixtureAddrule: ComponentFixture<AddImmutableRuleComponent>;
  let mockMetadata: RuleMetadate = {
    "templates": [
      {
        "rule_template": "latestPushedK",
        "display_text": "the most recently pushed # images",
        "action": "retain",
        "params": [
          {
            "type": "int",
            "unit": "COUNT",
            "required": true
          }
        ]
      },
      {
        "rule_template": "latestPulledN",
        "display_text": "the most recently pulled # images",
        "action": "retain",
        "params": [
          {
            "type": "int",
            "unit": "COUNT",
            "required": true
          }
        ]
      },
      {
        "rule_template": "nDaysSinceLastPush",
        "display_text": "pushed within the last # days",
        "action": "retain",
        "params": [
          {
            "type": "int",
            "unit": "DAYS",
            "required": true
          }
        ]
      },
      {
        "rule_template": "nDaysSinceLastPull",
        "display_text": "pulled within the last # days",
        "action": "retain",
        "params": [
          {
            "type": "int",
            "unit": "DAYS",
            "required": true
          }
        ]
      },
      {
        "rule_template": "always",
        "display_text": "always",
        "action": "retain",
        "params": []
      }
    ],
    "scope_selectors": [
      {
        "display_text": "Repositories",
        "kind": "doublestar",
        "decorations": [
          "repoMatches",
          "repoExcludes"
        ]
      }
    ],
    "tag_selectors": [
      {
        "display_text": "Tags",
        "kind": "doublestar",
        "decorations": [
          "matches",
          "excludes"
        ]
      }
    ]
  };
  let mockRules =
    [
      {
        "id": 1,
        "project_id": 1,
        "disabled": false,
        "priority": 0,
        "action": "immutable",
        "template": "immutable_template",
        "tag_selectors": [
          {
            "kind": "doublestar",
            "decoration": "matches",
            "pattern": "**"
          }
        ],
        "scope_selectors": {
          "repository": [
            {
              "kind": "doublestar",
              "decoration": "repoMatches",
              "pattern": "**"
            }
          ]
        }
      }, {
        "id": 2,
        "project_id": 1,
        "disabled": false,
        "priority": 0,
        "action": "immutable",
        "template": "immutable_template",
        "tag_selectors": [
          {
            "kind": "doublestar",
            "decoration": "matches",
            "pattern": "44"
          }
        ],
        "scope_selectors": {
          "repository": [
            {
              "kind": "doublestar",
              "decoration": "repoMatches",
              "pattern": "**"
            }
          ]
        }
      },
      {
        "id": 3,
        "project_id": 1,
        "disabled": false,
        "priority": 0,
        "action": "immutable",
        "template": "immutable_template",
        "tag_selectors": [
          {
            "kind": "doublestar",
            "decoration": "matches",
            "pattern": "555"
          }
        ],
        "scope_selectors": {
          "repository": [
            {
              "kind": "doublestar",
              "decoration": "repoMatches",
              "pattern": "**"
            }
          ]
        }
      },
      {
        "id": 4,
        "project_id": 1,
        "disabled": false,
        "priority": 0,
        "action": "immutable",
        "template": "immutable_template",
        "tag_selectors": [
          {
            "kind": "doublestar",
            "decoration": "matches",
            "pattern": "fff**"
          }
        ],
        "scope_selectors": {
          "repository": [
            {
              "kind": "doublestar",
              "decoration": "repoMatches",
              "pattern": "**ggg"
            }
          ]
        }
      }
    ];
    let cloneRule = clone(mockRules[0]);
    cloneRule.tag_selectors[0].pattern = 'rep';
    let cloneRuleNoId = clone(mockRules[0]);
    cloneRuleNoId.id = null;
    let cloneDisableRule = clone(mockRules[0]);
    cloneDisableRule.disabled = true;
  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [ImmutableTagComponent, AddImmutableRuleComponent, InlineAlertComponent],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
      imports: [
        SharedTestingModule
      ],
      providers: [
        ImmutableTagService,
        {
          provide: ActivatedRoute, useValue: {
            paramMap: of({ get: (key) => 'value' }),
            snapshot: {
              parent: {
                parent: {
                  parent: {
                    params: { id: 1 }
                  }
                }
              },
              data: 1
            }
          }
        },
      ]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ImmutableTagComponent);
    fixtureAddrule = TestBed.createComponent(AddImmutableRuleComponent);
    component = fixture.componentInstance;
    addRuleComponent = fixtureAddrule.componentInstance;
    addRuleComponent.open = () => {
      return null;
    };
    component.projectId = 1;

    component.addRuleComponent = TestBed.createComponent(AddImmutableRuleComponent).componentInstance;
    component.addRuleComponent = TestBed.createComponent(AddImmutableRuleComponent).componentInstance;
    component.addRuleComponent.open = () => {
      return null;
    };
    component.addRuleComponent.inlineAlert = TestBed.createComponent(InlineAlertComponent).componentInstance;

    immutableTagService = fixture.debugElement.injector.get(ImmutableTagService);
    errorHandler = fixture.debugElement.injector.get(ErrorHandler);
    spyOn(immutableTagService, "getRetentionMetadata")
      .and.returnValue(of(mockMetadata));
    spyOn(immutableTagService, "getRules")
      .withArgs(component.projectId)
      .and.returnValue(of(mockRules))
      .withArgs(0)
      .and.returnValue(throwError('error'));

    spyOn(immutableTagService, "updateRule")
      .withArgs(component.projectId, mockRules[0])
      .and.returnValue(of(null))
      .withArgs(component.projectId, cloneRule)
      .and.returnValue(of(null))
      .withArgs(component.projectId, cloneDisableRule)
      .and.returnValue(of(null));
    spyOn(immutableTagService, "deleteRule")
      .withArgs(component.projectId, mockRules[3].id)
      .and.returnValue(of(null));
    spyOn(immutableTagService, "createRule")
      .withArgs(component.projectId, cloneRuleNoId)
      .and.returnValue(of(null))
      .withArgs(0, cloneRuleNoId)
      .and.returnValue(throwError({error: { message: 'error'}}));
    spyOn(immutableTagService, "getProjectInfo")
      .withArgs(component.projectId)
      .and.returnValue(of(null));

    spyOn(errorHandler, "error")
      .and.returnValue(null);
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it("should show some rules in page", waitForAsync(() => {
    fixture.detectChanges();

    fixture.whenStable().then(() => {
      fixture.detectChanges();

      let elRep: HTMLLIElement[] = fixture.nativeElement.querySelectorAll(".rule");
      expect(elRep).toBeTruthy();
      expect(elRep.length).toEqual(4);
    });
  }));
  it("should show error in list rule", waitForAsync(() => {
    fixture.detectChanges();
    component.projectId = 0;
    component.getRules();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      component.projectId = 1;
    });
  }));
  it("should  toggle disable and enable", waitForAsync(() => {
    fixture.detectChanges();
    let elRep: HTMLButtonElement = fixture.nativeElement.querySelector("#action0");
    elRep.dispatchEvent(new Event('click'));
    elRep.click();
    let elRepDisable: HTMLButtonElement = fixture.nativeElement.querySelector("#disable-btn0");
    expect(elRepDisable).toBeTruthy();
    elRepDisable.dispatchEvent(new Event('click'));
    elRepDisable.click();
    mockRules[0].disabled = true;

    fixture.whenStable().then(() => {
      fixture.detectChanges();
      let elRepDisableIcon: HTMLButtonElement = fixture.nativeElement.querySelector("#disable-icon0");
      expect(elRepDisableIcon).toBeTruthy();
    });
  }));
  it("should be deleted", waitForAsync(() => {
    fixture.detectChanges();
    let elRep: HTMLButtonElement = fixture.nativeElement.querySelector("#action0");
    elRep.dispatchEvent(new Event('click'));
    elRep.click();
    let elRepDisable: HTMLButtonElement = fixture.nativeElement.querySelector("#delete-btn3");
    expect(elRepDisable).toBeTruthy();
    elRepDisable.dispatchEvent(new Event('click'));
    elRepDisable.click();
    let rule = mockRules.pop();

    fixture.whenStable().then(() => {
      fixture.detectChanges();
      let elRepRule: HTMLLIElement[] = fixture.nativeElement.querySelectorAll(".rule");
      expect(elRepRule.length).toEqual(3);
      mockRules.push(rule);
    });
  }));

  it("should be add rule", waitForAsync(() => {
    fixture.detectChanges();
    component.clickAdd(cloneRuleNoId);
    mockRules.push(cloneRuleNoId);
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      let elRepRule: HTMLLIElement[] = fixture.nativeElement.querySelectorAll(".rule");
      expect(elRepRule.length).toEqual(5);
      mockRules.pop();
    });

  }));
  it("should be add rule error", waitForAsync(() => {
    fixture.detectChanges();
    component.projectId = 0;
    component.clickAdd(cloneRuleNoId);
    // mockRules.push(cloneRuleNoId);
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      component.projectId = 1;
      let elRepRule: HTMLLIElement[] = fixture.nativeElement.querySelectorAll(".rule");
      expect(elRepRule.length).toEqual(4);
      // mockRules.pop();
    });

  }));
  it("should be edit rule ", waitForAsync(() => {
    fixture.detectChanges();
    component.clickAdd(cloneRule);
    mockRules[0].tag_selectors[0].pattern = 'rep';
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      let elRepRule: HTMLLIElement = fixture.nativeElement.querySelector("#tag-selectors-patten0");
      expect(elRepRule.textContent).toEqual('rep');
      mockRules[0].tag_selectors[0].pattern = '**';
    });

  }));
  it("should be edit rule with no add", waitForAsync(() => {
    fixture.detectChanges();
    component.addRuleComponent.isAdd = false;
    component.clickAdd(cloneRule);
    mockRules[0].tag_selectors[0].pattern = 'rep';
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      let elRepRule: HTMLLIElement = fixture.nativeElement.querySelector("#tag-selectors-patten0");
      expect(elRepRule.textContent).toEqual('rep');
      mockRules[0].tag_selectors[0].pattern = '**';
      component.addRuleComponent.isAdd = true;
    });

  }));

});
