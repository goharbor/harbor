import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule } from '@ngx-translate/core';
import { CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA } from '@angular/core';
import {
    BrowserAnimationsModule,
    NoopAnimationsModule,
} from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { AddP2pPolicyComponent } from './add-p2p-policy.component';
import { P2pProviderService } from '../p2p-provider.service';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import { PreheatService } from '../../../../../../ng-swagger-gen/services/preheat.service';
import { ActivatedRoute } from '@angular/router';
import { SessionService } from '../../../../shared/services/session.service';
import { AppConfigService } from '../../../../services/app-config.service';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { ProjectService } from '../../../../shared/services';
import { InlineAlertComponent } from '../../../../shared/components/inline-alert/inline-alert.component';
describe('AddP2pPolicyComponent', () => {
    let component: AddP2pPolicyComponent;
    let fixture: ComponentFixture<AddP2pPolicyComponent>;
    const mockedAppConfigService = {
        getConfig() {
            return {
                with_notary: true,
            };
        },
    };
    const mockPreheatService = {
        CreatePolicy() {
            return of(true).pipe(delay(0));
        },
        UpdatePolicy() {
            return of(true).pipe(delay(0));
        },
        ListPolicies() {
            return of([]).pipe(delay(0));
        },
    };
    const mockActivatedRoute = {
        snapshot: {
            parent: {
                parent: {
                    params: { id: 1 },
                    data: {
                        projectResolver: {
                            name: 'library',
                            metadata: {
                                prevent_vul: 'true',
                                enable_content_trust: 'true',
                                severity: 'none',
                            },
                        },
                    },
                },
            },
        },
    };
    const mockedSessionService = {
        getCurrentUser() {
            return {
                has_admin_role: true,
            };
        },
    };
    const mockedProjectService = {
        getProject() {
            return of({
                name: 'library',
                metadata: {
                    prevent_vul: 'true',
                    enable_content_trust: 'true',
                    severity: 'none',
                },
            });
        },
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA],
            imports: [
                BrowserAnimationsModule,
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule,
                RouterTestingModule,
                NoopAnimationsModule,
                HttpClientTestingModule,
            ],
            declarations: [AddP2pPolicyComponent, InlineAlertComponent],
            providers: [
                P2pProviderService,
                ErrorHandler,
                { provide: PreheatService, useValue: mockPreheatService },
                { provide: ActivatedRoute, useValue: mockActivatedRoute },
                { provide: SessionService, useValue: mockedSessionService },
                { provide: AppConfigService, useValue: mockedAppConfigService },
                { provide: ProjectService, useValue: mockedProjectService },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(AddP2pPolicyComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should open  and close modal', async () => {
        await fixture.whenStable();
        component.isOpen = true;
        fixture.detectChanges();
        await fixture.whenStable();
        let modalBody: HTMLDivElement =
            fixture.nativeElement.querySelector('.modal-body');
        expect(modalBody).toBeTruthy();
        component.closeModal();
        fixture.detectChanges();
        await fixture.whenStable();
        modalBody = fixture.nativeElement.querySelector('.modal-body');
        expect(modalBody).toBeFalsy();
    });
    it("should show a 'name is required' error", async () => {
        fixture.autoDetectChanges(true);
        component.isOpen = true;
        await fixture.whenStable();
        const nameInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#name');
        nameInput.value = 'test';
        nameInput.dispatchEvent(new Event('input'));
        nameInput.value = null;
        nameInput.dispatchEvent(new Event('input'));
        nameInput.blur();
        const errorEle: HTMLElement =
            fixture.nativeElement.querySelector('clr-control-error');
        expect(errorEle.innerText).toEqual('P2P_PROVIDER.NAME_TOOLTIP');
    });
    it('save button should work', async () => {
        fixture.autoDetectChanges(true);
        component.isOpen = true;
        await fixture.whenStable();
        const spy: jasmine.Spy = spyOn(component, 'addOrSave').and.returnValue(
            undefined
        );
        component.tags = '**';
        component.repos = '**';
        component.policy = {
            provider_id: 1,
        };
        const nameInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#name');
        nameInput.value = 'policy1';
        nameInput.dispatchEvent(new Event('input'));
        fixture.detectChanges();
        await fixture.whenStable();
        expect(component.valid()).toBeTruthy();
        const addButton: HTMLButtonElement =
            fixture.nativeElement.querySelector('#new-policy');
        addButton.click();
        fixture.detectChanges();
        await fixture.whenStable();
        expect(spy.calls.count()).toEqual(1);
    });
});
