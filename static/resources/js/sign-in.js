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
	
	new AjaxUtil({
		url: "/api/users/current",
		type: "get",
		success: function(data, status, xhr){
			if(xhr && xhr.status == 200){
				document.location = "/registry/project";
			}
		},
		error: function(jqxhr){
			return false;
		}
	}).exec();
	
	$(document).on("keydown", function(e){
		if(e.keyCode == 13){
			e.preventDefault();
			if($("#Principal").is(":focus") || $("#Password").is(":focus") || $("#btnPageSignIn").is(":focus")){
				$("#btnPageSignIn").trigger("click");
			}
		}
	});
	$("#btnForgot").on("click", function(){
		document.location = "/forgotPassword";
	});
	
	$("#btnPageSignIn").on("click", function(){
		
		var principal = $.trim($("#Principal").val());		
		var password = $.trim($("#Password").val());
		
		if($.trim(principal).length <= 0 || $.trim(password).length <= 0) {
			$("#dlgModal").dialogModal({"title": i18n.getMessage("title_login_failed"), "content": i18n.getMessage("input_your_username_and_password")});
			return;
		}
		
		$.ajax({
			url:'/login',
			data: {principal: principal, password: password},
			type: "post",
			success: function(jqXhr, status){
				var lastUri = location.search;
				if(lastUri != "" && lastUri.indexOf("=") > 0){
				    document.location = decodeURIComponent(lastUri.split("=")[1]);
				}else{
					document.location = "/registry/project";						
				}
			},
			error: function(jqXhr){
				var i18nKey = "";
				if(jqXhr.status == 500){
					i18nKey = "internal_error";
				}else{
					i18nKey = "check_your_username_or_password"
				}
				$("#dlgModal")
					.dialogModal({
						"title": i18n.getMessage("title_login_failed"),
						"content": i18n.getMessage(i18nKey)
					});
			}
		});
	});
});