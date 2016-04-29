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
				if(data.HasAdminRole == 1) {
					renderForAdminRole();
				}
				renderForAnyRole();
			}
		}
	}).exec();
	
	function renderForAnyRole(){
		$("#tabProject a:first").tab("show");
	
		$(document).on("keydown", function(e){
			if(e.keyCode == 13){
			  e.preventDefault();
			  if($("#tabProject li:eq(0)").is(":focus") || $("#txtSearchProject").is(":focus")){
			    $("#btnSearch").trigger("click");	
			  }else if($("#tabProject li:eq(1)").is(":focus") || $("#txtSearchPublicProjects").is(":focus")){
			    $("#btnSearchPublicProjects").trigger("click");	
			  }else if($("#tabProject li:eq(1)").is(":focus") || $("#txtSearchUsername").is(":focus")){
			    $("#btnSearchUsername").trigger("click");	
			  }else if($("#dlgAddProject").is(":focus") || $("#projectName").is(":focus")){
				$("#btnSave").trigger("click");
			  }
			}
		});
	
		function listProject(projectName, isPublic){
			currentPublic = isPublic;
			$.when(
				new AjaxUtil({
				  url: "/api/projects?is_public=" + isPublic + "&project_name=" + (projectName == null ? "" : projectName) + "&timestamp=" + new Date().getTime(),
				  type: "get",
				  success: function(data, status, xhr){
		              $("#tblProject tbody tr").remove();
		              $.each(data || [], function(i, e){
		                  var row = '<tr>' +
		                  '<td style="vertical-align: middle;"><a href="/registry/detail?project_id=' + e.ProjectId + '">' + e.Name + '</a></td>' +
		                  '<td style="vertical-align: middle;">' + moment(new Date(e.CreationTime)).format("YYYY-MM-DD HH:mm:ss") + '</td>';
		                  if(e.Public == 1 && e.Togglable){
		                      row += '<td><button type="button" class="btn btn-success" projectid="' + e.ProjectId + '">' + i18n.getMessage("button_on")+ '</button></td>'
		                  } else if (e.Public == 1) {
		                      row += '<td><button type="button" class="btn btn-success" projectid="' + e.ProjectId + '" disabled>' + i18n.getMessage("button_on")+ '</button></td>';
		                  } else if (e.Public == 0 && e.Togglable) {
		                      row += '<td><button type="button" class="btn btn-danger" projectid="' + e.ProjectId + '">' + i18n.getMessage("button_off")+ '</button></td>';
		                  } else if (e.Public == 0) {
		                      row += '<td><button type="button" class="btn btn-danger" projectid="' + e.ProjectId + '" disabled>' + i18n.getMessage("button_off")+ '</button></td>';
		                      row += '</tr>';
		                  }
		                  $("#tblProject tbody").append(row);
		              });
		          }
			}).exec())
			.done(function() {
                $("#tblProject tbody tr :button").on("click", function(){
                    var projectId = $(this).attr("projectid");
                    var self = this;
					 new AjaxUtil({
					   url: "/api/projects/" + projectId, 
					   data: {"public": ($(self).hasClass("btn-success") ? false : true)},
					   type: "put",
					   complete: function(jqXhr, status) {
							if($(self).hasClass("btn-success")){
								$(self).removeClass("btn-success").addClass("btn-danger");
								$(self).html(i18n.getMessage("button_off"));
							}else{
								$(self).removeClass("btn-danger").addClass("btn-success");
								$(self).html(i18n.getMessage("button_on"));
							}
						}
					 }).exec();
                });
            });
		}	
		listProject(null, 0);
		var currentPublic = 0;
		
		$("#tabProject a:eq(0)").on("click", function(){
			$("#btnAddProject").css({"visibility": "visible"});
			listProject(null, 0);		
		});
		
		$("#tabProject a:eq(1)").on("click", function(){
			$("#btnAddProject").css({"visibility": "hidden"});
			listProject(null, 1);
		});
		
		$("#divErrMsg").css({"display": "none"});
		validateOptions.Items.push("#projectName");
		
		$('#dlgAddProject').on('hide.bs.modal', function () {
			$("#divErrMsg").css({"display": "none"});
			$("#projectName").val("");
			$("#projectName").parent().removeClass("has-feedback").removeClass("has-error").removeClass("has-success");
			$("#projectName").siblings("span").removeClass("glyphicon-warning-sign").removeClass("glyphicon-ok");
		});
		$('#dlgAddProject').on('show.bs.modal', function () {
			$("#divErrMsg").css({"display": "none"});
			$("#projectName").val("");
			$("#projectName").parent().addClass("has-feedback");
			$("#projectName").siblings("span").removeClass("glyphicon-warning-sign").removeClass("glyphicon-ok");
	        $("#isPublic").prop('checked', false);
		});
		
		$("#btnSave").on("click", function(){	
			validateOptions.Validate(function() {
				new AjaxUtil({
					url: "/api/projects",
					data: {"project_name" : $.trim($("#projectName").val()), "public":$("#isPublic").prop('checked'), "timestamp" : new Date().getTime()},
					type: "post",
					errors: {
						409: i18n.getMessage("project_exists")
					},
					complete: function(jqXhr, status){
						$("#btnCancel").trigger("click");
						listProject(null, currentPublic);
					}
				}).exec();
			});
		});
		
		$("#btnSearch").on("click", function(){
			var projectName = $("#txtSearchProject").val();
			if($.trim(projectName).length == 0){
				projectName = null;
			}
			listProject(projectName, currentPublic);	
		});
	}
	
	
	function renderForAdminRole(){
		$("#liAdminOption").css({"visibility": "visible"});
		$("#tabAdminOption").css({"visibility": "visible"});	
		function listUserAdminRole(searchUsername){
			$.when(
				new AjaxUtil({
					url: "/api/users?username=" + (searchUsername == null ? "" : searchUsername),
					type: "get",
					success: function(data){
						$("#tblUser tbody tr").remove();
						$.each(data || [], function(i, e){
							var row = '<tr>' +
								'<td style="vertical-align: middle;">' + e.username + '</td>' +
								'<td style="vertical-align: middle;">' + e.email + '</td>';
							if(e.HasAdminRole == 1){
								row += '<td style="padding-left: 30px;"><button type="button" class="btn btn-success" userid="' + e.UserId + '">' + i18n.getMessage("button_on") + '</button></td>';
							} else {
								row += '<td style="padding-left: 30px;"><button type="button" class="btn btn-danger" userid="' + e.UserId + '">' + i18n.getMessage("button_off") + '</button></td>';
							}
							row += '<td style="padding-left: 30px; vertical-align: middle;"><a href="#" style="visibility: hidden;" class="tdDeleteUser" userid="' + e.UserId + '" username="' + e.Username + '"><span class="glyphicon glyphicon-trash"></span></a></td>';
							row += '</tr>';
							$("#tblUser tbody").append(row);
						});
					}
				}).exec()
			).done(function(){
				$("#tblUser tbody tr :button").on("click",function(){
					var userId = $(this).attr("userid");
					var self = this;
					new AjaxUtil({
						url: "/api/users/" + userId,
						type: "put",
						complete: function(jqXhr, status){
							if(jqXhr && jqXhr.status == 200){
								if($(self).hasClass("btn-success")){
									$(self).removeClass("btn-success").addClass("btn-danger");
									$(self).html(i18n.getMessage("button_off"));
								}else{
									$(self).removeClass("btn-danger").addClass("btn-success");
									$(self).html(i18n.getMessage("button_on"));
								}
							}		
						}
					}).exec();
				});
				$("#tblUser tbody tr").on("mouseover", function(){
					$(".tdDeleteUser", this).css({"visibility":"visible"});
				}).on("mouseout", function(){
					$(".tdDeleteUser", this).css({"visibility":"hidden"});
				});
				$("#tblUser tbody tr .tdDeleteUser").on("click", function(){
					var userId = $(this).attr("userid");
					$("#dlgModal")
						.dialogModal({
							"title": i18n.getMessage("delete_user"), 
							"content": i18n.getMessage("are_you_sure_to_delete_user") + $(this).attr("username") + " ?", 
							"enableCancel": true, 
							"callback": function(){
								new AjaxUtil({
									url: "/api/users/" + userId,
									type: "delete",
									complete: function(jqXhr, status){
										if(jqXhr && jqXhr.status == 200){
											$("#btnSearchUsername").trigger("click");
										}
									},
									error: function(jqXhr){}
								}).exec();
							}	
					});
				});
			});
		}
	    listUserAdminRole(null);
		$("#btnSearchUsername").on("click", function(){
			var username = $("#txtSearchUsername").val();
			if($.trim(username).length == 0){
				username = null;
			}
			listUserAdminRole(username);
		});
	}
})
