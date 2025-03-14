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
import { TopRepoComponent } from './top-repo.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { of } from 'rxjs';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { TopRepoService } from './top-repository.service';
import { SharedTestingModule } from '../../../shared/shared.module';

describe('TopRepoComponent', () => {
    let component: TopRepoComponent;
    let fixture: ComponentFixture<TopRepoComponent>;
    const mockMessageHandlerService = {
        showSuccess: () => {},
        handleError: () => {},
        isAppLevel: () => {},
    };
    const mockTopRepoService = {
        getTopRepos: () => of([]),
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [TopRepoComponent],
            providers: [
                { provide: TopRepoService, useValue: mockTopRepoService },
                {
                    provide: MessageHandlerService,
                    useValue: mockMessageHandlerService,
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(TopRepoComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
