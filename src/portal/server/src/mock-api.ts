import express from 'express';
import { Express } from 'express';
import * as Controllers from '../controllers';

const mockApi: Express = express();

const CURRENT_BASE_HREF: string = '/api/v2.0';

mockApi.get('/', (req, res) => {
    res.send('Hello World!');
});

mockApi.get(CURRENT_BASE_HREF + '/scanners', Controllers.getScanner);

mockApi.listen(3000, () => {
    // eslint-disable-next-line no-console
    console.log('Api server listening on port 3000!');
});

export default mockApi;
