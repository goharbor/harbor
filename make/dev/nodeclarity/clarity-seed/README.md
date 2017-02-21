![Clarity](logo.png)

Clarity Seed
============
This is a seed project for Angular 2 applications using [Clarity](https://github.com/vmware/clarity). For more information on the Clarity Design System, visit the [Clarity website](https://vmware.github.io/clarity/).

We offer this seed project in three different build systems:

1. **Angular-CLI version (branch: master)**

2. Webpack 2 version (branch: webpack)

3. SystemJS version (branch: systemjs)

Getting started
----------------------------------

#### Angular-CLI version

This seed version provides the following out of the box:

- Angular 2 application with [clarity-icons](https://www.npmjs.com/package/clarity-icons), [clarity-ui](https://www.npmjs.com/package/clarity-ui) and [clarity-angular](https://www.npmjs.com/package/clarity-angular) included
- Development and production builds
- Unit test setup with Jasmine and Karma
- End-to-end test setup with Protractor
- SASS processor
- TSLint
- And other goodies that come with [Angular-CLI](https://github.com/angular/angular-cli#generating-and-serving-an-angular2-project-via-a-development-server) (v1.0.0-beta.20-4)

#### Installation
*Prerequisite*: Please install Angular-CLI by following [these instructions](https://github.com/angular/angular-cli#installation).
*Note*: Even though it's optional, we recommend you to use [yarn](https://yarnpkg.com/) instead of `npm install` for installing the dependencies.

```bash
git clone https://github.com/vmware/clarity-seed.git
cd clarity-seed

# install the project's dependencies
yarn # or run "npm install"

# starts the application in dev mode and watches your files for livereload
ng serve
```

#### Using Angular-CLI
```bash
# generating a new component
ng g component my-new-component

# generating a new directive
ng g directive my-new-directive

# to learn more about Angular-CLI commands and their usages
ng help
```

For comprehensive documentation on Angular-CLI, please see their [github repository](https://github.com/angular/angular-cli).

#### Test and build scripts

```bash
# running unit tests
ng test

# running e2e tests
ng e2e

# dev build
ng build

# prod build
ng build --prod
```

## Documentation


For documentation on the Clarity Design System, including a list of components and example usage, see [our website](https://vmware.github.io/clarity).


#### Directory structure
```
.
├── README.md

├── karma.conf.js              <- configuration of the test runner
├── package.json               <- dependencies of the project
├── protractor.config.js       <- e2e tests configuration
├── src/                       <- source code of the application
│   ├── app/
│   │   └── component/
│   │       └── <component>.component.html
│   │       └── <component>.component.scss
│   │       └── <component>.component.spec.ts
│   │       └── <component>.component.ts
│   │   └── app.component.html
│   │   └── app.component.scss
│   │   └── app.component.ts
│   │   └── app.e2e-spec.js    <- sample e2e spec file
│   │   └── app.module.ts
│   │   └── app.routing.ts
│   │   └── main.ts            <- boostrap file for the angular app
│   └── index.html
├── angular-cli.json           <- configuration of the angular-cli
├── tsconfig.json              <- configuration of the typescript project
├── tslint.json                <- sample configuration file for tslint
└── yarn.lock
```


## Contributing

The Clarity project team welcomes contributions from the community. For more detailed information, see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

The clarity-seed project is licensed under the MIT license.

## Feedback

If you find a bug or want to request a new feature, please open a [GitHub issue](https://github.com/vmware/clarity-seed/issues).
