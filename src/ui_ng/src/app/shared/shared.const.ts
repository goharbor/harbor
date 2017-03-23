export const supportedLangs = ['en', 'zh'];
export const enLang = "en";
export const languageNames = {
  "en": "English",
  "zh": "中文简体"
};
export const enum AlertType {
  DANGER, WARNING, INFO, SUCCESS
};

export const dismissInterval = 15 * 1000;
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
  CONFIG_ROUTE
};

export const enum ActionType {
  ADD_NEW, EDIT
};

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