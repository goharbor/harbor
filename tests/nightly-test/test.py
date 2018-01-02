import os
import sys

dir_path = os.path.dirname(os.path.realpath(__file__))
sys.path.append(dir_path + '/utils')

import test_executor

test_executor.execute_test_ova('sc-rdops-vm10-dhcp-60-192.eng.vmware.com', 'zhu88jie', 'Nightly')