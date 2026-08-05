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
import { TestBed, inject } from '@angular/core/testing';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { SignInService } from './sign-in.service';
import {
    provideHttpClient,
    withInterceptorsFromDi,
} from '@angular/common/http';

describe('SignInService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [],
            providers: [
                SignInService,
                provideHttpClient(withInterceptorsFromDi()),
                provideHttpClientTesting(),
            ],
        });
    });

    it('should be created', inject(
        [SignInService],
        (service: SignInService) => {
            expect(service).toBeTruthy();
        }
    ));
});
