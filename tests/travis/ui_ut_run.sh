cd ./src/portal
npm install -g -q --no-progress angular-cli
npm install -g -q --no-progress karma
npm install -q --no-progress
npm run build_lib && npm run link_lib && cd ../..

cd ./src/portal && npm run lint && npm run lint:lib && npm run test && cd -