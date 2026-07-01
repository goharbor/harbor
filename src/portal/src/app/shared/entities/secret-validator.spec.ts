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

import { SecretValidator, SecretValidationError } from './secret-validator';

describe('SecretValidator', () => {
    describe('validate()', () => {
        it('should return valid for a secret meeting all requirements', () => {
            const result = SecretValidator.validate('TestSecret123');
            expect(result.isValid).toBe(true);
            expect(result.errors.length).toBe(0);
        });

        it('should return invalid for an empty secret', () => {
            const result = SecretValidator.validate('');
            expect(result.isValid).toBe(false);
            expect(result.errors.length).toBeGreaterThan(0);
        });

        it('should detect length errors', () => {
            const tooShort = SecretValidator.validate('Test1a');
            expect(tooShort.isValid).toBe(false);
            expect(
                tooShort.errors.find(e => e.code === 'LENGTH')
            ).toBeDefined();

            const tooLong = SecretValidator.validate('A' + 'a1'.repeat(65));
            expect(tooLong.isValid).toBe(false);
            expect(tooLong.errors.find(e => e.code === 'LENGTH')).toBeDefined();
        });

        it('should detect missing uppercase', () => {
            const result = SecretValidator.validate('testsecret1');
            expect(result.isValid).toBe(false);
            expect(
                result.errors.find(e => e.code === 'UPPERCASE')
            ).toBeDefined();
        });

        it('should detect missing lowercase', () => {
            const result = SecretValidator.validate('TESTSECRET1');
            expect(result.isValid).toBe(false);
            expect(
                result.errors.find(e => e.code === 'LOWERCASE')
            ).toBeDefined();
        });

        it('should detect missing digit', () => {
            const result = SecretValidator.validate('TestSecret');
            expect(result.isValid).toBe(false);
            expect(result.errors.find(e => e.code === 'DIGIT')).toBeDefined();
        });

        it('should allow uppercase at different positions', () => {
            expect(SecretValidator.validate('TestSecret123').isValid).toBe(
                true
            );
            expect(SecretValidator.validate('tESTsecret123').isValid).toBe(
                true
            );
            expect(SecretValidator.validate('testsecretA123').isValid).toBe(
                true
            );
        });

        it('should allow digits at different positions', () => {
            expect(SecretValidator.validate('1TestSecret23').isValid).toBe(
                true
            );
            expect(SecretValidator.validate('TestSecret1').isValid).toBe(true);
            expect(SecretValidator.validate('Test9Secret').isValid).toBe(true);
        });

        it('should return multiple errors for invalid secret', () => {
            const result = SecretValidator.validate('test');
            expect(result.isValid).toBe(false);
            expect(result.errors.length).toBeGreaterThanOrEqual(2);
        });

        it('should validate edge case: exactly MIN_LENGTH (8)', () => {
            expect(SecretValidator.validate('TestSec1').isValid).toBe(true);
        });

        it('should validate edge case: exactly MAX_LENGTH (128)', () => {
            const secret = 'A' + 'a'.repeat(126) + '1';
            expect(SecretValidator.validate(secret).isValid).toBe(true);
        });

        it('should invalidate just below MIN_LENGTH (7)', () => {
            const result = SecretValidator.validate('TestSe1');
            expect(result.isValid).toBe(false);
            expect(result.errors.find(e => e.code === 'LENGTH')).toBeDefined();
        });

        it('should invalidate just above MAX_LENGTH (129)', () => {
            const secret = 'A' + 'a'.repeat(127) + '1';
            const result = SecretValidator.validate(secret);
            expect(result.isValid).toBe(false);
            expect(result.errors.find(e => e.code === 'LENGTH')).toBeDefined();
        });
    });

    describe('validateLength()', () => {
        it('should return true for valid length', () => {
            expect(SecretValidator.validateLength('TestSecret123')).toBe(true);
            expect(SecretValidator.validateLength('TestSec1')).toBe(true);
        });

        it('should return false for empty string', () => {
            expect(SecretValidator.validateLength('')).toBe(false);
        });

        it('should return false for too short', () => {
            expect(SecretValidator.validateLength('Test1')).toBe(false);
        });

        it('should return false for too long', () => {
            const longSecret = 'A'.repeat(129);
            expect(SecretValidator.validateLength(longSecret)).toBe(false);
        });
    });

    describe('hasUppercase()', () => {
        it('should return true when uppercase is present', () => {
            expect(SecretValidator.hasUppercase('TestSecret')).toBe(true);
            expect(SecretValidator.hasUppercase('A')).toBe(true);
            expect(SecretValidator.hasUppercase('testA')).toBe(true);
        });

        it('should return false when no uppercase', () => {
            expect(SecretValidator.hasUppercase('testsecret')).toBe(false);
            expect(SecretValidator.hasUppercase('123')).toBe(false);
            expect(SecretValidator.hasUppercase('')).toBe(false);
        });
    });

    describe('hasLowercase()', () => {
        it('should return true when lowercase is present', () => {
            expect(SecretValidator.hasLowercase('TestSecret')).toBe(true);
            expect(SecretValidator.hasLowercase('a')).toBe(true);
            expect(SecretValidator.hasLowercase('TESTa')).toBe(true);
        });

        it('should return false when no lowercase', () => {
            expect(SecretValidator.hasLowercase('TESTSECRET')).toBe(false);
            expect(SecretValidator.hasLowercase('123')).toBe(false);
            expect(SecretValidator.hasLowercase('')).toBe(false);
        });
    });

    describe('hasDigit()', () => {
        it('should return true when digit is present', () => {
            expect(SecretValidator.hasDigit('TestSecret1')).toBe(true);
            expect(SecretValidator.hasDigit('9')).toBe(true);
            expect(SecretValidator.hasDigit('TEST5secret')).toBe(true);
        });

        it('should return false when no digit', () => {
            expect(SecretValidator.hasDigit('TestSecret')).toBe(false);
            expect(SecretValidator.hasDigit('abc')).toBe(false);
            expect(SecretValidator.hasDigit('')).toBe(false);
        });
    });

    describe('getValidationErrors()', () => {
        it('should return empty array for valid secret', () => {
            const errors = SecretValidator.getValidationErrors('TestSecret1');
            expect(errors.length).toBe(0);
        });

        it('should return all errors for invalid secret', () => {
            const errors = SecretValidator.getValidationErrors('test');
            expect(errors.length).toBeGreaterThan(0);
            const codes = errors.map(e => e.code);
            expect(codes).toContain('LENGTH');
            expect(codes).toContain('UPPERCASE');
            expect(codes).toContain('DIGIT');
        });

        it('should include error message keys for translation', () => {
            const errors = SecretValidator.getValidationErrors('test');
            errors.forEach(error => {
                expect(error.message).toMatch(/^ROBOT_ACCOUNT\./);
            });
        });
    });

    describe('edge cases', () => {
        it('should handle special characters correctly', () => {
            expect(SecretValidator.validate('Test!@#$%Secret1').isValid).toBe(
                true
            );
            expect(SecretValidator.validate('Test-_=+Secret1').isValid).toBe(
                true
            );
        });

        it('should handle unicode characters correctly', () => {
            expect(SecretValidator.validate('TestSecret1αβγ').isValid).toBe(
                true
            );
        });

        it('should handle spaces in secret', () => {
            expect(SecretValidator.validate('Test Secret 1').isValid).toBe(
                true
            );
        });
    });
});
