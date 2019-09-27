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

import { NgModule } from "@angular/core";
import { CoreModule } from "../core/core.module";
import { SharedModule } from "../shared/shared.module";

import { ConfigurationComponent } from "./config.component";
import { ConfigurationService } from "./config.service";
import { ConfirmMessageHandler } from "./config.msg.utils";
import { ConfigurationAuthComponent } from "./auth/config-auth.component";
import { ConfigurationEmailComponent } from "./email/config-email.component";
import { RobotApiRepository } from "../project/robot-account/robot.api.repository";
import { ConfigurationScannerComponent } from "./scanner/config-scanner.component";
import { NewScannerModalComponent } from "./scanner/new-scanner-modal/new-scanner-modal.component";
import { NewScannerFormComponent } from "./scanner/new-scanner-form/new-scanner-form.component";
import { ConfigScannerService } from "./scanner/config-scanner.service";
import { ScannerMetadataComponent } from "./scanner/scanner-metadata/scanner-metadata.component";


@NgModule({
    imports: [CoreModule, SharedModule],
    declarations: [
        ConfigurationComponent,
        ConfigurationAuthComponent,
        ConfigurationEmailComponent,
        ConfigurationScannerComponent,
        NewScannerModalComponent,
        NewScannerFormComponent,
        ScannerMetadataComponent
    ],
    exports: [ConfigurationComponent],
    providers: [
        ConfigurationService,
        ConfirmMessageHandler,
        RobotApiRepository,
        ConfigScannerService,
    ]
})
export class ConfigurationModule {
}
