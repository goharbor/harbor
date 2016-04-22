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
	
	$(document).on("keydown", function(e){
		
		if(e.keyCode == 13){
			e.preventDefault();
			if(!$("#txtCommonSearch").is(":focus")){								
				$("#btnPageSignUp").trigger("click");
			}
		}
	});
	
	$("#Username,#Email,#Realname,#Password,#ConfirmedPassword").on("blur", validateCallback);
	validateOptions.Items = ["#Username","#Email","#Realname","#Password","#ConfirmedPassword"];
	
	$("#btnPageSignUp").on("click", function(){
		validateOptions.Validate(function() {
            var username = $.trim($("#Username").val());
            var email	 = $.trim($("#Email").val());
            var password = $.trim($("#Password").val());
            var confirmedPassword = $.trim($("#ConfirmedPassword").val());
            var realname = $.trim($("#Realname").val());
            var comment  = $.trim($("#Comment").val());
            var isAdmin = $("#isAdmin").val();
            
			new AjaxUtil({
				url : "/api/users",
				data: {"username": username, "password": password, "realname": realname, "comment": comment, "email": email},
				type: "POST",
				beforeSend: function(e){
					$("#btnPageSignUp").prop("disabled", true);
				},
				error:function(jqxhr, status, error){
					$("#dlgModal")
							.dialogModal({
								"title": i18n.getMessage("title_sign_up"), 
								"content": i18n.getMessage("internal_error"),
								"callback": function(){ 								
									return;
								}
							});
				},
				complete: function(xhr, status){
					$("#btnPageSignUp").prop("disabled", false);
					if(xhr && xhr.status == 201){
						$("#dlgModal")
							.dialogModal({
								"title":  isAdmin == "true" ? i18n.getMessage("title_add_user") : i18n.getMessage("title_sign_up"), 
								"content": isAdmin == "true" ? i18n.getMessage("added_user_successfully") : i18n.getMessage("registered_successfully"),
								"callback": function(){
                                    if(isAdmin == "true") {
                                      document.location = "/registry/project"; 
                                    }else{	
                                      document.location = "/signIn";       
                                    }
								}
							});
					}
				}
			}).exec();
		});
	});
});