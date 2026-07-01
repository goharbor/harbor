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
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { AdditionsService } from '../additions.service';
import { of } from 'rxjs';
import { DockerfileComponent } from './dockerfile.component';
import { AdditionLink } from '../../../../../../../../ng-swagger-gen/models/addition-link';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';
import { SharedTestingModule } from '../../../../../../shared/shared.module';

describe('DockerfileComponent', () => {
    let component: DockerfileComponent;
    let fixture: ComponentFixture<DockerfileComponent>;
    const mockedLink: AdditionLink = {
        absolute: false,
        href: '/test',
    };
    const dockerfile: string =
        'FROM alpine:3.14\n\nRUN apk add --no-cache \\\n  python3 \\\n  py3-pip\n\nCOPY requirements.txt .\nRUN pip install -r requirements.txt\n\nCOPY . /app\nWORKDIR /app\n\nEXPOSE 5000\nCMD ["python3", "app.py"]\n';

    const fakedAdditionsService = {
        getDetailByLink() {
            return of(dockerfile);
        },
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [DockerfileComponent],
            providers: [
                ErrorHandler,
                { provide: AdditionsService, useValue: fakedAdditionsService },
            ],
            schemas: [NO_ERRORS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(DockerfileComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should get dockerfile and render', async () => {
        component.dockerfileLink = mockedLink;
        component.ngOnInit();
        fixture.detectChanges();
        await fixture.whenStable();
        fixture.detectChanges();
        expect(component.dockerfile).toEqual(dockerfile);
    });
});
