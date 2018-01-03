from __future__ import absolute_import

import logging
import sys

default_handler = logging.StreamHandler(sys.stderr)
default_handler.setFormatter(logging.Formatter(
    '[%(asctime)s] %(levelname)s in %(module)s: %(message)s'
))

def create_logger():
    logger = logging.getLogger('nightly.test')
    logger.setLevel(logging.DEBUG)
    logger.addHandler(default_handler) 
    return logger