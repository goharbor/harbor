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
import { Component, Input, ViewChild } from '@angular/core';
import { CopyInputComponent } from './copy-input.component';
import { InlineAlertComponent } from '../inline-alert/inline-alert.component';

@Component({
    selector: 'hbr-push-image-button',
    templateUrl: './push-image.component.html',
    styleUrls: ['./push-image.scss'],

    providers: [],
})
export class PushImageButtonComponent {
    @Input() registryUrl: string = 'unknown';
    @Input() projectName: string = 'unknown';

    @ViewChild('tagCopyImage') tagCopyImageInput: CopyInputComponent;
    @ViewChild('pushCopyImage') pushCopImageyInput: CopyInputComponent;
    @ViewChild('tagCopyChart') tagCopyChartInput: CopyInputComponent;
    @ViewChild('pushCopyChart') pushCopyChartInput: CopyInputComponent;
    @ViewChild('pushCopyCnab') pushCopCnabyInput: CopyInputComponent;
    @ViewChild('copyAlert') copyAlert: InlineAlertComponent;

    public get tagCommandImage(): string {
        return `docker tag SOURCE_IMAGE[:TAG] ${this.registryUrl}/${this.projectName}/REPOSITORY[:TAG]`;
    }

    public get pushCommandImage(): string {
        return `docker push ${this.registryUrl}/${this.projectName}/REPOSITORY[:TAG]`;
    }

    public get podmanPushCommand(): string {
        return `podman push IMAGE_ID ${this.registryUrl}/${this.projectName}/REPOSITORY[:TAG]`;
    }

    public get tagCommandChart(): string {
        return `helm package CHART_PATH`;
    }

    public get pushCommandChart(): string {
        return `helm push CHART_PACKAGE oci://${this.registryUrl}/${this.projectName}`;
    }

    public get pushCommandCnab(): string {
        return `cnab-to-oci push CNAB_PATH --target ${this.registryUrl}/${this.projectName}/REPOSITORY[:TAG] --auto-update-bundle`;
    }

    onclick(): void {
        if (this.tagCopyImageInput) {
            this.tagCopyImageInput.reset();
        }

        if (this.pushCopImageyInput) {
            this.pushCopImageyInput.reset();
        }
        if (this.tagCopyChartInput) {
            this.tagCopyChartInput.reset();
        }

        if (this.pushCopyChartInput) {
            this.pushCopyChartInput.reset();
        }

        if (this.pushCopCnabyInput) {
            this.pushCopCnabyInput.reset();
        }

        if (this.copyAlert) {
            this.copyAlert.close();
        }
    }

    onCpError($event: any): void {
        if (this.copyAlert) {
            this.copyAlert.showInlineError('PUSH_IMAGE.COPY_ERROR');
        }
    }
}
