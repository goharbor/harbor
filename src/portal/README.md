![Harbor UI](https://raw.githubusercontent.com/goharbor/website/master/docs/img/readme/harbor_logo.png)

Harbor UI
============
This is the project based on Clarity and Angular to build Harbor UI.



Start
============
1. npm install (should trigger 'npm postinstall')
2. npm run postinstall  (if not triggered, manually run this step)
3. create "proxy.config.json" file with below content under "portal" directory, and replace "hostname" with an available Harbor hostname
4. npm run start
5. open your browser on https://localhost:4200
```json
[
  {
    "context": [
      "/api",
      "/c",
      "/i18n",
      "/chartrepo",
      "/LICENSE",
      "/swagger.json",
      "/swagger2.json",
      "/devcenter-api-2.0",
      "/swagger-ui.bundle.js"
    ],
    "target": "https://hostname",
    "secure": false,
    "changeOrigin": true,
    "logLevel": "debug"
  }
]
```
