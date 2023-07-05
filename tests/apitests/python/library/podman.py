# -*- coding: utf-8 -*-
import base

def login(registry, username, password):
    command = ["podman", "login", "-u", username, "-p", password, registry]
    base.run_command(command)

def logout(registry):
    command = ["podman", "logout", registry]
    base.run_command(command)

def pull(artifact):
    command = ["podman", "pull", artifact]
    base.run_command(command)

def push(source_artifact, target_artifact):
    command = ["podman", "push", source_artifact, target_artifact]
    base.run_command(command)
