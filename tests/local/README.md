## Developer testing environment

The local testing environment supports running the "Unit" tests, and the DB API tests.


#### Preparation
Build the docker image described by the Dockerfile in this directory, and tag it `harbor-ci`:

```shell script
$ docker build -t harbor-ci .
```

This docker image will need to be rebuilt every time there is a change in any of the files under 
`harbor/tests/local/`

#### Running the tests

To run the "Unit" tests, execute the following:
```shell script
$ docker run --user=travis -v /Users/pivotal/workspace/harbor:/h:ro -e DEBUG=1 --privileged -it harbor-ci /home/travis/ut_test.sh
```
To run the DB API tests, execute the following:
```shell script
$ docker run --user=travis -v /Users/pivotal/workspace/harbor:/h:ro -e DEBUG=1 --privileged -it harbor-ci /home/travis/db_api_test.sh
```