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
import { ListProjectROComponent } from './list-project-ro.component';
import { SharedTestingModule } from '../../shared.module';
import { Project } from '../../../../../ng-swagger-gen/models/project';
import { Component } from '@angular/core';

// mock a TestHostComponent for ListProjectROComponent
@Component({
    template: ` <list-project-ro [projects]="projects"> </list-project-ro>`,
})
class TestHostComponent {
    projects: Project[] = [];
}

describe('ListProjectROComponent', () => {
    let component: TestHostComponent;
    let fixture: ComponentFixture<TestHostComponent>;
    const mockedProjects: Project[] = [
        {
            name: 'test1',
            metadata: {},
            project_id: 1,
            repo_count: 1,
        },
        {
            name: 'test2',
            metadata: {},
            project_id: 2,
            repo_count: 1,
        },
    ];
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ListProjectROComponent, TestHostComponent],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(TestHostComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should render project list', async () => {
        component.projects = mockedProjects;
        fixture.detectChanges();
        await fixture.whenStable();
        const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(rows.length).toEqual(2);
    });
});
