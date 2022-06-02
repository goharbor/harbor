```text
{
  "name": "harbor",
  "version": "2.5.0",
  "description": "Harbor UI with Clarity",
  "angular-cli": {},
  "scripts": {
  
    // triggered after running "npm install"
    "postinstall": "node scripts/convert-yaml-to-json.js && ng-swagger-gen -i ng-swagger-gen/swagger.json -o ng-swagger-gen && node scripts/delete-swagger-json.js",
    
    // For developing
    "start": "node --max_old_space_size=2048 ./node_modules/@angular/cli/bin/ng serve --ssl true --host 0.0.0.0 --proxy-config proxy.config.json",
    "start:prod": "node --max_old_space_size=2048 ./node_modules/@angular/cli/bin/ng serve --ssl true --host 0.0.0.0 --proxy-config proxy.config.json --configuration production",
    
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
    "pree2e": "webdriver-manager update",
    "e2e": "protractor",
    
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
    "@angular/animations": "~13.3.0",
    "@angular/common": "~13.3.0",
    "@angular/compiler": "~13.3.0",
    "@angular/core": "~13.3.0",
    "@angular/forms": "~13.3.0",
    "@angular/localize": "~13.3.0",
    "@angular/platform-browser": "~13.3.0",
    "@angular/platform-browser-dynamic": "~13.3.0",
    "@angular/router": "~13.2.2",
    "rxjs": "^7.4.0",
    "tslib": "^2.2.0",
    "zone.js": "~0.11.4",
    
    // Clarity UI. Required
    "@clr/angular": "13.0.2",
    "@cds/core": "5.6.4",
    "@clr/icons": "13.0.2",
    "@clr/ui": "13.0.2",
    
    // For Harbor i18n functionality. Required
    "@ngx-translate/core": "^13.0.0",
    "@ngx-translate/http-loader": "^6.0.0",
    
    // For cron string checking. Required
    "cron-validator": "^1.2.1",
    
    // Used by CopyInputComponent to copy pull command to clipboard. Required
    "ngx-clipboard": "^12.3.1",
    
    // For Harbor cookie service. Required
    "ngx-cookie": "^5.0.2",
    
    // To render markdown data. Required
    "ngx-markdown": "~13.0.0",
    
    // For swagger API center. Required
    "swagger-ui": "~4.9.0",
    "buffer": "^6.0.3",
    "stream-browserify": "^3.0.0",
    
    // To convert yaml to json. Required
    "js-yaml": "^4.1.0"
  },
  "devDependencies": {
    // Angular framework. Required
    "@angular-devkit/build-angular": "~13.3.0",
    "@angular/cli": "~13.3.0",
    "@angular/compiler-cli": "~13.3.0",
    "@types/jasmine": "~3.10.1",
    "@types/node": "^16.11.6",
    "typescript": "~4.5.4",
    
    // For unit test. Required
    "jasmine-core": "^4.0.0",
    "jasmine-spec-reporter": "~7.0.0",
    "karma": "^6.3.17",
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
    "protractor": "^7.0.0",
    
    // For code grammar checking. Optional
    "eslint": "^8.12.0",
    "@angular-eslint/eslint-plugin": "13.2.1",
    "@angular-eslint/eslint-plugin-template": "13.2.1",
    "@angular-eslint/schematics": "13.2.1",
    "@angular-eslint/template-parser": "13.2.1",
    "@typescript-eslint/eslint-plugin": "5.17.0",
    "@typescript-eslint/parser": "5.17.0",
    
    // For code format
    "prettier": "^2.6.2",
    "prettier-eslint": "^14.0.2",
    "eslint-config-prettier": "^8.5.0",
    "eslint-plugin-prettier": "^4.0.0",
  }
}
```
