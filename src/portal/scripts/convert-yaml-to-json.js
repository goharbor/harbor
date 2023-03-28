/**
 before execute:
 1.npm install js-yaml --save-dev
 2.npm install ng-swagger-gen --save-dev
 */
//configuration. For dev build, the input path is '../../api/v2.0/swagger.yaml'
let inputFile = '../../api/v2.0/swagger.yaml';
const outputDir = 'ng-swagger-gen';

//convert swagger.yaml to swagger.json
const yaml = require('js-yaml');
const fs = require('fs');
//when building portal container(production build), the input path is './swagger.yaml'. Refer to portal Dockerfile
if (fs.existsSync('swagger.yaml')) {
   inputFile = 'swagger.yaml';
}
if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir);
}
const swaggerObj = yaml.load(fs.readFileSync(inputFile, {encoding: 'utf-8'}));
// host is not needed as UI uses relative path for back-end APIs
if (swaggerObj.host) {
    delete swaggerObj.host;
}
// enhancement for property 'additionalProperties'
traverseObject(swaggerObj);

fs.writeFileSync(outputDir + '/swagger.json', JSON.stringify(swaggerObj, null, 2));


function traverseObject(obj) {
  if (obj) {
    if (Array.isArray(obj)) {
      for (let i = 0; i < obj.length; i++) {
        traverseObject(obj[i])
      }
    }
    if (typeof obj === 'object') {
      for (let name in obj) {
        if (obj.hasOwnProperty(name)) {
          if (name === 'additionalProperties'
            && obj[name].type === 'object'
            && obj[name].additionalProperties === true) {
            obj[name] = true;
          } else {
            traverseObject(obj[name])
          }
        }
      }
    }
  }
}

