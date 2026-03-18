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


# known issue for proxy ennvironment https://github.com/sigstore/cosign/issues/3269
def sign_artifact(artifact):
    print("*******Start coisgn sign artifact********")
    allow_insecure = base.getenv_bool("ALLOW_INSECURE", default=True)
    if allow_insecure:
        command = ["cosign", "sign", "-y", "--allow-insecure-registry", "--key", "cosign.key", artifact]
    else:
        command = ["cosign", "sign", "-y", "--key", "cosign.key", artifact]
    base.run_command(command)

def push_artifact_sbom(artifact, sbom_path, type="spdx"):
    allow_insecure = base.getenv_bool("ALLOW_INSECURE", default=True)
    if allow_insecure:
        command = ["cosign", "attach", "sbom", "--allow-insecure-registry", "--registry-referrers-mode", "oci-1-1",
               "--type", type, "--sbom", sbom_path, artifact]
    else:
        command = ["cosign", "attach", "sbom", "--registry-referrers-mode", "oci-1-1",
               "--type", type, "--sbom", sbom_path, artifact]
    base.run_command(command)
