<!--
    Copyright (c) 2016 VMware, Inc. All Rights Reserved.
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at
        
        http://www.apache.org/licenses/LICENSE-2.0
        
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
-->
<style>
.center {
    margin-left: auto;
    margin-right: auto;
    top: 10%;
}
</style>
<!-- Modal -->
<div class="center modal fade" id="dlgModal" tabindex="-1" role="dialog" aria-labelledby="myModalLabel" aria-hidden="true">
  <div class="modal-dialog">
    <div class="modal-content">
      <div class="modal-header">
        <button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
        <h4 class="modal-title" id="dlgLabel"></h4>
      </div>
      <div class="modal-body" id="dlgBody">
      </div>
      <div class="modal-footer">
        <button type="button" class="btn btn-primary" id="dlgConfirm" data-dismiss="modal">{{i18n .Lang "dlg_button_ok"}}</button>
		<button type="button" class="btn btn-primary" id="dlgCancel" data-dismiss="modal" style="display: none;">{{i18n .Lang "dlg_button_cancel"}}</button>
      </div>
    </div>
  </div>
</div>