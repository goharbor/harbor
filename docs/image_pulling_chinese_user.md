### A faster way to pull images for Chinese Harbor users
By default, Harbor not only build images according to Dockerfile but also pull images from Docker Hub. For the reason we all know, it is difficult for Chinese Harbor users to pull images from the Docker Hub. We put images on daocloud.io platform, we'll put images on other platforms later. If you have difficulty to pull images from Docker Hub, or you think it wastes too much time to build images. We recommend you to use the following way to accelerate the pulling procedure(make sure you're in the harbor diectory):
```
$ cd contrib
$ cp docker-compose.yml.daocloud ../Deploy
$ cd ../Deploy
$ mv docker-compose.yml docker-compose.yml.bak
$ mv docker-compose.yml.daocloud docker-compose.yml
$ docker-compose up -d 
```
Then you'll see docker pulling imges faster than before.
