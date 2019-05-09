# Copyright Project Harbor Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License

*** Settings ***
Documentation  This resource contains keywords related to creating and using certificates. Requires scripts in infra/integration-image/scripts be available in PATH

*** Keywords ***
Generate Certificate Authority For Chrome
    #  add the ca to chrome trust list to enable https testing.
    [Arguments]  ${password}=%{HARBOR_PASSWORD}
    ${rand}=  Evaluate  random.randint(0, 100000)  modules=random
    Log To Console  Generate Certificate Authority For Chrome
    ${rc}  ${out}=  Run And Return Rc And Output  echo ${password} > password${rand}.ca
    Log  ${out}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${out}=  Run And Return Rc And Output  certutil -d sql:$HOME/.pki/nssdb -A -t TC -f password${rand}.ca -n "Harbor${rand}" -i ./harbor_ca.crt
    Log  ${out}
    Should Be Equal As Integers  ${rc}  0

Generate Certificate Authority
    #  Generates CA (private/ca.key.pem, certs/ca.cert.pem, certs/STARK_ENTERPRISES_ROOT_CA.crt) in OUT_DIR
    [Arguments]  ${CA_NAME}=STARK_ENTERPRISES_ROOT_CA  ${OUT_DIR}=/root/ca
    Log To Console  Generating Certificate Authority
    ${rc}  ${out}=  Run And Return Rc And Output  generate-ca.sh -c ${CA_NAME} -d ${OUT_DIR}
    Log  ${out}
    Should Be Equal As Integers  ${rc}  0

Generate Wildcard Server Certificate
    # Generates key and signs with CA for *.DOMAIN (csr/*.DOMAIN.csr.pem,
    # private/*.DOMAIN.key.pem, certs/*.DOMAIN.cert.pem) in OUT_DIR
    [Arguments]  ${DOMAIN}=%{DOMAIN}  ${OUT_DIR}=/root/ca  ${CA_NAME}=STARK_ENTERPRISES_ROOT_CA
    Log To Console  Generating Wildcard Server Certificate
    Run Keyword  Generate Server Key And CSR  *.${DOMAIN}  ${OUT_DIR}
    Run Keyword  Sign Server CSR  ${CA_NAME}  *.${DOMAIN}  ${OUT_DIR}
    Run Keyword  Create Certificate Bundle  CA_NAME=${CA_NAME}  SRC_DIR=${OUT_DIR}  CN=*.${DOMAIN}
    ${out}=  Run  ls -al ${OUT_DIR}/csr
    Log  ${out}
    ${out}=  Run  ls -al ${OUT_DIR}/private
    Log  ${out}
    ${out}=  Run  ls -al ${OUT_DIR}/certs
    Log  ${out}


Generate Server Key And CSR
    # Generates key and CSR (private/DOMAIN.key.pem, csr/DOMAIN.csr.pem) in OUT_DIR
    [Arguments]  ${CN}=%{DOMAIN}  ${OUT_DIR}=/root/ca
    Log To Console  Generating Server Key And CSR
    ${out}=  Run  generate-server-key-csr.sh -d ${OUT_DIR} -n ${CN}
    Log  ${out}


Sign Server CSR
    # Generates certificate signed by CA (certs/DOMAIN.cert.pem) in OUT_DIR
    [Arguments]  ${CA_NAME}=STARK_ENTERPRISES_ROOT_CA  ${CN}=%{DOMAIN}  ${OUT_DIR}=/root/ca
    Log To Console  Signing Server CSR
    ${out}=  Run  sign-csr.sh -c ${CA_NAME} -d ${OUT_DIR} -n ${CN}
    Log  ${out}


Trust Certificate Authority
    # Installs root certificate into trust store on Debian based distro
    [Arguments]  ${CRT_FILE}=/root/ca/certs/STARK_ENTERPRISES_ROOT_CA.crt
    Log To Console  Installing CA
    ${rc}  ${out}=  Run And Return Rc And Output  ubuntu-install-ca.sh -f ${CRT_FILE}
    Should Be Equal As Integers  ${rc}  0
    Log  ${out}


Reload Default Certificate Authorities
    # Reloads default certificates into trust store on Debian based distro
    # Removes all user provided CAs
    Log To Console  Reloading Default CAs
    ${rc}  ${out}=  Run And Return Rc And Output  ubuntu-reload-cas.sh
    Should Be Equal As Integers  ${rc}  0
    Log  ${out}


Create Certificate Bundle
    [Arguments]  ${CA_NAME}=STARK_ENTERPRISES_ROOT_CA  ${SRC_DIR}=/root/ca  ${OUT_FILE}=/root/ca/cert-bundle.tgz  ${CN}=%{DOMAIN}  ${TMP_DIR}=/root/ca/bundle
    ${rc}  ${out}=  Run And Return Rc And Output  bundle-certs.sh -c ${CA_NAME} -d ${SRC_DIR} -f ${OUT_FILE} -n ${CN} -o ${TMP_DIR}
    Should Be Equal As Integers  ${rc}  0
    Log  ${out}


Get Certificate Authority CRT
    # Return ascii armored certificate from file e.g. `-----BEGIN CERTIFICATE-----`
    [Arguments]  ${CA_CRT}=STARK_ENTERPRISES_ROOT_CA.crt  ${DIR}=/root/ca/certs
    ${out}=  Run  cat ${DIR}/${CA_CRT}
    [Return]  ${out}


Get Server Certificate
    # Return ascii armored certificate from file e.g. `-----BEGIN CERTIFICATE-----`
    # PEM must be provided if using a wildcard cert not specified by DOMAIN
    [Arguments]  ${PEM}=%{DOMAIN}.cert.pem  ${DIR}=/root/ca/certs
    ${out}=  Run  cat ${DIR}/${PEM}
    [Return]  ${out}


Get Server Key
    # Return ascii armored key from file e.g. `-----BEGIN RSA PRIVATE KEY-----`
    # PEM must be provided if using a wildcard cert not specified by DOMAIN
    [Arguments]  ${PEM}=%{DOMAIN}.key.pem  ${DIR}=/root/ca/private
    ${out}=  Run  cat ${DIR}/${PEM}
    [Return]  ${out}
