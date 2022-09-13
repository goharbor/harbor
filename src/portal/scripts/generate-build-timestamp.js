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
