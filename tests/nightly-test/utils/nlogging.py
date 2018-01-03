from __future__ import absolute_import

import logging
import sys

def create_logger():
    default_handler = logging.StreamHandler(sys.stderr)
    default_handler.setFormatter(logging.Formatter(
        '[%(asctime)s] %(levelname)s in %(module)s: %(message)s'
    ))
    logger = logging.getLogger('nightly.test')
    logger.setLevel(logging.DEBUG)
    logger.addHandler(default_handler) 
    return logger