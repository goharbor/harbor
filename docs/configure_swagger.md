#Configure Swagger for Harbor API#
As for manipulating and documenting Harbor project APIs, we created it by using Swagger.
## Two forms of demonstration
There are two forms of demonstration of Swagger UI for Harbor API.
### Listed descriptions only
It means you only need to locate the YAML file of Harbor APIs in Swagger UI, 
then a listed descriptions will be shown. It doesn't affect the deployed Harbor project node actually.
### Full functions provided by Swagger
This form of use is a bit of difficult, because it must bind to an available Harbor project node in order to solve the problem of CORS. Well, it will provide full functions accompany with Harbor project node, you must be careful of your operations to avoid damaging backend data.
## Detail instructions for each
First, you should checkout the Harbor project from github.
```sh
   git clone git@github.com:vmware/harbor.git
```
### Listed descriptions only
* Download and untar the Swagger UI release package.
```sh
   wget https://github.com/swagger-api/swagger-ui/archive/v2.1.4.tar.gz \
     -O swagger.tar.gz
   tar -zxf swagger.tar.gz swagger-ui-2.1.4/dist
```
* Open the _index.html_ file with a browser, input the file path URL of the _swagger.yaml_ file onto Swagger UI page, then click _Explore_ button.
```
 file:///root/harbor/docs/swagger.yaml
```
### Full functions provided by Swagger UI
* Change the directory to _docs_
```sh
  cd docs
```
* Open script _prepare-swagger.sh_ located at _docs_ directory.
```sh
  vi prepare-swagger.sh
```
* Change SERVER_IP value with your deployed Harbor node.
```sh
  SERVER_ID=10.117.170.69
```
* Execute this shell script. It helps you to download a Swagger UI release package, untar it into the static files directory of the Harbor project.
```sh
   ./prepare-swagger.sh
```
* Change the directory to _Deploy_
```sh
  cd ../Deploy
```
* Open _docker-compose.yml_ file.
```sh
  vi docker-compose.yml
```
* Add two lines in the _docker-compose.yml_ file at _ui_ _volumes_ configure segment.
```docker
## omit other lines ##
ui:
  ## omit other lines ##
  volumes:
    - ./config/ui/app.conf:/etc/ui/app.conf
    - ./config/ui/private_key.pem:/etc/ui/private_key.pem
    ## add two lines as below ##
    - ../static/vendors/swagger-ui-2.1.4/dist:/go/bin/static/vendors/swagger
    - ../static/resources/yaml/swagger.yaml:/go/bin/static/resources/yaml/swagger.yaml
  ## omit other lines ##
```
* Rebuild Harbor project
```docker
    docker-compose build
```
* Clean up left before .
```docker
   docker-compose rm
```
* Start new updated Harbor project
```docker
   docker-compose up
```
* Because session ID is required for using Harbor APIs. **You should log in first from UI by using a browser.**
* Open another tab in the current browser to keep one session.
* Input the URL deployed by Harbor project. **You should change the IP address with your deployed Harbor project node.**
```
  http://10.117.170.69/static/vendors/swagger/index.html
```
* Then you should see a Swagger UI deployed by Harbor project, loaded a _swagger.yaml_ file in the same domain, it works actually, **be careful of your operations to avoid damaging backend data**.
