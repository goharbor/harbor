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
import { ImageNameInputComponent } from './image-name-input.component';
import { of } from 'rxjs';
import { SharedTestingModule } from '../../shared.module';
import { ProjectService } from 'ng-swagger-gen/services/project.service';
import { Project } from 'ng-swagger-gen/models/project';

describe('ImageNameInputComponent (inline template)', () => {
    let comp: ImageNameInputComponent;
    let fixture: ComponentFixture<ImageNameInputComponent>;
    let spy: jasmine.Spy;

    let mockProjects: Project[] = [
        {
            project_id: 1,
            name: 'project_01',
            creation_time: '',
        },
        {
            project_id: 2,
            name: 'project_02',
            creation_time: '',
        },
    ];
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ImageNameInputComponent],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ImageNameInputComponent);
        comp = fixture.componentInstance;

        let projectService: ProjectService;
        projectService = fixture.debugElement.injector.get(ProjectService);
        spy = spyOn(projectService, 'listProjects').and.returnValues(
            of(mockProjects)
        );
    });

    it('should load data', () => {
        expect(spy.calls.any).toBeTruthy();
    });
});
