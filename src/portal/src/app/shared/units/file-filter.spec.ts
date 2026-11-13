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
    fileIcon,
    isExactPattern,
    matchFileFilter,
    parseFileFilter,
} from './file-filter';

describe('file-filter', () => {
    it('parses the string and the object form', () => {
        expect(parseFileFilter('*.json, !onnx/**\n*.txt')).toEqual({
            include: ['*.json', '*.txt'],
            exclude: ['onnx/**'],
        });
        expect(parseFileFilter({ include: ['a'] })).toEqual({
            include: ['a'],
            exclude: [],
        });
        expect(parseFileFilter('')).toEqual({ include: [], exclude: [] });
    });

    // same cases as pkg/reg/model.FileFilter.Match
    it('matches like the backend', () => {
        const f = { include: ['*.json'], exclude: ['onnx/**'] };
        expect(matchFileFilter(f, 'config.json')).toBeTrue();
        expect(matchFileFilter(f, 'sub/dir/config.json')).toBeTrue();
        expect(matchFileFilter(f, 'onnx/config.json')).toBeFalse();
        expect(matchFileFilter(f, 'model.safetensors')).toBeFalse();
        expect(matchFileFilter({ include: ['**'] }, 'a/b/c')).toBeTrue();
        expect(matchFileFilter({ exclude: ['*.bin'] }, 'x.bin')).toBeFalse();
        expect(matchFileFilter({}, 'anything')).toBeTrue();
        expect(
            matchFileFilter({ include: ['*.{json,txt}'] }, 'merges.txt')
        ).toBeTrue();
        expect(
            matchFileFilter(
                { include: ['model-0000[1-2].bin'] },
                'model-00002.bin'
            )
        ).toBeTrue();
        expect(
            matchFileFilter({ include: ['**/config.json'] }, 'config.json')
        ).toBeTrue();
    });

    it('detects exact names', () => {
        expect(isExactPattern('org/model')).toBeTrue();
        expect(isExactPattern('org/**')).toBeFalse();
        expect(isExactPattern('{a,b}')).toBeFalse();
        expect(isExactPattern('')).toBeFalse();
    });

    it('picks icons from the class, then the name', () => {
        expect(fileIcon('x', 'weight')).toBe('layers');
        expect(fileIcon('model.safetensors')).toBe('layers');
        expect(fileIcon('LICENSE')).toBe('certificate');
        expect(fileIcon('dir/README.md')).toBe('note');
        expect(fileIcon('unknown.xyz')).toBe('file');
    });
});
