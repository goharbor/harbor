![Harbor UI](https://raw.githubusercontent.com/goharbor/website/master/docs/img/readme/harbor_logo.png)

Harbor UI
============
This is the project based on Clarity and Angular to build Harbor UI.



Start
============
1. Use the specified Node version:
Run the following command to use the Node version specified in the .nvmrc file:
```bash
nvm install   # Install the Node version specified in .nvmrc (if not already installed)
nvm use       # Switch to the specified Node version
```
This step helps avoid compatibility issues, especially with dependencies.
2. npm install (should trigger 'npm postinstall')
3. npm run postinstall  (if not triggered, manually run this step)
4. copy "proxy.config.mjs.temp" file to "proxy.config.mjs"
   `cp proxy.config.mjs.temp proxy.config.mjs`
5. Modify "proxy.config.mjs" to specify a Harbor server. And you can specify the agent if you work behind a corporate proxy
6. npm run start
7. open your browser on https://localhost:4200

