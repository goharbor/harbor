/**
 before execute:
 1.npm install js-yaml --save-dev
 2.npm install ng-swagger-gen --save-dev
 */
//configuration
let inputFile = '../../api/v2.0/swagger.yaml';
const outputDir = 'ng-swagger-gen';

//convert swagger.yaml to swagger.json
const yaml = require('js-yaml');
const fs = require('fs');
if (fs.existsSync('swagger2.yaml')) {
   inputFile = 'swagger2.yaml';
}
if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir);
}
const swaggerObj = yaml.load(fs.readFileSync(inputFile, {encoding: 'utf-8'}));
fs.writeFileSync(outputDir + '/swagger.json', JSON.stringify(swaggerObj, null, 2));
