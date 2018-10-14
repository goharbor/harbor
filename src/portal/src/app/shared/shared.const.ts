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
export const supportedLangs = ['en-us', 'zh-cn', 'es-es', 'fr-fr', 'pt-br'];
export const enLang = "en-us";
export const languageNames = {
  "en-us": "English",
  "zh-cn": "中文简体",
  "es-es": "Español",
  "fr-fr": "Français",
  "pt-br": "Português do Brasil"
};
export const enum AlertType {
  DANGER, WARNING, INFO, SUCCESS
}

export const dismissInterval = 10 * 1000;

export const httpStatusCode = {
  "Unauthorized": 401,
  "Forbidden": 403
};

export const enum ConfirmationTargets {
  EMPTY,
  PROJECT,
  PROJECT_MEMBER,
  USER,
  POLICY,
  TOGGLE_CONFIRM,
  TARGET,
  REPOSITORY,
  TAG,
  CONFIG,
  CONFIG_ROUTE,
  CONFIG_TAB
}

export const enum ActionType {
  ADD_NEW, EDIT
}

export const ListMode = {
  READONLY: "readonly",
  FULL: "full"
};


export const CommonRoutes = {
  SIGN_IN: "/sign-in",
  EMBEDDED_SIGN_IN: "/harbor/sign-in",
  SIGN_UP: "/sign-in?sign_up=true",
  EMBEDDED_SIGN_UP: "/harbor/sign-in?sign_up=true",
  HARBOR_ROOT: "/harbor",
  HARBOR_DEFAULT: "/harbor/projects"
};

export const AdmiralQueryParamKey = "admiral_redirect_url";
export const HarborQueryParamKey = "harbor_redirect_url";
export const CookieKeyOfAdmiral = "admiral.endpoint.latest";

export const enum ConfirmationState {
  NA, CONFIRMED, CANCEL
}
export const enum ConfirmationButtons {
  CONFIRM_CANCEL, YES_NO, DELETE_CANCEL, CLOSE, SWITCH_CANCEL
}

export const ProjectTypes = { 0: 'PROJECT.ALL_PROJECTS', 1: 'PROJECT.PRIVATE_PROJECTS', 2: 'PROJECT.PUBLIC_PROJECTS' };
export const RoleInfo = { 1: 'MEMBER.PROJECT_ADMIN', 2: 'MEMBER.DEVELOPER', 3: 'MEMBER.GUEST' };
export const RoleMapping = { 'projectAdmin': 'MEMBER.PROJECT_ADMIN', 'developer': 'MEMBER.DEVELOPER', 'guest': 'MEMBER.GUEST' };
export const ProjectRoles = [
  { id: 1, value: "MEMBER.PROJECT_ADMIN" },
  { id: 2, value: "MEMBER.DEVELOPER" },
  { id: 3, value: "MEMBER.GUEST" }
];

export enum Roles {
  PROJECT_ADMIN = 1,
  DEVELOPER = 2,
  GUEST = 3,
  OTHER = 0,
}
