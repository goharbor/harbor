var path = require('path');
var webpack = require('webpack');

var BUILD_DIR = path.resolve(__dirname, 'build');
var APP_DIR = path.resolve(__dirname, 'src');

module.exports = {
  entry: APP_DIR + '/index.js',
  output: { path: BUILD_DIR, filename: 'work.js' },
  plugins: [
    new webpack.DefinePlugin({
      'process.env': {
        'NODE_ENV': JSON.stringify('production')
      }
    }),
    new webpack.optimize.UglifyJsPlugin({
      compressor: {
        warnings: false
      },
      comments: false
    })
  ],
  module: {
    loaders: [
      {
        test: /\.js$/,
        loader: 'babel-loader',
        exclude: /node_modules/,
      },
      {
        test: /\.css$/,
        loader: 'style-loader!css-loader?modules&camelCase&-url&localIdentName=[hash:base64:5]-[local]'
      },
      {
        test: /\.(png|woff|woff2|eot|ttf|svg)(\?[#a-z_]+)?$/,
        loader: 'url'
      }
    ]
  }
};

