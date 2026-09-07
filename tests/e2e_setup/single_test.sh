#!/usr/bin/env bash

set -euo pipefail

# Usage: ./single_test.sh [include_tag]
INCLUDE_TAG="${1:-user_view_logs}"

robot -V /drone/tests/e2e_setup/robotvars.py /drone/tests/robot-cases/Group1-Nightly/Setup_Nightly.robot

robot --include "${INCLUDE_TAG}" -V /drone/tests/e2e_setup/robotvars.py \
	/drone/tests/robot-cases/Group1-Nightly/Setup_Nightly.robot \
	/drone/tests/robot-cases/Group1-Nightly/Common_GC.robot \
	/drone/tests/robot-cases/Group1-Nightly/Webhook.robot \
	/drone/tests/robot-cases/Group1-Nightly/Routing.robot \
	/drone/tests/robot-cases/Group1-Nightly/P2P_Preheat.robot \
	/drone/tests/robot-cases/Group1-Nightly/Trivy.robot \
	/drone/tests/robot-cases/Group1-Nightly/DB.robot \
	/drone/tests/robot-cases/Group1-Nightly/Common.robot \
	/drone/tests/robot-cases/Group1-Nightly/Teardown.robot
