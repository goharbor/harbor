# -*- coding: utf-8 -*-
import base
import os

def generate_key_pair():
    config_key_file = "cosign.key"
    config_pub_file = "cosign.pub"
    if os.path.exists(config_key_file) and os.path.exists(config_pub_file):
        os.remove(config_key_file)
        os.remove(config_pub_file)
    command = ["cosign", "generate-key-pair"]
    base.run_command(command)

def sign_artifact(artifact):
    command = ["cosign", "sign", "-y", "--allow-insecure-registry", "--key", "cosign.key", artifact]
    base.run_command(command)

def push_artifact_sbom(artifact, sbom_path, type="spdx"):
    command = ["cosign", "attach", "sbom", "--allow-insecure-registry", "--registry-referrers-mode", "oci-1-1",
               "--type", type, "--sbom", sbom_path, artifact]
    base.run_command(command)
