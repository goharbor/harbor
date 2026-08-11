# -*- coding: utf-8 -*-
import base

def generate_cert(common_name = "wabbit-networks.io"):
    # notation refuses to overwrite an existing key, so a test that runs in the same suite as
    # another notation test has to ask for a common name of its own.
    command = ["notation", "cert", "generate-test", "--default", common_name]
    base.run_command(command)

def sign_artifact(artifact):
    command = ["notation", "sign", "-d", "--allow-referrers-api", artifact]
    base.run_command(command)
