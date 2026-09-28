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
    validateRepositoryFilterPattern,
    checkBalancedChars,
} from './repository-filter.util';

describe('repository-filter.util', () => {
    describe('validateRepositoryFilterPattern', () => {
        it('should return null for empty pattern', () => {
            expect(validateRepositoryFilterPattern('regex', '')).toBeNull();
            expect(
                validateRepositoryFilterPattern('doublestar', null)
            ).toBeNull();
        });

        it('should return null for valid regex', () => {
            expect(validateRepositoryFilterPattern('regex', '^.*$')).toBeNull();
        });

        it('should return error key for invalid regex', () => {
            expect(validateRepositoryFilterPattern('regex', '[')).toBe(
                'PROJECT.REPOSITORY_FILTER_REGEX_INVALID'
            );
        });

        it('should return null for valid doublestar', () => {
            expect(
                validateRepositoryFilterPattern('doublestar', 'library/**')
            ).toBeNull();
            expect(
                validateRepositoryFilterPattern('doublestar', '{a,b}/**')
            ).toBeNull();
        });

        it('should return error key for unbalanced brackets in doublestar', () => {
            expect(
                validateRepositoryFilterPattern('doublestar', '{a,b/**')
            ).toBe('PROJECT.REPOSITORY_FILTER_DOUBLESTAR_INVALID');
            expect(
                validateRepositoryFilterPattern('doublestar', '[a-z/**')
            ).toBe('PROJECT.REPOSITORY_FILTER_DOUBLESTAR_INVALID');
        });

        it('should return null for unknown kind', () => {
            expect(
                validateRepositoryFilterPattern('unknown', '[unbalanced')
            ).toBeNull();
        });
    });

    describe('checkBalancedChars', () => {
        it('should return true for balanced brackets', () => {
            expect(checkBalancedChars('[]')).toBeTrue();
            expect(checkBalancedChars('{}')).toBeTrue();
            expect(checkBalancedChars('[{}]')).toBeTrue();
            expect(checkBalancedChars('foo{bar[baz]}')).toBeTrue();
        });

        it('should return false for unbalanced brackets', () => {
            expect(checkBalancedChars('[')).toBeFalse();
            expect(checkBalancedChars('{')).toBeFalse();
            expect(checkBalancedChars(']')).toBeFalse();
            expect(checkBalancedChars('}')).toBeFalse();
            expect(checkBalancedChars('[{]')).toBeFalse();
            expect(checkBalancedChars('[}]')).toBeFalse();
            expect(checkBalancedChars('][')).toBeFalse();
        });
    });
});
