import SwaggerUI from 'swagger-ui'
import 'swagger-ui/dist/swagger-ui.css';

const helpInfo =
    ' If you want to enable basic authorization,' +
    ' please logout Harbor first or manually delete the cookies under the current domain.';
const SAFE_METHODS = ['GET', 'HEAD', 'OPTIONS', 'TRACE'];

// get swagger.json from portal container then render swagger ui
// before rendering, the ui shows a loading style
fetch('/swagger.json').then(value => value.json()).then(res => {
    res['host'] = window.location.host;
    const protocal = window.location.protocol;
    res['schemes'] = [protocal.replace(':', '')];
    res.info.description = res.info.description + helpInfo;
        // start to render
        SwaggerUI({
            spec: res,
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
