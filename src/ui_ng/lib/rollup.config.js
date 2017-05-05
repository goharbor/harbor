export default {
    entry: 'dist/index.js',
    dest: 'dist/bundles/harborui.umd.js',
    sourceMap: false,
    format: 'umd',
    moduleName: 'harborui',
    external: [
        '@angular/animations',
        '@angular/core',
        '@angular/common',
        '@angular/forms',
        '@angular/platform-browser',
        '@angular/http',
        'clarity-angular',
        '@ngx-translate/core',
        '@ngx-translate/http-loader',
        'rxjs',
        'rxjs/Subject',
        'rxjs/Observable',
        'rxjs/add/observable/of',
        'rxjs/add/operator/toPromise',
        'rxjs/add/operator/debounceTime',
        'rxjs/add/operator/distinctUntilChanged'
    ],
    globals: {
        '@angular/core': 'ng.core',
        '@angular/animations': 'ng.animations',
        '@angular/common': 'ng.common',
        '@angular/forms': 'ng.forms',
        '@angular/http': 'ng.http',
        '@angular/platform-browser': 'ng.platformBrowser',
        'rxjs': 'rxjs',
        'rxjs/Subject': 'rxjs.Subject',
        'rxjs/Observable': 'Rx',
        'rxjs/ReplaySubject': 'Rx',
        'rxjs/add/operator/map': 'Rx.Observable.prototype',
        'rxjs/add/operator/mergeMap': 'Rx.Observable.prototype',
        'rxjs/add/operator/catch': 'Rx.Observable.prototype',
        'rxjs/add/operator/toPromise': 'Rx.Observable.prototype',
        'rxjs/add/observable/of': 'Rx.Observable',
        'rxjs/add/observable/throw': 'Rx.Observable'
    },
    onwarn: function(warning) {
        // Skip certain warnings

        // should intercept ... but doesn't in some rollup versions
        if (warning.code === 'THIS_IS_UNDEFINED') { return; }
        // intercepts in some rollup versions
        if (typeof warning === 'string' && warning.indexOf("The 'this' keyword is equivalent to 'undefined'") > -1) { return; }

        // console.warn everything else
        console.warn(warning.message);
    }
}