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
export const enum DeletionTargets {
  EMPTY, PROJECT, PROJECT_MEMBER, USER, POLICY, TARGET, REPOSITORY, TAG
};
export const harborRootRoute = "/harbor/dashboard";
export const signInRoute = "/sign-in";

export const enum ActionType {
  ADD_NEW, EDIT
};

export const ListMode = {
  READONLY: "readonly",
  FULL: "full"
};