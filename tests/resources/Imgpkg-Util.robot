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
Imgpkg Push
    [Arguments]  ${server}  ${project}  ${repository}  ${tag}  ${directory}
    Wait Unitl Command Success  imgpkg push -b ${server}/${project}/${repository}:${tag} -f ${directory}

Imgpkg Pull
    [Arguments]  ${server}  ${project}  ${repository}  ${tag}  ${directory}
    Wait Unitl Command Success  imgpkg pull -b ${server}/${project}/${repository}:${tag} -o ${directory}

Imgpkg Copy From Registry To Registry
    [Arguments]  ${source_registry}  ${target_repo}
    Wait Unitl Command Success  imgpkg copy -b ${source_registry} --to-repo=${target_repo}

Imgpkg Copy From Registry To Local Tarball
    [Arguments]  ${source_registry}  ${file_path}
    Wait Unitl Command Success  imgpkg copy -b ${source_registry} --to-tar=${file_path}

Imgpkg Copy From Local Tarball To Registry
    [Arguments]  ${file_path}  ${target_repo}
    Wait Unitl Command Success  imgpkg copy --tar ${file_path} --to-repo=${target_repo}
