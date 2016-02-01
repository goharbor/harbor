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
var validateOptions = {
	"Result" : [],
	"Items" : [],
	"Validate": function(callback){
		for(var i = 0; i < this.Items.length; i++){
			if(validateCallback(this.Items[i]) == false){
				return false;
			}
		}
		callback();
	},
	"Username" :{
		"Required": { "value" : true, "errMsg" : i18n.getMessage("username_is_required")},
		"CheckExist": { "value" : function(value){
				var result = true;
				$.ajax({
					 url: "/userExists",
					data: {"target": "username", "value" : value},
				dataType: "json",
					type: "post",
					async: false,
		  		 success: function(data){
						result = data;
					}
				});
				return result;
		}, "errMsg" : i18n.getMessage("username_has_been_taken")},
		"MaxLength": {"value" : 20, "errMsg" : i18n.getMessage("username_is_too_long")},
		"IllegalChar": {"value": [",","~","#", "$", "%"] , "errMsg": i18n.getMessage("username_contains_illegal_chars")}
	},	
	"Email" :{
		"Required": { "value" : true, "errMsg" : i18n.getMessage("email_is_required")},
		"RegExp": {"value": /^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/, 
				   "errMsg": i18n.getMessage("email_contains_illegal_chars")},
		"CheckExist": { "value" : function(value){
					var result = true;
					$.ajax({
						 url: "/userExists",
						data: {"target": "email", "value": value},
					dataType: "json",
						type: "post",
						async: false,
			  		 success: function(data){
							result = data;
						}
					});
					return result;
			}, "errMsg" : i18n.getMessage("email_has_been_taken")}
	},
	"EmailF" :{
		"Required": { "value" : true, "errMsg" : i18n.getMessage("email_is_required")},
		"RegExp": {"value": /^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/, 
				   "errMsg": i18n.getMessage("email_content_illegal")},
		"CheckIfNotExist": { "value" : function(value){
					var result = true;
					$.ajax({
						 url: "/userExists",
						data: {"target": "email", "value": value},
					dataType: "json",
						type: "post",
						async: false,
			  		 success: function(data){
							result = data;
						}
					});
					return result;
			}, "errMsg" : i18n.getMessage("email_does_not_exist")}
	},
	"Realname" :{
		"Required": { "value" : true, "errMsg" : i18n.getMessage("realname_is_required")},
		"MaxLength": {"value" : 20, "errMsg" : i18n.getMessage("realname_is_too_long")},
		"IllegalChar": {"value": [",","~","#", "$", "%"] , "errMsg": i18n.getMessage("realname_contains_illegal_chars")}
	},
	"OldPassword" :{
		"Required": { "value" : true, "errMsg" : i18n.getMessage("password_is_required")}
	},
	"Password" :{
		"Required": { "value" : true, "errMsg" : i18n.getMessage("password_is_required")},
		"RegExp": {"value" : /^(?=.*\d)(?=.*[a-z])(?=.*[A-Z])(?!.*\s).{7,20}$/,  "errMsg" : i18n.getMessage("password_is_invalid")},
		"MaxLength": {"value" : 20, "errMsg" : i18n.getMessage("password_is_too_long")}
	},
	"ConfirmedPassword" :{
		"CompareWith": {"value" : "#Password", "errMsg" : i18n.getMessage("password_does_not_match")},
		"RegExp": {"value" : /^(?=.*\d)(?=.*[a-z])(?=.*[A-Z])(?!.*\s).{7,20}$/,  "errMsg" : i18n.getMessage("password_is_invalid")}
	},
	"Comment" :{
		"MaxLength": {"value" : 20, "errMsg" : i18n.getMessage("comment_is_too_long")},
		"IllegalChar": {"value": [",","~","#", "$", "%"] , "errMsg": i18n.getMessage("comment_contains_illegal_chars")}
	},
	"projectName" :{
		"Required": { "value" : true, "errMsg" : i18n.getMessage("project_name_is_required")},
		"MinLength": {"value" : 4, "errMsg" : i18n.getMessage("project_name_is_too_short")},
		"MaxLength": {"value" : 30, "errMsg" : i18n.getMessage("project_name_is_too_long")},
		"IllegalChar": {"value": ["~","$","-", "\\", "[", "]", "{", "}", "(", ")", "&", "^", "%", "*", "<", ">", "\"", "'","/","?","@"] , "errMsg": i18n.getMessage("project_name_contains_illegal_chars")}
	}
};	
function validateCallback(target){
	
	if (typeof target != "string"){
		target = this;
	}
	
	var isValid = true;
	var inputValue = $.trim($(target).val());
	var currentId  = $(target).attr("id");
	var validateItem = validateOptions[currentId];
	
    var errMsg = "";

	for(var checkTitle in validateItem){

		var checkValue = validateItem[checkTitle].value;
		
		if(checkTitle == "Required" && checkValue && inputValue.length == 0){
			isValid = false;
			errMsg = validateItem[checkTitle].errMsg;
			break;
		}else if(checkTitle == "CheckOldPasswordIsCorrect" && checkValue(inputValue) == false){		
			isValid = false;
			errMsg = validateItem[checkTitle].errMsg;
			break;
		}else if(checkTitle == "CheckExist" && checkValue(inputValue) == true){		
			isValid = false;
			errMsg = validateItem[checkTitle].errMsg;
			break;
		}else if(checkTitle == "CheckIfNotExist" && checkValue(inputValue) == false){		
			isValid = false;
			errMsg = validateItem[checkTitle].errMsg;
			break;
		}else if(checkTitle == "RegExp" && checkValue.test(inputValue) == false){
			isValid = false;
			errMsg = validateItem[checkTitle].errMsg;
			break;
		}else if(checkTitle == "MinLength" && inputValue.length < checkValue){
			isValid = false;
			errMsg = validateItem[checkTitle].errMsg;
			break;
		}else if(checkTitle == "MaxLength" && inputValue.length > checkValue){
			isValid = false;
			errMsg = validateItem[checkTitle].errMsg;
			break;
		}else if(checkTitle == "CompareWith" && $.trim($(checkValue).val()).length > 0 && inputValue != $.trim($(checkValue).val())){
			isValid = false;
			errMsg = validateItem[checkTitle].errMsg;
			break;
		}else if(checkTitle == "IllegalChar"){
			for(var i = 0; i < checkValue.length; i++){
				if(inputValue.indexOf(checkValue[i]) > -1){
					isValid = false;
					errMsg = validateItem[checkTitle].errMsg;
				}
			}
			break;
		}
	}	
	
	if(isValid == false){
		$(target).parent().removeClass("has-success").addClass("has-error");
		$(target).siblings("span").removeClass("glyphicon-ok").addClass("glyphicon-warning-sign");
		$("#divErrMsg").css({"display": "block"});
		$("#divErrMsg").text(errMsg);
	}else {
		$(target).parent().removeClass("has-error").addClass("has-success");
		$(target).siblings("span").removeClass("glyphicon-warning-sign").addClass("glyphicon-ok");
		$("#divErrMsg").css({"display": "none"});			
	}	
	validateOptions.Result.push(isValid);
	return isValid;
}