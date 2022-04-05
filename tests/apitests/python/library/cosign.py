# -*- coding: utf-8 -*-
import base

def generate_key_pair():
    command = ["cosign", "generate-key-pair"]
    base.run_command(command)

def sign_artifact(artifact):
    command = ["cosign", "sign", "--allow-insecure-registry", "--key", "cosign.key", artifact]
    base.run_command(command)
