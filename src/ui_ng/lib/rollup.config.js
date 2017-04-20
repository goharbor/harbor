export default {
    entry: 'dist/index.js',
    dest: 'dist/bundles/harborui.umd.js',
    sourceMap: false,
    format: 'umd',
    moduleName: 'harborui',
    globals: {
        '@angular/core': 'ng.core',
        'rxjs/Observable': 'Rx',
        'rxjs/ReplaySubject': 'Rx',
        'rxjs/add/operator/map': 'Rx.Observable.prototype',
        'rxjs/add/operator/mergeMap': 'Rx.Observable.prototype',
        'rxjs/add/operator/catch': 'Rx.Observable.prototype',
        'rxjs/add/operator/toPromise': 'Rx.Observable.prototype',
        'rxjs/add/observable/of': 'Rx.Observable',
        'rxjs/add/observable/throw': 'Rx.Observable'
    }
}