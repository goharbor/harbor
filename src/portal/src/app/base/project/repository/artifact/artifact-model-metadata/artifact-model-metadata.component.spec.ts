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
import { SharedTestingModule } from '../../../../../shared/shared.module';
import { ArtifactModelMetadataComponent } from './artifact-model-metadata.component';

describe('ArtifactModelMetadataComponent', () => {
    let fixture: ComponentFixture<ArtifactModelMetadataComponent>;
    let comp: ArtifactModelMetadataComponent;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ArtifactModelMetadataComponent],
        }).compileComponents();
        fixture = TestBed.createComponent(ArtifactModelMetadataComponent);
        comp = fixture.componentInstance;
    });

    it('renders the model-spec descriptor and config of a CNAI artifact', () => {
        comp.artifactDetails = {
            type: 'CNAI',
            extra_attrs: {
                descriptor: {
                    name: 'Org/Model',
                    sourceURL: 'https://huggingface.co/Org/Model',
                    revision: 'abc',
                    licenses: ['apache-2.0'],
                },
                config: { format: 'safetensors', quantization: 'gptq' },
                modelfs: { type: 'layers', diffIds: ['a', 'b'] },
            },
        };
        comp.ngOnChanges();
        fixture.detectChanges();
        const el: HTMLElement = fixture.nativeElement;
        expect(el.querySelector('#model-metadata')).toBeTruthy();
        expect(comp.licenses).toEqual(['apache-2.0']);
        const link = el.querySelector('a');
        expect(link.getAttribute('href')).toBe(
            'https://huggingface.co/Org/Model'
        );
        const keys = comp.rows.map(r => r.key);
        expect(keys).toContain('ARTIFACT.MODEL_QUANTIZATION');
        expect(
            comp.rows.find(r => r.key === 'ARTIFACT.MODEL_LAYERS').value
        ).toBe('2');
    });

    it('does not link a source URL that is not http(s)', () => {
        comp.artifactDetails = {
            type: 'CNAI',
            extra_attrs: { descriptor: { sourceURL: 'javascript:alert(1)' } },
        };
        comp.ngOnChanges();
        expect(comp.rows[0].href).toBeUndefined();
    });

    it('renders nothing for other artifact types', () => {
        comp.artifactDetails = { type: 'IMAGE', extra_attrs: {} };
        comp.ngOnChanges();
        fixture.detectChanges();
        expect(
            fixture.nativeElement.querySelector('#model-metadata')
        ).toBeNull();
    });
});
