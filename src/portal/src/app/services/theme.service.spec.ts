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
import { TestBed } from '@angular/core/testing';
import { DOCUMENT } from '@angular/common';
import { ThemeService } from './theme.service';

describe('ThemeService', () => {
    let service: ThemeService;
    let document: Document;

    beforeEach(() => {
        TestBed.configureTestingModule({});
        service = TestBed.inject(ThemeService);
        document = TestBed.inject(DOCUMENT);
    });

    it('should be created', () => {
        expect(service).toBeTruthy();
    });

    it('loadStyle should set cds-theme to "dark" for dark-theme stylesheets', () => {
        service.loadStyle('dark-theme.css');
        expect(document.body.getAttribute('cds-theme')).toBe('dark');
    });

    it('loadStyle should set cds-theme to "light" for non-dark stylesheets', () => {
        service.loadStyle('light-theme.css');
        expect(document.body.getAttribute('cds-theme')).toBe('light');
    });

    it('loadStyle should create a link element when no existing theme link is present', () => {
        const existingLink = document.getElementById('client-theme');
        if (existingLink) {
            existingLink.remove();
        }
        service.loadStyle('light-theme.css');
        const link = document.getElementById('client-theme') as HTMLLinkElement;
        expect(link).toBeTruthy();
        expect(link.rel).toBe('stylesheet');
        expect(link.href).toContain('light-theme.css');
    });

    it('loadStyle should update href on existing theme link', () => {
        // First call creates the link
        service.loadStyle('light-theme.css');
        // Second call should update href on the existing element
        service.loadStyle('dark-theme.css');
        const link = document.getElementById('client-theme') as HTMLLinkElement;
        expect(link.href).toContain('dark-theme.css');
    });
});
