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
import { Component, OnInit } from '@angular/core';
import { MessageHandlerService } from '../../shared/services/message-handler.service';

@Component({
    selector: 'api-tokens-modal',
    templateUrl: 'api-tokens-modal.component.html',
    styleUrls: ['./api-tokens-modal.component.scss'],
    standalone: false,
})
export class ApiTokensModalComponent implements OnInit {
    opened = false;
    staticBackdrop = false;

    tokens: any[] = [];
    selectedTokens: any[] = [];
    tokenLoading = false;

    showCreateTokenModal = false;
    createdTokenSecret: string;
    newTokenForm: any = {
        name: '',
        description: '',
    };

    constructor(private msgHandler: MessageHandlerService) {}

    ngOnInit(): void {
        this.loadTokens();
    }

    open(): void {
        this.opened = true;
        this.loadTokens();
    }

    close(): void {
        this.opened = false;
        this.resetForm();
    }

    loadTokens(): void {
        // TODO: Implement API call to fetch user tokens
        // this.tokenLoading = true;
        // this.userService.getUserTokens().subscribe(
        //   (tokens) => {
        //     this.tokens = tokens;
        //     this.tokenLoading = false;
        //   },
        //   (error) => {
        //     this.msgHandler.showError('Failed to load tokens', {});
        //     this.tokenLoading = false;
        //   }
        // );
    }

    openCreateTokenModal(): void {
        this.showCreateTokenModal = true;
        this.resetForm();
    }

    closeCreateTokenModal(): void {
        this.showCreateTokenModal = false;
        this.resetForm();
    }

    createToken(): void {
        if (!this.newTokenForm.name) {
            this.msgHandler.showError('Token name is required', {});
            return;
        }

        // TODO: Implement API call to create token
        // this.tokenLoading = true;
        // this.userService.createUserToken(this.newTokenForm).subscribe(
        //   (response) => {
        //     this.createdTokenSecret = response.secret;
        //     this.tokens.push(response);
        //     this.tokenLoading = false;
        //   },
        //   (error) => {
        //     this.msgHandler.showError('Failed to create token', {});
        //     this.tokenLoading = false;
        //   }
        // );
    }

    copyTokenSecret(): void {
        const copyInput = document.createElement('textarea');
        copyInput.value = this.createdTokenSecret;
        document.body.appendChild(copyInput);
        copyInput.select();
        document.execCommand('copy');
        document.body.removeChild(copyInput);
        this.msgHandler.showSuccess('Token copied to clipboard');
    }

    revokeToken(tokenId: string): void {
        // TODO: Implement API call to revoke/disable token
    }

    deleteToken(tokenId: string): void {
        // TODO: Implement API call to delete token
    }

    formatScope(scope: any): string {
        if (!scope) {
            return 'All';
        }
        if (typeof scope === 'string') {
            return scope;
        }
        if (Array.isArray(scope)) {
            return scope.join(', ');
        }
        return JSON.stringify(scope);
    }

    private resetForm(): void {
        this.newTokenForm = {
            name: '',
            description: '',
        };
        this.createdTokenSecret = '';
    }
}
