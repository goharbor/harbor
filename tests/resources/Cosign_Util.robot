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
Cosign Generate Key Pair
    Remove Files  cosign.key  cosign.pub
    Wait Unitl Command Success  cosign generate-key-pair

Cosign Sign
    [Arguments]  ${artifact}
    Wait Unitl Command Success  cosign sign --allow-insecure-registry --key cosign.key ${artifact}

Cosign Verify
    [Arguments]  ${artifact}  ${signed}
    Run Keyword If  ${signed}==${true}  Wait Unitl Command Success  cosign verify --key cosign.pub ${artifact}
    ...  ELSE  Command Should be Failed  cosign verify --key cosign.pub ${artifact}