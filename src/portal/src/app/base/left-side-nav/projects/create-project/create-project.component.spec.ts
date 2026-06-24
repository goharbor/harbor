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
        createProject: function () {
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

    it('should default access level to private', () => {
        component.newProject();
        expect(component.project.metadata.public).toBe('false');
    });

    it('should render Private, Auth Only, and Public radio buttons', async () => {
        component.createProjectOpened = true;
        fixture.detectChanges();
        await fixture.whenStable();
        const privateRadio: HTMLInputElement =
            fixture.nativeElement.querySelector('#create_project_private');
        const authOnlyRadio: HTMLInputElement =
            fixture.nativeElement.querySelector('#create_project_auth_only');
        const publicRadio: HTMLInputElement =
            fixture.nativeElement.querySelector('#create_project_public');
        expect(privateRadio).toBeTruthy();
        expect(authOnlyRadio).toBeTruthy();
        expect(publicRadio).toBeTruthy();
        expect(privateRadio.type).toBe('radio');
        expect(authOnlyRadio.type).toBe('radio');
        expect(publicRadio.type).toBe('radio');
    });

    it('should update metadata.public to auth_only when Auth Only radio is clicked', async () => {
        component.createProjectOpened = true;
        fixture.detectChanges();
        await fixture.whenStable();
        const authOnlyRadio: HTMLInputElement =
            fixture.nativeElement.querySelector('#create_project_auth_only');
        authOnlyRadio.click();
        fixture.detectChanges();
        await fixture.whenStable();
        expect(component.project.metadata.public).toBe('auth_only');
    });

    it('should update metadata.public to true when Public radio is clicked', async () => {
        component.createProjectOpened = true;
        fixture.detectChanges();
        await fixture.whenStable();
        const publicRadio: HTMLInputElement =
            fixture.nativeElement.querySelector('#create_project_public');
        publicRadio.click();
        fixture.detectChanges();
        await fixture.whenStable();
        expect(component.project.metadata.public).toBe('true');
    });

    it('should pass auth_only string to API on submit', async () => {
        let capturedArg: any;
        spyOn(mockProjectService, 'createProject').and.callFake(
            ((params: any) => {
                capturedArg = params;
                return of(true);
            }) as any
        );
        component.createProjectOpened = true;
        fixture.detectChanges();
        await fixture.whenStable();
        component.project.name = 'myproject';
        component.project.metadata.public = 'auth_only';
        component.onSubmit();
        await fixture.whenStable();
        expect(capturedArg).toBeTruthy();
        expect(capturedArg.project.metadata.public).toBe('auth_only');
    });

    it('should pass true string to API for public projects', async () => {
        let capturedArg: any;
        spyOn(mockProjectService, 'createProject').and.callFake(
            ((params: any) => {
                capturedArg = params;
                return of(true);
            }) as any
        );
        component.createProjectOpened = true;
        fixture.detectChanges();
        await fixture.whenStable();
        component.project.name = 'mypublicproject';
        component.project.metadata.public = 'true';
        component.onSubmit();
        await fixture.whenStable();
        expect(capturedArg.project.metadata.public).toBe('true');
    });
});
