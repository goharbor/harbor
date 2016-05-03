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
var AjaxUtil = function(params){
		
	this.url = params.url;
	this.data = params.data;
	this.dataRaw = params.dataRaw;
	this.type = params.type;
	this.errors = params.errors || {};
	
	this.success = params.success;
	this.complete = params.complete;
	this.error = params.error;
	
};

AjaxUtil.prototype.exec = function(){
	
	var self = this;
	
	return $.ajax({
      url: self.url,
	  contentType: (self.dataRaw ? "application/x-www-form-urlencoded; charset=UTF-8" : "application/json; charset=utf-8"),
      data: JSON.stringify(self.data) || self.dataRaw,
      type: self.type,
	  dataType: "json",
	  success: function(data, status, xhr){
		if(self.success != null){
		  self.success(data, status, xhr);
		}
	  },
      complete: function(jqXhr, status) {
		if(self.complete != null){
	 	  self.complete(jqXhr, status);				
		}
      },
      error: function(jqXhr){
	    if(self.error != null){
			self.error(jqXhr);
		}else{
			var errorMessage = self.errors[jqXhr.status] || jqXhr.responseText;
			if(jqXhr.status == 401){				
				var lastUri = location.pathname + location.search;
				if(lastUri != ""){
				    document.location = "/signIn?uri=" + encodeURIComponent(lastUri);
				}else{
					document.location = "/signIn";						
				}
			}else if($.trim(errorMessage).length > 0){
			    $("#dlgModal").dialogModal({"title": i18n.getMessage("operation_failed"), "content": errorMessage});
			}
		}
	  }
    });
};

var SUPPORT_LANGUAGES = {
	"en-US": "English",
	"zh-CN": "Chinese",
	"de-DE": "German"
};

var DEFAULT_LANGUAGE = "en-US";

var I18n = function(messages) {
	this.messages = messages;
};

I18n.prototype.isSupportLanguage = function(lang){
	return (lang in SUPPORT_LANGUAGES);
}

I18n.prototype.getLocale = function(){
	var lang = $("#currentLanguage").val();
	if(this.isSupportLanguage(lang)){
		return lang;
	}else{
		return DEFAULT_LANGUAGE;
	}
};

I18n.prototype.getMessage = function(key){
	return this.messages[key][this.getLocale()];
};

var i18n = new I18n(global_messages);

moment().locale(i18n.getLocale());

jQuery(function(){
	
	$("#aLogout").on("click", function(){
		new AjaxUtil({
			url:'/logout',
			dataRaw:{"timestamp" : new Date().getTime()},
			type: "get",
			complete: function(jqXhr){
				if(jqXhr && jqXhr.status == 200){
					document.location = "/";
				}
			}
		}).exec();
	});
	
	$.fn.dialogModal = function(options){
		var settings = $.extend({
			title: '',
			content: '',
			text: false,
			callback: null,
			enableCancel: false,
		}, options || {});
		
		if(settings.enableCancel){
			$("#dlgCancel").show();
			$("#dlgCancel").on("click", function(){
				$(self).modal('close');
			});
		}
		
		var self = this;
		$("#dlgLabel", self).text(settings.title);
				
		if(options.text){
			$("#dlgBody", self).html(settings.content);
		}else if(typeof settings.content == "object"){
			$(".modal-dialog", self).addClass("modal-lg");
			var lines = ['<form class="form-horizontal">'];
			for(var item in settings.content){
				lines.push('<div class="form-group">'+
				      '<label class="col-sm-2 control-label">'+ item +'</label>' +
				      	'<div class="col-sm-10"><p class="form-control-static">' + settings.content[item] + '</p></div>' +
			          '</div>');
			}
			lines.push('</form>');
			$("#dlgBody", self).html(lines.join(""));
		}else{
			$(".modal-dialog", self).removeClass("modal-lg");
			$("#dlgBody", self).text(settings.content);
		}
		
		if(settings.callback != null){
			$("#dlgConfirm").on("click", function(){
			   settings.callback();
			});
		}
		$(self).modal('show');
	}	
});
