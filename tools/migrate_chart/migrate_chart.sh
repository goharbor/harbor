#! /bin/bash

HELM_CMD=$1
V2_CHART_PATH=$2
OCI_REF=$3

${HELM_CMD} chart save ${V2_CHART_PATH} ${OCI_REF}
${HELM_CMD} chart push ${OCI_REF}
${HELM_CMD} chart remove ${OCI_REF}
