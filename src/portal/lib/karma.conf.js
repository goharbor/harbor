// Karma configuration file, see link for more information
// https://karma-runner.github.io/1.0/config/configuration-file.html

const path = require('path');
module.exports = function (config) {
    config.set({
      basePath: '',
      frameworks: ['jasmine', '@angular-devkit/build-angular'],
      plugins: [
        require('karma-jasmine'),
        require('karma-chrome-launcher'),
        require('karma-mocha-reporter'),
        require('karma-coverage-istanbul-reporter'),
        require('@angular-devkit/build-angular/plugins/karma')
      ],
      client: {
        clearContext: false // leave Jasmine Spec Runner output visible in browser
      },
      coverageIstanbulReporter: {
        // reports can be any that are listed here: https://github.com/istanbuljs/istanbuljs/tree/aae256fb8b9a3d19414dcf069c592e88712c32c6/packages/istanbul-reports/lib
        reports: ['html', 'lcovonly', 'text-summary'],
   
        // base output directory. If you include %browser% in the path it will be replaced with the karma browser name
        dir: path.join(__dirname, 'coverage'),
   
        // Combines coverage information from multiple browsers into one report rather than outputting a report
        // for each browser.
        combineBrowserReports: true,
   
        // if using webpack and pre-loaders, work around webpack breaking the source path
        fixWebpackSourcePaths: true,
   
        // Omit files with no statements, no functions and no branches from the report
        skipFilesWithNoCoverage: true,
   
        // Most reporters accept additional config options. You can pass these through the `report-config` option
        'report-config': {
          // all options available at: https://github.com/istanbuljs/istanbuljs/blob/aae256fb8b9a3d19414dcf069c592e88712c32c6/packages/istanbul-reports/lib/html/index.js#L135-L137
          html: {
            // outputs the report in ./coverage/html
            subdir: 'html'
          }
        },
   
        // enforce percentage thresholds
        // anything under these percentages will cause karma to fail with an exit code of 1 if not running in watch mode
        thresholds: {
          emitWarning: false, // set to `true` to not fail the test command when thresholds are not met
          // thresholds for all files
          global: {
            statements: 37,
            branches: 20,
            functions: 28,
            lines: 36
          },
          // thresholds per file
          each: {
            statements: 0,
            lines: 0,
            branches: 0,
            functions: 0
          }
        }

      },
      reporters: ['progress', 'mocha','coverage-istanbul'],
      mochaReporter: {
        output: 'minimal'
      },
      reportSlowerThan: 100,
      port: 9876,
      colors: true,
      logLevel: config.LOG_INFO,
      autoWatch: true,
      singleRun: true,
      browsers: ['ChromeHeadlessNoSandbox'],
      browserDisconnectTolerance: 2,
      browserNoActivityTimeout: 50000,
      customLaunchers: {
        ChromeHeadlessNoSandbox: {
          base: 'ChromeHeadless',
          flags: ['--no-sandbox']
        }
      },
      restartOnFileChange: true
    });
  };