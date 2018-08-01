import nodeResolve from 'rollup-plugin-node-resolve';
import commonjs    from 'rollup-plugin-commonjs';
import uglify      from 'rollup-plugin-uglify';

export default {
  entry: 'src/main-aot.js',
  dest: 'dist/build.min.js', // output a single application compressed file.
  sourceMap: false,
  format: 'iife',
  onwarn: function(warning) {
    // Skip certain warnings

    // should intercept ... but doesn't in some rollup versions
    if ( warning.code === 'THIS_IS_UNDEFINED' ) { return; }
    // intercepts in some rollup versions
    if ( typeof warning === 'string' && warning.indexOf("The 'this' keyword is equivalent to 'undefined'") > -1 ) { return; }

    // console.warn everything else
    console.warn( warning.message );
  },
  plugins: [
      nodeResolve({jsnext: true, module: true, browser: true}),
      commonjs({
        namedExports: {
          'node_modules/ngx-markdown/dist/lib/index.js': ['MarkdownModule']
        },
        include: ['node_modules/**',
        'node_modules/ngx-markdown/**'],
      }),
      uglify()
  ]
}
