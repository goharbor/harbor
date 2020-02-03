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
      mockApi.get('/api/scanners', Controllers.getScanner);
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

   4.(Optional)Edit file proxy.config.json, and add below code on the top:
   ```json
     "/api/scanner": {
     "target": "http://localhost:3000",
     "secure": false,
     "changeOrigin": true,
     "logLevel": "debug"
    }
   ```

   5.Now, you can get mocked scanners when you send "get request" to "/api/scanners"
