# -*- coding: utf-8 -*-

import base
import subprocess
import json
from testutils import DOCKER_USER, DOCKER_PWD

try:
    import docker
except ImportError:
    import pip
    pip.main(['install', 'docker'])
    import docker

def docker_info_display():
    command = ["docker", "info", "-f", "'{{.OSType}}/{{.Architecture}}'"]
    print("Docker Info: ", command)
    ret = base.run_command(command)
    print("Command return: ", ret)

def docker_login_cmd(harbor_host, username, password, cfg_file = "./tests/apitests/python/update_docker_cfg.sh",  enable_manifest = True):
    if  username == "" or password == "":
        print("[Warnig]: No docker credential was provided.")
        return
    command = ["sudo", "docker", "login", harbor_host, "-u", username, "-p", password]
    print( "Docker Login Command: ", command)
    base.run_command(command)
    if enable_manifest == True:
        try:
            ret = subprocess.check_output([cfg_file], shell=False)
            print("docker login cmd ret:", ret)
        except subprocess.CalledProcessError as exc:
            raise Exception("Failed to update docker config, error is {} {}.".format(exc.returncode, exc.output))

def docker_manifest_create(index, manifests):
    command = ["sudo", "docker","manifest","create", "--amend", index]
    command.extend(manifests)
    print( "Docker Manifest Command: ", command)
    base.run_command(command)

def docker_manifest_push(index):
    command = ["sudo", "docker","manifest","push",index]
    print( "Docker Manifest Command: ", command)
    ret = base.run_command(command)
    index_sha256=""
    manifest_list=[]
    for line in ret.split("\n"):
        if line[:7] == "sha256:":
            index_sha256 = line
        if line.find('Pushed ref') == 0:
            manifest_list.append(line[-71:])
    return index_sha256, manifest_list

def docker_manifest_push_to_harbor(index, manifests, harbor_server, username, password, cfg_file = "./tests/apitests/python/update_docker_cfg.sh"):
    docker_login_cmd(harbor_server, username, password, cfg_file=cfg_file)
    docker_manifest_create(index, manifests)
    return docker_manifest_push(index)

def list_repositories(harbor_host, username, password, n = None, last = None):
    if n is not None and last is not None:
        command = ["curl", "-s", "-u", username+":"+password, "https://"+harbor_host+"/v2/_catalog"+"?n=%d"%n+"&last="+last, "--insecure"]
    elif n is not None:
            command = ["curl", "-s", "-u", username+":"+password, "https://"+harbor_host+"/v2/_catalog"+"?n=%d"%n, "--insecure"]
    else:
        command = ["curl", "-s", "-u", username+":"+password, "https://"+harbor_host+"/v2/_catalog", "--insecure"]
    print( "List Repositories Command: ", command)
    ret = base.run_command(command)
    repos = json.loads(ret).get("repositories","")
    return repos

def list_image_tags(harbor_host, repository, username, password, n = None, last = None):
    if n is not None and last is not None:
        command = ["curl", "-s", "-u", username+":"+password, "https://"+harbor_host+"/v2/"+repository+"/tags/list"+"?n=%d"%n+"&last="+last, "--insecure"]
    elif n is not None:
        command = ["curl", "-s", "-u", username+":"+password, "https://"+harbor_host+"/v2/"+repository+"/tags/list"+"?n=%d"%n, "--insecure"]
    else:
        command = ["curl", "-s", "-u", username+":"+password, "https://"+harbor_host+"/v2/"+repository+"/tags/list", "--insecure"]
    print( "List Image Tags Command: ", command)
    ret = base.run_command(command)
    tags = json.loads(ret).get("tags","")
    return tags

class DockerAPI(object):
    def __init__(self):
        self.DCLIENT = docker.APIClient(base_url='unix://var/run/docker.sock',version='auto',timeout=30)
        self.DCLIENT2 = docker.from_env()

    def docker_login(self, registry, username, password, expected_error_message = None):
        if  username == "" or password == "":
            print("[Warnig]: No docker credential was provided.")
            return
        if expected_error_message == "":
            expected_error_message = None
        if registry == "docker":
            registry = None
        ret = ""
        try:
            print("Docker login: {}:{}:{}".format(registry,username,password))
            ret = self.DCLIENT.login(registry = registry, username=username, password=password)
            print("Docker image login commond return:", ret)
            return ret
        except docker.errors.APIError as err:
            if expected_error_message is not None:
                print( "docker login error:", str(err))
                if str(err).lower().find(expected_error_message.lower()) < 0:
                    raise Exception(r"Docker login: Return message {} is not as expected {}".format(str(err), expected_error_message))
            else:
                raise Exception(r" Docker login failed, error is [{}]".format (str(err)))

    def docker_image_pull(self, image, tag = None, expected_error_message = None):
        if tag is not None:
            _tag = tag
        else:
            _tag = "latest"
        if expected_error_message is "":
            expected_error_message = None
        ret = ""
        try:
            ret = self.DCLIENT.pull(r'{}:{}'.format(image, _tag))
            print("Docker image pull commond return:", ret)
            return ret
        except Exception as err:
            if expected_error_message is not None:
                print( "docker image pull error:", str(err))
                if str(err).lower().find(expected_error_message.lower()) < 0:
                    raise Exception(r"Pull image: Return message {} is not as expected {}".format(str(err), expected_error_message))
            else:
                raise Exception(r" Docker pull image {} failed, error is [{}]".format (image, str(err)))
        else:
            if expected_error_message is not None:
                if str(ret).lower().find(expected_error_message.lower()) < 0:
                    raise Exception(r" Failed to catch error [{}] when pull image {}, return message: {}".format (expected_error_message, image, str(ret)))
            else:
                if str(ret).lower().find("error".lower()) >= 0:
                    raise Exception(r" It's was not suppose to catch error when pull image {}, return message is [{}]".format (image, ret))

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
        if expected_error_message is "":
            expected_error_message = None
        try:
            ret = self.DCLIENT.push(harbor_registry, tag)
            print("Docker image push commond return:", ret)
        except Exception as err:
            print( "docker image push catch Exception:", str(err))
            if expected_error_message is not None:
                print( "docker image push error:", str(err))
                if str(err).lower().find(expected_error_message.lower()) < 0:
                    raise Exception(r"Push image: Return message {} is not as expected {}".format(str(err), expected_error_message))
            else:
                raise Exception(r" Docker push image {} failed, error is [{}]".format (harbor_registry, message))
        else:
            print( "docker image push does not catch Exception:", str(expected_error_message))
            if expected_error_message is not None:
                if str(ret).lower().find(expected_error_message.lower()) < 0:
                    raise Exception(r" Failed to catch error [{}] when push image {}, return message: {}".
                                    format (expected_error_message, harbor_registry, str(ret)))
                else:
                    print("docker image push action return expected error message [{}]".format(expected_error_message))

            else:
                if str(ret).lower().find("errorDetail".lower()) >= 0:
                    raise Exception(r" It's was not suppose to catch error when push image {}, return message is [{}]".
                                    format (harbor_registry, ret))

    def docker_image_build(self, harbor_registry, tags=None, size=1, expected_error_message = None):
        ret = ""
        try:
            baseimage='busybox:latest'
            self.DCLIENT.login(username=DOCKER_USER, password=DOCKER_PWD)
            if not self.DCLIENT.images(name=baseimage):
                print( "docker pull is triggered when building {}".format(harbor_registry))
                self.DCLIENT.pull(baseimage)
            c=self.DCLIENT.create_container(image='busybox:latest',command='dd if=/dev/urandom of=test bs=1M count=%d' % size )
            self.DCLIENT.start(c)
            self.DCLIENT.wait(c)
            if not tags:
                tags=['latest']
            firstrepo="%s:%s" % (harbor_registry, tags[0])
            #self.DCLIENT.commit(c, firstrepo)
            self.DCLIENT2.containers.get(c).commit(harbor_registry, tags[0])
            for tag in tags[1:]:
                repo="%s:%s" % (harbor_registry, tag)
                self.DCLIENT.tag(firstrepo, repo)
            for tag in tags:
                repo="%s:%s" % (harbor_registry, tag)
                ret = self.DCLIENT.push(repo)
                print("docker_image_push ret:", ret)
                print("build image %s with size %d" % (repo, size))
                self.DCLIENT.remove_image(repo)
            self.DCLIENT.remove_container(c)
            #self.DCLIENT.pull(repo)
            #image = self.DCLIENT2.images.get(repo)
            return repo
        except Exception as err:
            if expected_error_message is not None:
                print( "docker image build error:", str(err))
                if str(err).lower().find(expected_error_message.lower()) < 0:
                    raise Exception(r"Push image: Return message {} is not as expected {}".format(str(err), expected_error_message))
            else:
                raise Exception(r" Docker build image {} failed, error is [{}]".format (harbor_registry, str(err)))
        else:
            print("docker image build does not catch Exception:", str(expected_error_message))
            print("Docker build -> docker image push ret:", ret)
            if expected_error_message is not None:
                if str(ret).lower().find(expected_error_message.lower()) < 0:
                    raise Exception(r" Failed to catch error [{}] when build image {}, return message: {}".
                                    format (expected_error_message, harbor_registry, str(ret)))
                else:
                    print("docker image build return expected error message [{}]".format(expected_error_message))
            else:
                if str(ret).lower().find("errorDetail".lower()) >= 0:
                    raise Exception(r" It's was not suppose to catch error when push image {}, return message is [{}]".format (harbor_registry, ret))
