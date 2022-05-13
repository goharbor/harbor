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

import { NgModule } from '@angular/core';
import { SharedModule } from '../../../shared/shared.module';
import { NewScannerModalComponent } from './scanner/new-scanner-modal/new-scanner-modal.component';
import { ScannerMetadataComponent } from './scanner/scanner-metadata/scanner-metadata.component';
import { NewScannerFormComponent } from './scanner/new-scanner-form/new-scanner-form.component';
import { RouterModule, Routes } from '@angular/router';
import { ConfigurationScannerComponent } from './scanner/config-scanner.component';
import { VulnerabilityConfigComponent } from './vulnerability/vulnerability-config.component';
import { InterrogationServicesComponent } from './interrogation-services.component';
import { ScanAllRepoService } from './vulnerability/scanAll.service';
import {
    ScanApiDefaultRepository,
    ScanApiRepository,
} from './vulnerability/scanAll.api.repository';

const routes: Routes = [
    {
        path: '',
        component: InterrogationServicesComponent,
        children: [
            {
                path: 'scanners',
                component: ConfigurationScannerComponent,
            },
            {
                path: 'vulnerability',
                component: VulnerabilityConfigComponent,
            },
            {
                path: '',
                redirectTo: 'scanners',
                pathMatch: 'full',
            },
        ],
    },
];
@NgModule({
    imports: [SharedModule, RouterModule.forChild(routes)],
    declarations: [
        NewScannerModalComponent,
        NewScannerFormComponent,
        ScannerMetadataComponent,
        ConfigurationScannerComponent,
        InterrogationServicesComponent,
        VulnerabilityConfigComponent,
    ],
    providers: [
        ScanAllRepoService,
        { provide: ScanApiRepository, useClass: ScanApiDefaultRepository },
    ],
})
export class InterrogationServicesModule {}
