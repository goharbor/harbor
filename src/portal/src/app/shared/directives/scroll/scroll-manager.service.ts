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
import { ScrollSectionDirective } from './scroll-section.directive';
import { Injectable } from '@angular/core';

@Injectable({
    providedIn: 'root',
})
export class ScrollManagerService {
    private sections = new Map<string | number, ScrollSectionDirective>();

    scroll(id: string | number) {
        this.sections.get(id)!.scroll();
    }

    register(section: ScrollSectionDirective) {
        this.sections.set(section.id, section);
    }

    remove(section: ScrollSectionDirective) {
        this.sections.delete(section.id);
    }
}
