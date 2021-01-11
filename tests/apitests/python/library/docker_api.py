# -*- coding: utf-8 -*-

import base

try:
    import docker
except ImportError:
    import pip
    pip.main(['install', 'docker'])
    import docker
class DockerAPI(object):
    def __init__(self):
        self.DCLIENT = docker.APIClient(base_url='unix://var/run/docker.sock',version='auto',timeout=10)
        self.DCLIENT2 = docker.from_env()

    def docker_login(self, registry, username, password, expected_error_message = None):
        ret = ""
        err_message = ""
        if  username == "" or password == "":
            print("[Warnig]: No docker credential was provided.")
            return
        if expected_error_message == "":
            expected_error_message = None
        if registry == "docker":
            registry = None
        try:
            print("Docker login: {}:{}:{}".format(registry,username,password))
            ret = self.DCLIENT.login(registry = registry, username=username, password=password)
        except Exception as err:
            print( "Docker image pull catch exception:", str(err))
            err_message = str(err)
            if expected_error_message is None:
                raise Exception(r" Docker pull image {} failed, error is [{}]".format (image, str(err)))
        else:
            print("Docker image login did not catch exception and return message is:", ret)
            err_message = ret
        finally:
            if expected_error_message is not None:
                if str(err_message).lower().find(expected_error_message.lower()) < 0:
                    raise Exception(r" Failed to catch error [{}] when login image {}, return message: {}".format (expected_error_message, image, err_message))
                else:
                    print(r"Docker image login got expected error message:{}".format(expected_error_message))
            else:
                if str(err_message).lower().find("error".lower()) >= 0:
                    raise Exception(r" It's was not suppose to catch error when login image {}, return message is [{}]".format (image, err_message))

    def docker_image_pull(self, image, tag = None, expected_error_message = None):
        ret = ""
        err_message = ""
        if tag is not None:
            _tag = tag
        else:
            _tag = "latest"
        if expected_error_message is "":
            expected_error_message = None
        try:
            ret = self.DCLIENT.pull(r'{}:{}'.format(image, _tag))
        except Exception as err:
            print( "Docker image pull catch exception:", str(err))
            err_message = str(err)
            if expected_error_message is None:
                raise Exception(r" Docker pull image {} failed, error is [{}]".format (image, str(err)))
        else:
            print("Docker image pull did not catch exception and return message is:", ret)
            err_message = ret
        finally:
            if expected_error_message is not None:
                if str(err_message).lower().find(expected_error_message.lower()) < 0:
                    raise Exception(r" Failed to catch error [{}] when pull image {}, return message: {}".format (expected_error_message, image, err_message))
                else:
                    print(r"Docker image pull got expected error message:{}".format(expected_error_message))
            else:
                if str(err_message).lower().find("error".lower()) >= 0:
                    raise Exception(r" It's was not suppose to catch error when pull image {}, return message is [{}]".format (image, err_message))

    def docker_image_tag(self, image, harbor_registry, tag = None):
        _tag = base._random_name("tag")
        if tag is not None:
            _tag = tag
        ret = ""
        try:
            ret = self.DCLIENT.tag(image, harbor_registry, _tag, force=True)
            print("Docker image tag commond return:", ret)
            return harbor_registry, _tag
        except docker.errors.APIError as err:
            raise Exception(r" Docker tag image {} failed, error is [{}]".format (image, str(err)))

    def docker_image_push(self, harbor_registry, tag, expected_error_message = None):
        ret = ""
        err_message = ""
        if expected_error_message is "":
            expected_error_message = None
        try:
            ret = self.DCLIENT.push(harbor_registry, tag)
        except Exception as err:
            print( "Docker image push catch exception:", str(err))
            err_message = str(err)
            if expected_error_message is None:
                raise Exception(r" Docker push image {} failed, error is [{}]".format (image, str(err)))
        else:
            print("Docker image push did not catch exception and return message is:", ret)
            err_message = ret
        finally:
            if expected_error_message is not None:
                if str(err_message).lower().find(expected_error_message.lower()) < 0:
                    raise Exception(r" Failed to catch error [{}] when push image {}, return message: {}".format (expected_error_message, harbor_registry, err_message))
                else:
                    print(r"Docker image push got expected error message:{}".format(expected_error_message))
            else:
                if str(err_message).lower().find("error".lower()) >= 0:
                    raise Exception(r" It's was not suppose to catch error when push image {}, return message is [{}]".format (harbor_registry, err_message))

    def docker_image_build(self, harbor_registry, tags=None, size=1, expected_error_message = None):
        ret = ""
        err_message = ""
        try:
            baseimage='busybox:latest'
            if not self.DCLIENT.images(name=baseimage):
                print( "Docker pull is triggered when building {}".format(harbor_registry))
                self.DCLIENT.pull(baseimage)
            c=self.DCLIENT.create_container(image='busybox:latest',
                command='dd if=/dev/urandom of=test bs=1M count={}'.format(size))
            self.DCLIENT.start(c)
            self.DCLIENT.wait(c)
            if not tags:
                tags=['latest']
            firstrepo="{}:{}".format(harbor_registry, tags[0])
            #self.DCLIENT.commit(c, firstrepo)
            self.DCLIENT2.containers.get(c).commit(harbor_registry, tags[0])
            for tag in tags[1:]:
                repo="{}:{}".format(harbor_registry, tag)
                self.DCLIENT.tag(firstrepo, repo)
            for tag in tags:
                repo="{}:{}".format(harbor_registry, tag)
                ret = self.DCLIENT.push(repo)
                print("docker_image_push ret:", ret)
                print("build image {} with size {}".format(repo, size))
                self.DCLIENT.remove_image(repo)
            self.DCLIENT.remove_container(c)
            #self.DCLIENT.pull(repo)
            #image = self.DCLIENT2.images.get(repo)
        except Exception as err:
            print( "Docker image build catch exception:", str(err))
            err_message = str(err)
            if expected_error_message is None:
                raise Exception(r" Docker push image {} failed, error is [{}]".format (harbor_registry, str(err)))
        else:
            print("Docker image build did not catch exception and return message is:", ret)
            err_message = ret
        finally:
            if expected_error_message is not None:
                if str(err_message).lower().find(expected_error_message.lower()) < 0:
                    raise Exception(r" Failed to catch error [{}] when build image {}, return message: {}".format (expected_error_message, harbor_registry, err_message))
                else:
                    print(r"Docker image build got expected error message:{}".format(expected_error_message))
            else:
                if str(err_message).lower().find("error".lower()) >= 0:
                    raise Exception(r" It's was not suppose to catch error when build image {}, return message is [{}]".format (harbor_registry, err_message))
