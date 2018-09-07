// Karma configuration file, see link for more information
// https://karma-runner.github.io/0.13/config/configuration-file.html

module.exports = function (config) {
    config.set({
        basePath: '/',
        frameworks: ['jasmine', '@angular-devkit/build-angular'],
        plugins: [
            require('karma-jasmine'),
            require('karma-chrome-launcher'),
            require('karma-mocha-reporter'),
            require('karma-remap-istanbul'),
            require('@angular-devkit/build-angular/plugins/karma')
        ],
        files: [
            {pattern: './src/test.ts', watched: false}
        ],
        preprocessors: {
            
        },
        mime: {
            'text/x-typescript': ['ts', 'tsx']
        },
        remapIstanbulReporter: {
            dir: require('path').join(__dirname, 'coverage'), reports: {
                html: 'coverage',
                lcovonly: './coverage/coverage.lcov'
            }
        },
        
        reporters: config.angularCli && config.angularCli.codeCoverage
            ? ['mocha', 'karma-remap-istanbul']
            : ['mocha'],
        port: 9876,
        colors: true,
        logLevel: config.LOG_INFO,
        autoWatch: true,
        browsers: ['ChromeHeadlessNoSandbox'],
        customLaunchers: {
            ChromeHeadlessNoSandbox: {
              base: 'ChromeHeadless',
              flags: ['--no-sandbox']
            }
          },
        singleRun: true
    });
};