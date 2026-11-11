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
import { ClipboardDirective } from './clipboard.directive';
import { CLIPBOARD_SERVICE_PROVIDER } from './clipboard.service';
import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { WindowTokenModule } from '../ngx-window-token/window-token';
export * from './clipboard.directive';
export * from './clipboard.service';

@NgModule({
    imports: [CommonModule, WindowTokenModule],
    declarations: [ClipboardDirective],
    exports: [ClipboardDirective, WindowTokenModule],
    providers: [CLIPBOARD_SERVICE_PROVIDER],
})
export class ClipboardModule {}
