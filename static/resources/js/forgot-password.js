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
	
	validateOptions.Items = ["#EmailF"];
	function bindEnterKey(){
		$(document).on("keydown", function(e){
			if(e.keyCode == 13){
			  e.preventDefault();
			  if($("#txtCommonSearch").is(":focus")){
				document.location = "/search?q=" + $("#txtCommonSearch").val();
			  }else{
			    $("#btnSubmit").trigger("click");	
			  }
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
			var username = $("#UsernameF").val();
			var email = $("#EmailF").val();
			$.ajax({
				"url":"/sendEmail",
				"type": "get",
				"data": {"username": username, "email": email},
				"beforeSend": function(e){
				   unbindEnterKey();
				   $("h1").append(spinner.el);
				   $("#btnSubmit").prop("disabled", true);	
				},
				"success": function(data, status, xhr){
					if(xhr && xhr.status == 200){
						$("#dlgModal")
							.dialogModal({
								"title": i18n.getMessage("title_forgot_password"), 
								"content": i18n.getMessage("email_has_been_sent"), 
								"callback": function(){
									document.location="/";
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
								"title": i18n.getMessage("title_forgot_password"), 
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