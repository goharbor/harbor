import express from "express";
import { Express } from 'express';
import * as Controllers from '../controllers';


const mockApi: Express = express();

mockApi.get('/',  (req, res) => {
  res.send('Hello World!');
});

mockApi.get('/api/scanners', Controllers.getScanner);

mockApi.listen(3000,  ()  => {
  console.log('Api server listening on port 3000!');
});

export default mockApi;
