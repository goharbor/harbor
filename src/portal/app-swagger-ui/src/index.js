import SwaggerUI from 'swagger-ui'
import 'swagger-ui/dist/swagger-ui.css';

const helpInfo =
    ' If you want to enable basic authorization,' +
    ' please logout Harbor first or manually delete the cookies under the current domain.';
const SAFE_METHODS = ['GET', 'HEAD', 'OPTIONS', 'TRACE'];

// get swagger.json and swagger2.json from portal container then render swagger ui
// before rendering, the ui shows a loading style
Promise.all([
    fetch('/swagger.json').then(value => value.json()),
    fetch('/swagger2.json').then(value => value.json())
])
    .then(value => {
        // merger swagger.json and swagger2.json
        const json = {};
        mergeDeep(json, value[0], value[1]);
        json['host'] = window.location.host;
        const protocal = window.location.protocol;
        json['schemes'] = [protocal.replace(':', '')];
        json.info.description = json.info.description + helpInfo;
        // start to render
        SwaggerUI({
            spec: json,
            dom_id: '#swagger-ui-container',
            deepLinking: true,
            presets: [SwaggerUI.presets.apis],
            requestInterceptor: request => {
                // Get the csrf token from localstorage
                const token = localStorage.getItem('__csrf');
                const headers = request.headers || {};
                if (token) {
                    if (
                        request.method &&
                        SAFE_METHODS.indexOf(
                            request.method.toUpperCase()
                        ) === -1
                    ) {
                        headers['X-Harbor-CSRF-Token'] = token;
                    }
                }
                return request;
            },
            responseInterceptor: response => {
                const headers = response.headers || {};
                const responseToken =
                    headers['X-Harbor-CSRF-Token'];
                if (responseToken) {
                    // Set the csrf token to localstorage
                    localStorage.setItem('__csrf', responseToken);
                }
                return response;
            },
        });
        // remove loading style
       document.getElementById('swagger-ui-container').removeAttribute('class');

    })
    .catch((err) => {
        console.error(err);
    });







function mergeDeep(target, ...sources) {
    if (!sources.length) {
        return target;
    }
    const source = sources.shift();

    if (isObject(target) && isObject(source)) {
        for (const key in source) {
            if (isObject(source[key])) {
                if (!target[key]) {
                    Object.assign(target, { [key]: {} });
                }
                mergeDeep(target[key], source[key]);
            } else {
                Object.assign(target, { [key]: source[key] });
            }
        }
    }
    return mergeDeep(target, ...sources);
}

function isObject(item) {
    return item && typeof item === 'object' && !Array.isArray(item);
}
