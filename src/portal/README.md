![Harbor UI](https://raw.githubusercontent.com/goharbor/website/master/docs/img/readme/harbor_logo.png)

Harbor UI
============
This is the project based on Clarity and Angular to build Harbor UI.



Start
============
1. npm install (should trigger 'npm postinstall')
2. npm run postinstall  (if not triggered, manually run this step)
3. copy "proxy.config.mjs.temp" file to "proxy.config.mjs"                 
   `cp proxy.config.mjs.temp proxy.config.mjs`
4. Modify "proxy.config.mjs" to specify a Harbor server. And you can specify the agent if you work behind a corporate proxy
5. npm run start
6. open your browser on https://localhost:4200

