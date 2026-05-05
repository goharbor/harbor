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
import { ListAllProjectsComponent } from './list-all-projects.component';
import { Project } from '../../../../../../ng-swagger-gen/models/project';
import { SharedTestingModule } from '../../../../shared/shared.module';

describe('ListAllProjectsComponent', () => {
    let component: ListAllProjectsComponent;
    let fixture: ComponentFixture<ListAllProjectsComponent>;
    const project1: Project = {
        project_id: 1,
        name: 'project1',
    };
    const project2: Project = {
        project_id: 2,
        name: 'project2',
    };
    const project3: Project = {
        project_id: 3,
        name: 'project3',
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ListAllProjectsComponent],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ListAllProjectsComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should render list', async () => {
        component.projects = [project1, project2, project3];
        fixture.detectChanges();
        await fixture.whenStable();
        const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(rows.length).toEqual(3);
    });
});
