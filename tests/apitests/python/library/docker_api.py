# -*- coding: utf-8 -*-

import base
import subprocess
import json
from testutils import DOCKER_USER, DOCKER_PWD, BASE_IMAGE, BASE_IMAGE_ABS_PATH_NAME

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
    command = ["docker", "login", harbor_host, "-u", username, "-p", password]
    base.run_command(command)
    if enable_manifest == True:
        try:
            ret = subprocess.check_output([cfg_file], shell=False)
            print("docker login cmd ret:", ret)
        except subprocess.CalledProcessError as exc:
            raise Exception("Failed to update docker config, error is {} {}.".format(exc.returncode, exc.output))

def docker_manifest_create(index, manifests):
    command = ["docker","manifest","create", "--amend", index]
    command.extend(manifests)
    print( "Docker Manifest Command: ", command)
    base.run_command(command)

def docker_images_all_list():
    command = ["docker","images","-a"]
    base.run_command(command)

def docker_load_image(image):
    command = ["docker","load","-i", image]
    base.run_command(command)

def docker_image_clean_all():
    docker_images_all_list()
    command = ["docker rmi -f $(docker images -a -q)"]
    base.run_command_with_popen(command)
    command = ["docker","system", "prune", "-a", "-f"]
    base.run_command(command)
    docker_images_all_list()

def docker_manifest_push(index):
    command = ["docker","manifest","push",index]
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
    ret = base.run_command(command)
    tags = json.loads(ret).get("tags","")
    return tags

class DockerAPI(object):
    def __init__(self):
        self.DCLIENT = docker.APIClient(base_url='unix://var/run/docker.sock',version='auto',timeout=30)
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
                    raise Exception(r" Failed to catch error [{}] when login registry {}, return message: {}".format (expected_error_message, registry, err_message))
                else:
                    print(r"Docker image login got expected error message:{}".format(expected_error_message))
            else:
                if str(err_message).lower().find("error".lower()) >= 0:
                    raise Exception(r" It's was not suppose to catch error when login registry {}, return message is [{}]".format (registry, err_message))

    def docker_image_pull(self, image, tag = None, expected_error_message = None, is_clean_all_img = True):
        ret = ""
        err_message = ""
        if tag is not None:
            _tag = tag
        else:
            _tag = "latest"
        if expected_error_message == "":
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
            if is_clean_all_img:
                docker_image_clean_all()

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
        docker_images_all_list()
        if expected_error_message == "":
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
        docker_images_all_list()

    def docker_image_build(self, harbor_registry, tags=None, size=1, expected_error_message = None):
        ret = ""
        err_message = ""
        docker_images_all_list()
        try:
            baseimage = BASE_IMAGE['name'] + ":" + BASE_IMAGE['tag']
            if not self.DCLIENT.images(name=baseimage):
                print( "Docker load is triggered when building {}".format(harbor_registry))
                docker_load_image(BASE_IMAGE_ABS_PATH_NAME)
                docker_images_all_list()
            c = self.DCLIENT.create_container(image=baseimage,
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
            self.DCLIENT.remove_container(c)
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
                    print(r"Docker image build got expected error message: {}".format(expected_error_message))
            else:
                if str(err_message).lower().find("error".lower()) >= 0:
                    raise Exception(r" It's was not suppose to catch error when build image {}, return message is [{}]".format (harbor_registry, err_message))
            docker_image_clean_all()
