# -*- coding: utf-8 -*-
import base

def generate_cert():
    command = ["notation", "cert", "generate-test", "--default", "wabbit-networks.io"]
    base.run_command(command)

def sign_artifact(artifact):
    command = ["notation", "sign", "-d", "--allow-referrers-api", artifact]
    base.run_command(command)
