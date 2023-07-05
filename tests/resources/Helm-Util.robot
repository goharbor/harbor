# Copyright Project Harbor Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#	http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License

*** Settings ***
Documentation  This resource provides helper functions for docker operations
Library  OperatingSystem
Library  Process

*** Keywords ***
Helm Registry Login
    [Arguments]  ${ip}  ${user}  ${password}
    Wait Unitl Command Success  helm registry login ${ip} -u ${user} -p ${password} --insecure

Helm Package
    [Arguments]  ${file_path}
    Wait Unitl Command Success  helm package ${file_path}

Helm Push
    [Arguments]  ${file_path}  ${ip}  ${repo_name}
    Wait Unitl Command Success  helm push ${file_path} oci://${ip}/${repo_name} --insecure-skip-tls-verify

Helm Pull
    [Arguments]  ${ip}  ${repo_name}  ${version}
    Wait Unitl Command Success  helm pull oci://${ip}/${repo_name}/harbor --version ${version} --insecure-skip-tls-verify

Helm Registry Logout
    [Arguments]  ${ip}
    Wait Unitl Command Success  helm registry logout ${ip}
