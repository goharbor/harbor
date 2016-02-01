/*
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
*/
jQuery(function(){
	$("#divErrMsg").css({"display": "none"});
	
	$("#Password,#ConfirmedPassword").on("blur", validateCallback);
	validateOptions.Items = ["#Password", "#ConfirmedPassword"];
    function bindEnterKey(){
		$(document).on("keydown", function(e){
			if(e.keyCode == 13){
			  e.preventDefault();
			  $("#btnSubmit").trigger("click");				  
			}
		});
	}
	function unbindEnterKey(){
		$(document).off("keydown");
	}
	bindEnterKey();
	
	
	var spinner = new Spinner({scale:1}).spin();
	
	$("#btnSubmit").on("click", function(){
		validateOptions.Validate(function(){
			var resetUuid = $("#resetUuid").val();
			var password = $("#Password").val();
			$.ajax({
				"url": "/reset",
				"type": "post",
				"data": {"reset_uuid": resetUuid, "password": password},
				"beforeSend": function(e){
				   unbindEnterKey();
				   $("h1").append(spinner.el);
				   $("#btnSubmit").prop("disabled", true);	
				},
				"success": function(data, status, xhr){					
					if(xhr && xhr.status == 200){
						$("#dlgModal")
							.dialogModal({
								"title": i18n.getMessage("title_reset_password"), 
								"content": i18n.getMessage("reset_password_successfully"),
							    "callback": function(){ 
									document.location="/signIn";
								}
							});
					}	
					
				},
				"complete": function(){
					spinner.stop();
					$("#btnSubmit").prop("disabled", false);		
				},
				"error": function(jqXhr, status, error){
					if(jqXhr){
						$("#dlgModal")
							.dialogModal({
								"title": i18n.getMessage("title_reset_password"), 
								"content": i18n.getMessage(jqXhr.responseText),
							    "callback": function(){ 
									bindEnterKey();
									return;
								}
							});
					}
				}
			});
		});
	});
});