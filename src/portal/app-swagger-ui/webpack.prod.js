const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const { CleanWebpackPlugin } = require('clean-webpack-plugin');
const outputPath = path.resolve(__dirname, 'dist');

module.exports = {
    mode: 'production',
    entry: {
        app: require.resolve('./src/index'),
    },
    module: {
        rules: [
            {
                test: /\.css$/,
                use: [
                    { loader: 'style-loader' },
                    { loader: 'css-loader' },
                ]
            }
        ]
    },
    plugins: [
        new CleanWebpackPlugin(),
        new HtmlWebpackPlugin({
            template: 'index.html'
        })
    ],
    output: {
        filename: 'swagger-ui.bundle.js',
        path: outputPath,
    }
};
