/**
 * generate timestamp for each production build
 */
const fs = require('fs');
const data = fs.readFileSync('src/environments/environment.prod.ts', 'utf8').split('\n');
const timestamp = new Date().getTime();

let buildTimestampIndex = 0;
data.forEach((item,index) => {
  if(item.indexOf('buildTimestamp') !== -1) {
    buildTimestampIndex = index;
  }
});
// modify buildTimestamp value in src/environments/environment.prod.ts file
if (buildTimestampIndex > 0) {
  data[buildTimestampIndex] = `  buildTimestamp: ${timestamp},`;
  fs.writeFileSync('src/environments/environment.prod.ts', data.join('\n'), 'utf8');
}

// modify below lines in src/index.html to add buildTimestamp query string, in case of css cache in different builds
//     <link rel="preload" as="style" href="./light-theme.css">
//     <link rel="preload" as="style" href="./dark-theme.css">
const indexHtmlData = fs.readFileSync('src/index.html', 'utf8').split('\n');
let lightIndex = 0;
let darkIndex =0;
indexHtmlData.forEach((item,index) => {
  if(item.indexOf('light-theme.css') !== -1) {
    lightIndex = index;
  }
  if(item.indexOf('dark-theme.css') !== -1) {
    darkIndex = index;
  }
});

if (lightIndex > 0 && darkIndex > 0) {
  indexHtmlData[lightIndex] = `    <link rel="preload" as="style" href="./light-theme.css?buildTimestamp=${timestamp}">`;
  indexHtmlData[darkIndex] = `    <link rel="preload" as="style" href="./dark-theme.css?buildTimestamp=${timestamp}">`;
  fs.writeFileSync('src/index.html', indexHtmlData.join('\n'), 'utf8');
}