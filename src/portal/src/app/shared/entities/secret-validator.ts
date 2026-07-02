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

export interface SecretValidationError {
    code: 'LENGTH' | 'UPPERCASE' | 'LOWERCASE' | 'DIGIT';
    message: string;
}

export interface SecretValidationResult {
    isValid: boolean;
    errors: SecretValidationError[];
}

export class SecretValidator {
    static readonly MIN_LENGTH = 8;
    static readonly MAX_LENGTH = 128;

    static validate(secret: string): SecretValidationResult {
        const errors: SecretValidationError[] = [];

        if (!this.validateLength(secret)) {
            errors.push({
                code: 'LENGTH',
                message: 'ROBOT_ACCOUNT.SECRET_LENGTH_ERROR',
            });
        }

        if (!this.hasUppercase(secret)) {
            errors.push({
                code: 'UPPERCASE',
                message: 'ROBOT_ACCOUNT.SECRET_UPPERCASE_ERROR',
            });
        }

        if (!this.hasLowercase(secret)) {
            errors.push({
                code: 'LOWERCASE',
                message: 'ROBOT_ACCOUNT.SECRET_LOWERCASE_ERROR',
            });
        }

        if (!this.hasDigit(secret)) {
            errors.push({
                code: 'DIGIT',
                message: 'ROBOT_ACCOUNT.SECRET_DIGIT_ERROR',
            });
        }

        return {
            isValid: errors.length === 0,
            errors,
        };
    }

    static validateLength(secret: string): boolean {
        if (!secret) {
            return false;
        }
        return (
            secret.length >= this.MIN_LENGTH && secret.length <= this.MAX_LENGTH
        );
    }

    static hasUppercase(secret: string): boolean {
        return /[A-Z]/.test(secret);
    }

    static hasLowercase(secret: string): boolean {
        return /[a-z]/.test(secret);
    }

    static hasDigit(secret: string): boolean {
        return /[0-9]/.test(secret);
    }

    static getValidationErrors(secret: string): SecretValidationError[] {
        return this.validate(secret).errors;
    }
}
