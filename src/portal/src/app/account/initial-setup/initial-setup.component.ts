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
import { Component, OnInit, ViewChild } from '@angular/core';
import { Router } from '@angular/router';
import { NgForm } from '@angular/forms';
import { SetupService } from '../../services/setup.service';
import { SkinableConfig } from '../../services/skinable-config.service';
import { CommonRoutes } from '../../shared/entities/shared.const';

const STATUS_NORMAL = 0;
const STATUS_ONGOING = 1;
const STATUS_ERROR = -1;
const STATUS_SUCCESS = 2;

@Component({
    selector: 'initial-setup',
    templateUrl: './initial-setup.component.html',
    styleUrls: ['./initial-setup.component.scss'],
})
export class InitialSetupComponent implements OnInit {
    showPwd: boolean = false;
    showConfirmPwd: boolean = false;
    password: string = '';
    confirmPassword: string = '';
    status: number = STATUS_NORMAL;
    errorMessage: string = '';
    customLoginBgImg: string;
    customAppTitle: string;

    @ViewChild('setupForm', { static: true }) currentForm: NgForm;

    constructor(
        private router: Router,
        private setupService: SetupService,
        private skinableConfig: SkinableConfig
    ) {}

    ngOnInit(): void {
        // custom skin
        let customSkinObj = this.skinableConfig.getSkinConfig();
        if (customSkinObj) {
            if (customSkinObj.loginBgImg) {
                this.customLoginBgImg = customSkinObj.loginBgImg;
            }
            if (customSkinObj.loginTitle) {
                this.customAppTitle = customSkinObj.loginTitle;
            }
        }

        // Check if setup is required; if not, redirect to sign-in
        this.setupService.clearCache();
        this.setupService.isSetupRequired().subscribe(required => {
            if (!required) {
                this.router.navigateByUrl(CommonRoutes.EMBEDDED_SIGN_IN);
            }
        });
    }

    get isOnGoing(): boolean {
        return this.status === STATUS_ONGOING;
    }

    get isError(): boolean {
        return this.status === STATUS_ERROR;
    }

    get isSuccess(): boolean {
        return this.status === STATUS_SUCCESS;
    }

    get isValid(): boolean {
        return (
            this.currentForm &&
            this.currentForm.valid &&
            this.password === this.confirmPassword &&
            this.isPasswordStrong()
        );
    }

    get passwordMismatch(): boolean {
        return (
            this.confirmPassword.length > 0 &&
            this.password !== this.confirmPassword
        );
    }

    isPasswordStrong(): boolean {
        if (
            !this.password ||
            this.password.length < 8 ||
            this.password.length > 128
        ) {
            return false;
        }
        const hasLower = /[a-z]/.test(this.password);
        const hasUpper = /[A-Z]/.test(this.password);
        const hasNumber = /[0-9]/.test(this.password);
        return hasLower && hasUpper && hasNumber;
    }

    submitSetup(): void {
        if (!this.isValid || this.isOnGoing) {
            return;
        }

        this.status = STATUS_ONGOING;
        this.errorMessage = '';

        this.setupService.setupAdminPassword(this.password).subscribe(
            () => {
                this.status = STATUS_SUCCESS;
                // Redirect to sign-in page after a brief delay
                setTimeout(() => {
                    this.router.navigateByUrl(CommonRoutes.EMBEDDED_SIGN_IN);
                }, 1500);
            },
            error => {
                this.status = STATUS_ERROR;
                if (error.status === 403) {
                    this.errorMessage = 'INITIAL_SETUP.ERROR_ALREADY_COMPLETED';
                } else if (error.status === 400) {
                    this.errorMessage = 'INITIAL_SETUP.ERROR_WEAK_PASSWORD';
                } else {
                    this.errorMessage = 'INITIAL_SETUP.ERROR_GENERIC';
                }
            }
        );
    }
}
