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
import {
    of,
    throwError,
} from 'rxjs';
import { ArtifactVEXComponent } from './artifact-vex.component';
import { AdditionsService } from '../additions.service';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';
import * as utils from '../../../../../../shared/units/utils';

describe('ArtifactVEXComponent', () => {
    let component: ArtifactVEXComponent;
    let fixture: ComponentFixture<ArtifactVEXComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ArtifactVEXComponent],
            providers: [
                {
                    provide: AdditionsService,
                    useValue: {
                        getDetailByLink: () => of('{"@context":"https://openvex.dev/ns/v0.2.0"}'),
                    },
                },
                { provide: ErrorHandler, useValue: { error() {} } },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ArtifactVEXComponent);
        component = fixture.componentInstance;
        component.vexLink = { absolute: false, href: '/vex' };
        fixture.detectChanges();
    });

    it('should display the VEX document', () => {
        expect(fixture.nativeElement.querySelector('.vex-content').textContent).toContain('openvex.dev');
    });

    it('should not display the empty-state message when loading fails', () => {
        const additionsService = TestBed.inject(AdditionsService);
        additionsService.getDetailByLink = () => throwError(() => new Error('Request failed'));
        fixture = TestBed.createComponent(ArtifactVEXComponent);
        component = fixture.componentInstance;
        component.vexLink = { absolute: false, href: '/vex' };
        fixture.detectChanges();

        expect(component.error).toBeTrue();
        expect(fixture.nativeElement.querySelector('.vex-content')).toBeNull();
        expect(fixture.nativeElement.textContent).not.toContain('No VEX document');
    });

    it('should download the parsed object, not the pretty-printed string', () => {
        const downloadJsonSpy = spyOn(utils, 'downloadJson');
        component.download();
        expect(downloadJsonSpy).toHaveBeenCalledWith(
            { '@context': 'https://openvex.dev/ns/v0.2.0' },
            'vex.json'
        );
    });
});