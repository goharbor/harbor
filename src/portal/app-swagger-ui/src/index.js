import SwaggerUI from 'swagger-ui'
import 'swagger-ui/dist/swagger-ui.css';

const helpInfo =
    ' If you want to enable basic authorization,' +
    ' please logout Harbor first or manually delete the cookies under the current domain.';

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
        });
        // remove loading style
       document.getElementById('swagger-ui-container').removeAttribute('class');

    })
    .catch((err) => {
        console.error(err);
    });
