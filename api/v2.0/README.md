# References
 
This file lists all the files that are referring the swagger yaml file.

- Makefile
  - `java -jar swagger-codegen-cli.jar generate -i api/harbor/swagger.yaml -l python -o harborclient`
- README
  - `Harbor RESTful API` in `API` section
- docs/configure_swagger.md
  - `https://raw.githubusercontent.com/goharbor/harbor/master/api/harbor/swagger.yaml`
- make/photon/portal/Dockerfile
  - `COPY ./api/harbor/swagger.yaml /build_dir`
- tests/swaggerchecker.sh
  - `HARBOR_SWAGGER_FILE="https://raw.githubusercontent.com/$TRAVIS_REPO_SLUG/$TRAVIS_COMMIT/api/harbor/swagger.yaml"`
  - else `HARBOR_SWAGGER_FILE="https://raw.githubusercontent.com/$TRAVIS_PULL_REQUEST_SLUG/$TRAVIS_PULL_REQUEST_SHA/api/harbor/swagger.yaml"`


**Notes:** Base path is the code repository root dir.