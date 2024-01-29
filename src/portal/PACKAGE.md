```text
{
  "name": "harbor",
  "version": "2.10.0",
  "description": "Harbor UI with Clarity",
  "angular-cli": {},
  "scripts": {

    // triggered after running "npm install"
    "postinstall": "node scripts/convert-yaml-to-json.js && ng-swagger-gen -i ng-swagger-gen/swagger.json -o ng-swagger-gen && node scripts/delete-swagger-json.js",

    // For developing
    "start": "node --max_old_space_size=2048 ./node_modules/@angular/cli/bin/ng serve --ssl true --host 0.0.0.0 --proxy-config proxy.config.json",
    "start:prod": "node --max_old_space_size=2048 ./node_modules/@angular/cli/bin/ng serve --ssl true --host 0.0.0.0 --proxy-config proxy.config.json --configuration production",
    "start_default_port": "node --max_old_space_size=2048 ./node_modules/@angular/cli/bin/ng serve --ssl true --host 0.0.0.0 --port 443 --disable-host-check --proxy-config proxy.config.json",
    
    // For code grammar checking
    "lint": "tslint \"src/**/*.ts\"",
    "lint_fix": "tslint --fix \"src/**/*.ts\"",

    // For unit test
    "test": "node --max_old_space_size=2048 ./node_modules/@angular/cli/bin/ng test --code-coverage",
    "test:watch": "ng test --code-coverage --watch",
    "test:debug": "ng test --code-coverage --source-map false",
    "test:chrome": "ng test --code-coverage --browsers Chrome",
    "test:headless": "ng test --watch=false --no-progress --code-coverage --browsers=ChromeNoSandboxHeadless",
    "test:chrome-debug": "ng test --code-coverage --browsers Chrome --watch",

     // E2e related. Currently not used
    "e2e": "ng e2e",

    "build": "ng build --aot",
    "release": "ng build --configuration production",

    "build-mock-api-server": "tsc -p server",

    // to run a mocked node express api server
    "mock-api-server": "npm run build-mock-api-server && node server/dist/server/src/mock-api.js",


    // Run this command before the production building. It will set the current timestamp to "buildTimestamp" property in "environment.prod.ts" file
    // And "buildTimestamp" will be used as a query parameter to avoid browser cache after upgrading Harbor UI
    "generate-build-timestamp": "node scripts/generate-build-timestamp.js"

  },
  "private": true,
  "dependencies": {
     // Angular framework. Required
    "@angular/animations": "^16.0.2",
    "@angular/common": "^16.0.2",
    "@angular/compiler": "^16.0.2,
    "@angular/core": "^16.0.2",
    "@angular/forms": "^16.0.2",
    "@angular/localize": "^16.0.2",
    "@angular/platform-browser": "^16.0.2",
    "@angular/platform-browser-dynamic": "^16.0.2",
    "@angular/router": "^16.0.2",
    "rxjs": "^7.4.0",
    "tslib": "^2.2.0",
    "zone.js": "~0.13.0",

    // Clarity UI. Required
    "@cds/core": "6.4.2",
    "@clr/angular": "15.4.0",
    "@clr/icons": "13.0.2",
    "@clr/ui": "15.4.0",

    // For Harbor i18n functionality. Required
    "@ngx-translate/core": "15.0.0",
    "@ngx-translate/http-loader": "8.0.0",

    // For cron string checking. Required
    "cron-validator": "^1.2.1",

    // Used by CopyInputComponent to copy pull command to clipboard. Required
    "ngx-clipboard": "^12.3.1",

    // For Harbor cookie service. Required
    "ngx-cookie": "^5.0.2",

    // To render markdown data. Required
    "ngx-markdown": "16.0.0",

    // To convert yaml to json. Required
    "js-yaml": "^4.1.0"
  },
  "devDependencies": {
    // Angular framework. Required
    "@angular-devkit/build-angular": "^16.0.2",
    "@angular/cli": "^16.0.2",
    "@angular/compiler-cli": "^16.0.2",
    "@types/jasmine": "~4.3.0",
    "@types/node": "^16.11.6",
    "typescript": "~5.0.4",

    // For unit test. Required
    "jasmine-core": "~4.5.0",
    "jasmine-spec-reporter": "~7.0.0",
    "karma": "^6.4.2",
    "karma-chrome-launcher": "~3.1.0",
    "karma-coverage": "^2.2.0",
    "karma-jasmine": "~4.0.1",
    "karma-jasmine-html-reporter": "^1.7.0",

    // To run a local mocked API server. Required
    "@types/express": "^4.17.12",
    "express": "^4.17.1",

    // To generate models and Angular services based on swagger.yaml. Required
    // do not upgrade it
    "ng-swagger-gen": "^1.8.1",

    // For e2e test. Required
    "cypress": "12.12.0"

    // For code grammar checking. Optional
    "eslint": "^8.12.0",
    "@angular-eslint/eslint-plugin": "15.1.0",
    "@angular-eslint/eslint-plugin-template": "15.1.0",
    "@angular-eslint/schematics": "15.1.0",
    "@angular-eslint/template-parser": "15.1.0",
    "@typescript-eslint/eslint-plugin": "5.29.0",
    "@typescript-eslint/parser": "5.29.0",

    // For code format
    "prettier": "^2.6.2",
    "prettier-eslint": "^14.0.2",
    "eslint-config-prettier": "^8.5.0",
    "eslint-plugin-prettier": "^4.0.0",
    "stylelint": "^14.8.5",
    "stylelint-config-prettier": "^9.0.3",
    "stylelint-config-prettier-scss": "^0.0.1",
    "stylelint-config-standard": "^25.0.0",
    "stylelint-config-standard-scss": "^4.0.0",
  }
}
```
