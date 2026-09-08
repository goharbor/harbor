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
import { Instance } from '../../../../../ng-swagger-gen/models/instance';

export class AuthMode {
    static NONE = 'NONE';
    static BASIC = 'BASIC';
    static OAUTH = 'OAUTH';
    static CUSTOM = 'CUSTOM';
}

export enum PreheatingStatusEnum {
    // front status
    NOT_PREHEATED = 'NOT_PREHEATED',
    // back-end status
    PENDING = 'PENDING',
    RUNNING = 'RUNNING',
    SUCCESS = 'SUCCESS',
    FAIL = 'FAIL',
}

export interface FrontInstance extends Instance {
    hasCheckHealth?: boolean;
    pingStatus?: string;
}

export const HEALTHY: string = 'Healthy';
export const UNHEALTHY: string = 'Unhealthy';
