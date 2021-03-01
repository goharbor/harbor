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
{
  "/api/*": {
    "target": "https://hostname",
    "secure": false,
    "changeOrigin": true,
    "logLevel": "debug"
  },
  "/service/*": {
    "target": "https://hostname",
    "secure": false,
    "logLevel": "debug"
  },
  "/c/login": {
    "target": "https://hostname",
    "secure": false,
    "logLevel": "debug"
  },
  "/c/oidc/login": {
    "target": "https://hostname",
    "secure": false,
    "logLevel": "debug"
  },
  "/sign_in": {
    "target": "https://hostname",
    "secure": false,
    "logLevel": "debug"
  },
  "/c/log_out": {
    "target": "https://hostname",
    "secure": false,
    "logLevel": "debug"
  },
  "/sendEmail": {
    "target": "https://hostname",
    "secure": false,
    "logLevel": "debug"
  },
  "/language": {
    "target": "https://hostname",
    "secure": false,
    "logLevel": "debug"
  },
  "/reset": {
    "target": "https://hostname",
    "secure": false,
    "logLevel": "debug"
  },
  "/c/userExists": {
    "target": "https://hostname",
    "secure": false,
    "logLevel": "debug"
  },
  "/reset_password": {
    "target": "https://hostname",
    "secure": false,
    "logLevel": "debug"
  },
  "/i18n/lang/*.json": {
    "target": "https://hostname",
    "secure": false,
    "logLevel": "debug"
  },
  "/swagger.json": {
    "target": "https://hostname",
    "secure": false,
    "logLevel": "debug"
  },
  "/swagger2.json": {
    "target": "https://hostname",
    "secure": false,
    "logLevel": "debug"
  },
  "/chartrepo/*": {
    "target": "https://hostname",
    "secure": false,
    "logLevel": "debug"
  },
  "/LICENSE": {
    "target": "https://hostname",
    "secure": false,
    "logLevel": "debug"
  }
}
```
