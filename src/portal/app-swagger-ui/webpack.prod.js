const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const { CleanWebpackPlugin } = require('clean-webpack-plugin');
const outputPath = path.resolve(__dirname, 'dist');

module.exports = {
    mode: 'production',
    entry: {
        app: require.resolve('./src/index'),
    },
    resolve: {
        // swagger-ui 5.28+ nests React 19 which removed createRoot from the
        // main react-dom entry. swagger-ui-bundle.js is the standalone UMD
        // build that ships with React 18 already bundled, so redirect the
        // bare "swagger-ui" import to it to avoid the React 19 breakage.
        alias: {
            'swagger-ui$': path.resolve(__dirname, 'node_modules/swagger-ui/dist/swagger-ui-bundle.js'),
        },
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
