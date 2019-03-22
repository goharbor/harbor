import os, shutil

def prepare_uaa_cert_file(uaa_ca_cert, core_cert_dir):
    if os.path.isfile(uaa_ca_cert):
        if not os.path.isdir(core_cert_dir):
            os.makedirs(core_cert_dir)
        core_uaa_ca = os.path.join(core_cert_dir, "uaa_ca.pem")
        print("Copying UAA CA cert to %s" % core_uaa_ca)
        shutil.copyfile(uaa_ca_cert, core_uaa_ca)
    else:
        print("Can not find UAA CA cert: %s, skip" % uaa_ca_cert)