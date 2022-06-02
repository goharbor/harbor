Steps to deploy Harbor UI in a nginx container, it can be used for testing

1. Go to docker-build dir

`cd   ./docker-build`
   
2. Copy `nginx.conf.example` to `nginx.conf`, and modify nginx.conf file to specify an available back-end server

`cp nginx.conf.example nginx.conf`

`
location ~ ^/(api|c|chartrepo)/ {
   proxy_pass ${an available back-end server addr};
}
`

3. Build harbor-ui image

`docker build -f ./Dockerfile -t harbor-ui:test ./../../..`
   
4. Run  harbor-ui image

`docker run -p 8080:8080 harbor-ui:test`

5. Open your browser on http://localhost:8080   
