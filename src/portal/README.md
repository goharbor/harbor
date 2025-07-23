![Harbor UI](https://raw.githubusercontent.com/goharbor/website/master/docs/img/readme/harbor_logo.png)

# Harbor UI

This project is the web interface for [Harbor](https://goharbor.io), built using [Clarity Design System](https://clarity.design/) and Angular.

## Getting Started

### 1. Use the correct Node version

To ensure compatibility with dependencies, use the Node version defined in `.nvmrc`.

```
nvm install   # Install the Node version from .nvmrc (if not already installed)
nvm use       # Switch to the specified Node version
```

### 2. Install dependencies

```
npm install
```

> Note: `npm install` should automatically trigger the `postinstall` script.
If `postinstall` scripts were not triggered, then run manually:  `npm run postinstall`


### 3. Copy the template proxy file

```
cp proxy.config.mjs.temp proxy.config.mjs
```

### 4. Configure the proxy

Edit `proxy.config.mjs` to specify the Harbor server.
You can specify the agent if you work behind a corporate proxy.

### 5. Start the development server

```
npm run start
```

### 6. Open the application

Open your browser at https://localhost:4200



