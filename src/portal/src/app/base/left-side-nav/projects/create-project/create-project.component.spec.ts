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
import { CreateProjectComponent } from './create-project.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { InlineAlertComponent } from '../../../../shared/components/inline-alert/inline-alert.component';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { ProjectService } from '../../../../../../ng-swagger-gen/services/project.service';

describe('CreateProjectComponent', () => {
    let component: CreateProjectComponent;
    let fixture: ComponentFixture<CreateProjectComponent>;
    const mockProjectService = {
        listProjects: function (params: ProjectService.ListProjectsParams) {
            if (params && params.q === encodeURIComponent('name=test')) {
                return of([true]).pipe(delay(10));
            } else {
                return of([]).pipe(delay(10));
            }
        },
        createProject: function (_params: ProjectService.CreateProjectParams) {
            return of(true);
        },
    };
    const mockMessageHandlerService = {
        showSuccess: function () {},
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [CreateProjectComponent, InlineAlertComponent],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            providers: [
                { provide: ProjectService, useValue: mockProjectService },
                {
                    provide: MessageHandlerService,
                    useValue: mockMessageHandlerService,
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(CreateProjectComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should open and close', async () => {
        let modelBody: HTMLDivElement;
        modelBody = fixture.nativeElement.querySelector('.modal-body');
        expect(modelBody).toBeFalsy();
        component.createProjectOpened = true;
        fixture.detectChanges();
        await fixture.whenStable();
        modelBody = fixture.nativeElement.querySelector('.modal-body');
        expect(modelBody).toBeTruthy();
        const cancelButton: HTMLButtonElement =
            fixture.nativeElement.querySelector('#new-project-cancel');
        cancelButton.click();
        fixture.detectChanges();
        await fixture.whenStable();
        modelBody = fixture.nativeElement.querySelector('.modal-body');
        expect(modelBody).toBeFalsy();
    });

    it('should check project name', async () => {
        fixture.autoDetectChanges(true);
        component.createProjectOpened = true;
        await fixture.whenStable();
        const nameInput: HTMLInputElement = fixture.nativeElement.querySelector(
            '#create_project_name'
        );
        nameInput.blur();
        nameInput.dispatchEvent(new Event('blur'));
        await fixture.whenStable();
        let el: HTMLSpanElement;
        el = fixture.nativeElement.querySelector('#name-error');
        expect(el).toBeTruthy();
        nameInput.value = 'test';
        nameInput.dispatchEvent(new Event('input'));
        nameInput.blur();
        nameInput.dispatchEvent(new Event('blur'));
        await fixture.whenStable();
        el = fixture.nativeElement.querySelector('#name-error');
        expect(el).toBeTruthy();
        nameInput.value = 'test1';
        nameInput.dispatchEvent(new Event('input'));
        nameInput.blur();
        nameInput.dispatchEvent(new Event('blur'));
        await fixture.whenStable();
        el = fixture.nativeElement.querySelector('#name-error');
        expect(el).toBeFalsy();
        const okButton: HTMLButtonElement =
            fixture.nativeElement.querySelector('#new-project-ok');
        okButton.click();
        await fixture.whenStable();
        const modelBody: HTMLDivElement =
            fixture.nativeElement.querySelector('.modal-body');
        expect(modelBody).toBeFalsy();
    });

    it('should enable proxy cache', async () => {
        component.enableProxyCache = true;
        component.isSystemAdmin = true;
        fixture.detectChanges();
        await fixture.whenStable();
        const endpoint: HTMLDivElement =
            fixture.nativeElement.querySelector('#endpoint');
        expect(endpoint).toBeFalsy();
    });

    it('should show the default retention period only for proxy cache creation', async () => {
        component.createProjectOpened = true;
        component.isSystemAdmin = true;
        fixture.detectChanges();
        await fixture.whenStable();
        expect(
            fixture.nativeElement.querySelector('#retentionDays')
        ).toBeNull();

        component.enableProxyCache = true;
        fixture.detectChanges();
        await fixture.whenStable();
        const retentionInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#retentionDays');
        expect(retentionInput.value).toBe('7');

        component.enableProxyCache = false;
        fixture.detectChanges();
        await fixture.whenStable();
        expect(
            fixture.nativeElement.querySelector('#retentionDays')
        ).toBeNull();
    });

    it('should create the project with the retention period entered in the dialog', async () => {
        const createProject = spyOn(
            mockProjectService,
            'createProject'
        ).and.callThrough();
        component.createProjectOpened = true;
        component.isSystemAdmin = true;
        component.enableProxyCache = true;
        component.project.registry_id = 1;
        component.project.name = 'proxy-project';
        component.storageLimit = -1;
        fixture.detectChanges();
        await fixture.whenStable();
        const retentionInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#retentionDays');
        retentionInput.value = '30';
        retentionInput.dispatchEvent(new Event('input'));
        fixture.detectChanges();
        await fixture.whenStable();

        const okButton: HTMLButtonElement =
            fixture.nativeElement.querySelector('#new-project-ok');
        expect(okButton.disabled).toBeFalse();
        okButton.click();
        expect(createProject).toHaveBeenCalledWith({
            project: jasmine.objectContaining({
                project_name: 'proxy-project',
                registry_id: 1,
                retention_days: 30,
            }),
        });
    });

    it('should disable creation for an invalid retention period and recover when proxy cache is disabled', async () => {
        component.createProjectOpened = true;
        component.isSystemAdmin = true;
        component.enableProxyCache = true;
        component.project.registry_id = 1;
        component.project.name = 'proxy-project';
        component.storageLimit = -1;
        fixture.detectChanges();
        await fixture.whenStable();
        const retentionInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#retentionDays');
        retentionInput.value = '1.5';
        retentionInput.dispatchEvent(new Event('input'));
        fixture.detectChanges();
        await fixture.whenStable();
        const okButton: HTMLButtonElement =
            fixture.nativeElement.querySelector('#new-project-ok');
        expect(okButton.disabled).toBeTrue();
        expect(
            retentionInput.parentElement.querySelector('clr-control-error')
        ).toBeTruthy();

        component.enableProxyCache = false;
        fixture.detectChanges();
        await fixture.whenStable();
        fixture.detectChanges();
        expect(okButton.disabled).toBeFalse();
    });

    [null, -1, 1.5, 106752].forEach(retentionDays => {
        it(`should reject retention period ${retentionDays} before submitting`, () => {
            const createProject = spyOn(mockProjectService, 'createProject');
            const showError = spyOn(component.inlineAlert, 'showInlineError');
            component.enableProxyCache = true;
            component.retentionDays = retentionDays;
            component.onSubmit();
            expect(createProject).not.toHaveBeenCalled();
            expect(showError).toHaveBeenCalledWith(
                'PROJECT.PROXY_CACHE_RETENTION_DAYS_INVALID'
            );
        });
    });

    [0, 7, 18250].forEach(retentionDays => {
        it(`should submit valid retention period ${retentionDays}`, () => {
            const createProject = spyOn(
                mockProjectService,
                'createProject'
            ).and.callThrough();
            component.enableProxyCache = true;
            component.project.registry_id = 1;
            component.retentionDays = retentionDays;
            component.onSubmit();
            expect(createProject).toHaveBeenCalledWith({
                project: jasmine.objectContaining({
                    registry_id: 1,
                    retention_days: retentionDays,
                }),
            });
        });
    });

    it('should omit retention settings for a normal project', () => {
        const createProject = spyOn(
            mockProjectService,
            'createProject'
        ).and.callThrough();
        component.retentionDays = null;
        component.onSubmit();
        expect(createProject).toHaveBeenCalled();
        expect(
            Object.prototype.hasOwnProperty.call(
                createProject.calls.mostRecent().args[0].project,
                'retention_days'
            )
        ).toBeFalse();
    });

    it('should restore the default retention period when reopening the dialog', () => {
        component.retentionDays = 30;
        component.onCancel();
        component.newProject();
        expect(component.retentionDays).toBe(7);
    });
});
