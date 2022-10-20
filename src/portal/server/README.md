Mocked Api-server based on  "Node.js" + "Express" + "Typescript"

How to use?

suppose you want to mock a api server for below code:
```typescript
     getScannersByName(name: string): Observable<Scanner[]> {
        name = encodeURIComponent(name);
        return this.http.get(`/api/scanners?ex_name=${name}`)
                .pipe(catchError(error => observableThrowError(error)))
                .pipe(map(response => response as Scanner[]));
    }
```
   
   1. Edit mock-api.ts, and add below code: 
   ```typescript
      mockApi.get('/api/v2.0/scanners', Controllers.getScanner);
   ```
   2. Add your method implementation into folder "controllers" and export it:
   ```typescript
       export function getScanner(req: Request, res: Response) {
         const scanners: Scanner[] =  [new Scanner(), new Scanner()];
         res.json(scanners);
       }
   ```
   3.cd portal and run 
   ```shell script
     npm run mock-api-server
   ```

   4.Edit the file proxy.config.json, and add a proxy object at the top as below:
   ```json
[
  // redirect requests to the mocked api server
  {
    "context": [
      "/api/v2.0/scanners"
    ],
    "target": "http://localhost:3000",
    "secure": false,
    "changeOrigin": true,
    "logLevel": "debug"
  },
  
  // redirect requests to the back-end server
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

   5. Run `npm run start`, then all the mocked APIs will be redirected to the mocked server
