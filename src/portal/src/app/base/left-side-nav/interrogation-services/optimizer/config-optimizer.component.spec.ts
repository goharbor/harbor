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
import {
    ComponentFixture,
    ComponentFixtureAutoDetect,
    TestBed,
} from '@angular/core/testing';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { ConfigurationOptimizerComponent } from './config-optimizer.component';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { OptimizerMetadataComponent } from './optimizer-metadata/optimizer-metadata.component';
import { NewOptimizerModalComponent } from './new-optimizer-modal/new-optimizer-modal.component';
import { NewOptimizerFormComponent } from './new-optimizer-form/new-optimizer-form.component';
import { OptimizerService } from '../../../../../../ng-swagger-gen/services/optimizer.service';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { Registry } from '../../../../../../ng-swagger-gen/models/registry';
import { ClrLoadingState } from '@clr/angular';

describe('ConfigurationOptimizerComponent', () => {
    let mockOptimizerMetadata = {
        optimizer: {
            name: 'test1',
            vendor: 'trivy',
            version: '1.0.1',
        },
        capabilities: [
            {
                consumes_mime_types: ['consumes_mime_types'],
                produces_mime_types: ['consumes_mime_types'],
            },
        ],
    };
    let mockOptimizer1 = {
        name: 'test1',
        description: 'just a sample',
        version: '1.0.0',
        url: 'http://168.0.0.1',
    };
    let component: ConfigurationOptimizerComponent;
    let fixture: ComponentFixture<ConfigurationOptimizerComponent>;
    let fakedConfigOptimizerService = {
        getOptimizerMetadata() {
            return of(mockOptimizerMetadata).pipe(delay(10));
        },
        listOptimizersResponse() {
            const response: HttpResponse<Array<Registry>> = new HttpResponse<
                Array<Registry>
            >({
                headers: new HttpHeaders({
                    'x-total-count': [mockOptimizer1].length.toString(),
                }),
                body: [mockOptimizer1],
            });
            return of(response).pipe(delay(10));
        },
        updateOptimizer() {
            return of(true);
        },
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [
                ConfigurationOptimizerComponent,
                OptimizerMetadataComponent,
                NewOptimizerModalComponent,
                NewOptimizerFormComponent,
            ],
            providers: [
                {
                    provide: OptimizerService,
                    useValue: fakedConfigOptimizerService,
                },
                // open auto detect
                { provide: ComponentFixtureAutoDetect, useValue: true },
            ],
        });
    });
    beforeEach(() => {
        fixture = TestBed.createComponent(ConfigurationOptimizerComponent);
        component = fixture.componentInstance;
        component.newOptimizerDialog.saveBtnState = ClrLoadingState.LOADING;
        fixture.detectChanges();
    });
    it('should create', async () => {
        await fixture.whenStable();
        expect(component).toBeTruthy();
        expect(component.optimizers.length).toBe(1);
    });
    it('should be clickable', () => {
        component.selectedRow = mockOptimizer1;
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            let el: HTMLElement =
                fixture.nativeElement.querySelector('#set-default');
            expect(el.getAttribute('disable')).toBeFalsy();
        });
    });
    it('edit a optimizer', () => {
        component.selectedRow = mockOptimizer1;
        component.editOptimizer();
        expect(component.newOptimizerDialog.opened).toBeTruthy();
        fixture.detectChanges();
        fixture.nativeElement.querySelector('#optimizer-name').value = 'test456';
        fixture.nativeElement.querySelector('#button-save').click();
        fixture.detectChanges();
        expect(component.newOptimizerDialog.opened).toBeFalsy();
    });
});
