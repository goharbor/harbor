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
	
	$.when(
		new AjaxUtil({
			url: "/api/users/current",
			type: "get",
			error: function(jqXhr){
				if(jqXhr){				
					if(jqXhr.status == 403){
						return false;
					}
			    }
			}
		}).exec()
	).then(function(){
		noNeedToLoginCallback();
		needToLoginCallback();		
	}).fail(function(){
		noNeedToLoginCallback();
    });
	
	function noNeedToLoginCallback(){
		
		$("#tabItemDetail a:first").tab("show");
		$("#btnFilterOption button:first").addClass("active");
		$("#divErrMsg").hide();
	
		if($("#public").val() == 1){
			$("#tabItemDetail li:eq(1)").hide();
			$("#tabItemDetail li:eq(2)").hide();
		}
		
		listRepo($("#repoName").val());	
		
		function listRepo(repoName){
		
			$("#divErrMsg").hide();

			new AjaxUtil({
			  url: "/api/repositories?project_id=" + $("#projectId").val() + "&q=" + repoName,
			  type: "get",
			  success: function(data, status, xhr){
				if(xhr && xhr.status == 200){
						$("#accordionRepo").children().remove();
						if(data == null){
							$("#divErrMsg").show();
							$("#divErrMsg center").html(i18n.getMessage("no_repo_exists"));
							return;
						}
						$.each(data, function(i, e){						
							var targetId = e.replace(/\//g, "------");						
							var row = '<div class="panel panel-default"  targetId="' + targetId + '">' +
								'<div class="panel-heading" role="tab" id="heading' + i + '"+ >' + 
							      '<h4 class="panel-title">' +
							        '<a data-toggle="collapse" data-parent="#accordion" href="#collapse'+ i + '" aria-expanded="true" aria-controls="collapse' + i + '">' +
										'<span class="list-group-item-heading"> <span class="glyphicon glyphicon-book blue"></span> ' + e + ' </span>' +        	
							        '</a>' +
							      '</h4>' +
							    '</div>' +
							    '<div id="collapse' + i + '" targetId="' + targetId + '"  class="panel-collapse collapse" role="tabpanel" aria-labelledby="heading' + i + '">' +
								    '<div class="panel-body" id="' + targetId + '">' +
								        '<div class="table-responsive" style="height: auto;">' + 
										  '<table class="table table-striped table-bordered table-condensed">' +
											'<thead>' +
												'<tr>' +
												  '<th class="st-sort-ascent" st-sort="name" st-sort-default=""><span class="glyphicon glyphicon-tag blue"></span> ' + i18n.getMessage("tag")+ ' </th>' +
												  '<th class="st-sort-ascent" st-sort="name" st-sort-default=""><span class="glyphicon glyphicon-tag blue"></span> ' + i18n.getMessage("pull_command") + ' </th>' +
												'</tr>' +
											'</thead>' +
											'<tbody>'  +
											'</tbody>' + 
										  '</table>'
									    '</div>' +
								    '</div>' +
							    '</div>' +
							'</div>';					
							$("#accordionRepo").append(row);
						});
						if(repoName != ""){
							$("#txtRepoName").val(repoName);
							$("#accordionRepo #heading0 a").trigger("click");
						}
					}
				}
			}).exec();
		}
		$("#btnSearchRepo").on("click", function(){
			listRepo($.trim($("#txtRepoName").val()));
		});
		
		$('#accordionRepo').on('show.bs.collapse', function (e) {
			$('#accordionRepo .in').collapse('hide');
			var targetId = $(e.target).attr("targetId");
			var repoName = targetId.replace(/------/g, "/");
			new AjaxUtil({
			  url: "/api/repositories/tags?repo_name=" + repoName,
			  type: "get",
			  success: function(data, status, xhr){
					$('#' + targetId +' table tbody tr').remove();
					var row = [];
					for(var i in data){
						var tagName = data[i]
						row.push('<tr><td><a href="#" imageId="' + tagName + '" repoName="' + repoName + '">' + tagName + '</a></td><td><input type="text" style="width:100%" readonly value="  docker pull '+ $("#harborRegUrl").val() +'/'+ repoName + ':' + tagName +'"></td></tr>');							
					}
					$('#' + targetId +' table tbody').append(row.join(""));
					$('#' + targetId +' table tbody tr a').on("click", function(e){
						var imageId = $(this).attr("imageId");
						var repoName = $(this).attr("repoName");
						new AjaxUtil({
						  url: "/api/repositories/manifests?repo_name=" + repoName + "&tag=" + imageId,
						  type: "get",
						  success: function(data, status, xhr){
							  if(data){	
							     for(var i in data){
									if(data[i] == ""){
										data[i] = "N/A";
									}
								 }								
								 data.Created = moment(new Date(data.Created)).format("YYYY-MM-DD HH:mm:ss");
							 
							     $("#dlgModal").dialogModal({"title": i18n.getMessage("image_details"), "content": data});		
							  }
						  }
						}).exec();
					});
				}
			}).exec();
		});
	}
	
	function needToLoginCallback(){
		
		var hasAuthorization = false;
		
		$.when(
			new AjaxUtil({
		  		url: "/api/projects/" + $("#projectId").val() + "/members/current",
		  		type: "get",
			    success: function(data, status, xhr){
					if(xhr && xhr.status == 200 && data.roles != null && data.roles.length > 0){
						hasAuthorization = true;
					}
		  		}
			}).exec())
		.done(function(){
			
			if(!hasAuthorization) return false;
			
			$("#tabItemDetail a:eq(1)").css({"visibility": "visible"});
			$("#tabItemDetail a:eq(2)").css({"visibility": "visible"});

			$(".glyphicon .glyphicon-pencil", "#tblUser").on("click", function(e){	
				$("#txtUserName").hide();
				$("#lblUserName").show();
				$("#dlgUserTitle").text(i18n.getMessage("edit_members"));
			});
		
			$("#btnAddUser").on("click", function(){		
				$("#operationType").val("add");
				$("#spnSearch").show();
				$("#txtUserName").prop("disabled", false)
				$("#txtUserName").val("");
				$("#lstRole input[name=chooseRole]:radio").prop("checked", false);
				$("#dlgUserTitle").text(i18n.getMessage("add_members"));
			});
		
			$("#btnSave").on("click", function(){
			
				var username = $("#txtUserName").val();
				if($.trim(username).length == 0){
					$("#dlgModal").dialogModal({"title": i18n.getMessage("add_member_failed"), "content": i18n.getMessage("please_input_username")});
					return;
				}
				var projectId = $("#projectId").val();	
				var operationType = $("#operationType").val();
				var userId = $("#editUserId").val();
				
				var checkedRole = $("#lstRole input[name='chooseRole']:checked")			
				if(checkedRole.length == 0){
					$("#dlgModal").dialogModal({"title": i18n.getMessage("add_member_failed"), "content": i18n.getMessage("please_assign_a_role_to_user")});
					return;
				}
			
				var checkedRoleItemList = [];
				$.each(checkedRole, function(i, e){
					checkedRoleItemList.push(new Number($(this).val()));
				});		
			
				var ajaxOpts = {};
				if(operationType == "add"){
					ajaxOpts.url = "/api/projects/" + projectId + "/members/";
					ajaxOpts.type = "post";
					ajaxOpts.data = {"roles" : checkedRoleItemList, "user_name": username};
				}else if(operationType == "edit"){
					ajaxOpts.url = "/api/projects/" + projectId + "/members/" + userId;
					ajaxOpts.type = "put";
					ajaxOpts.data = {"roles" : checkedRoleItemList};
				}
		
				new AjaxUtil({
				  url: ajaxOpts.url,
				  data: ajaxOpts.data,
				  type: ajaxOpts.type,
				  complete: function(jqXhr, status){
					  if(jqXhr && jqXhr.status == 200){
					  	  $("#btnCancel").trigger("click");
					      listUser(null);
					  }
				  },
				  errors: {
					404: i18n.getMessage("user_id_does_not_exist"),
					409: i18n.getMessage("user_id_exists"),
					403: i18n.getMessage("insufficient_privileges")
				  }
				}).exec();			
			});
		
			var name_mapping = {
				"projectAdmin": "Project Admin",
				"developer": "Developer",
				"guest": "Guest"
			}
		
			function listUserByProjectCallback(userList){
				var loginedUserId = $("#userId").val();
				var loginedUserRoleId = $("#roleId").val();
				var ownerId = $("#ownerId").val();
		
				$("#tblUser tbody tr").remove();
				for(var i = 0; i < userList.length; ){
				
					var userId = userList[i].UserId;
					var roleId = userList[i].RoleId;
					var username = userList[i].username;
					var roleNameList = [];
				
					for(var j = i; j < userList.length; i++, j++){
						if(userList[j].UserId == userId){
							roleNameList.push(name_mapping[userList[j].Rolename]);					
						}else{
							break;
						}
					}
							
					var row = '<tr>' + 
					'<td>' + username + '</td>' + 
					'<td>' + roleNameList.join(",") + '</td>' +
					'<td>';
					var isShowOperations = true;
					if(loginedUserRoleId >= 3 /*role: developer guest*/){
						isShowOperations = false;
					}else if(ownerId == userId){
					    isShowOperations = false;
				    }else if (loginedUserId == userId){
						isShowOperations = false;
				    }
					if(isShowOperations){
						row += '<a href="#" userid="' + userId + '" class="glyphicon glyphicon-pencil" data-toggle="modal" data-target="#dlgUser"></a>&nbsp;' +
						'<a href="#" userid="' + userId + '" roleid="' + roleId + '" class="glyphicon glyphicon-trash"></a>';
					}
				
					row += '</td></tr>';
					$("#tblUser tbody").append(row);
				
				}
			}
	
			function searchAccessLogCallback(LogList){
				$("#tabOperationLog tbody tr").remove();
				$.each(LogList || [], function(i, e){
					$("#tabOperationLog tbody").append(
						'<tr>' + 
						'<td>' + e.Username + '</td>' + 
						'<td>' + e.RepoName + '</td>' +
						'<td>' + e.Operation + '</td>' +
						'<td>' + moment(new Date(e.OpTime)).format("YYYY-MM-DD HH:mm:ss") + '</td>' +
						'</tr>');
				});
			}
	
			function getUserRoleCallback(userId){	
				new AjaxUtil({
				  url: "/api/projects/" + $("#projectId").val() + "/members/" + userId,
				  type: "get",
				  success: function(data, status, xhr){
					  var user = data;
					  $("#operationType").val("edit");
					  $("#editUserId").val(user.user_id);
					  $("#spnSearch").hide();
					  $("#txtUserName").val(user.user_name);
					  $("#txtUserName").prop("disabled", true);	
					  $("#btnSave").removeClass("disabled");				
					  $("#dlgUserTitle").text(i18n.getMessage("edit_members"));					
					  $("#lstRole input[name=chooseRole]:radio").not('[value=' + user.role_id + ']').prop("checked", false)
					  $.each(user.roles, function(i, e){
						  $("#lstRole input[name=chooseRole]:radio").filter('[value=' + e.role_id + ']').prop("checked", "checked");									
					  });
				  }
				}).exec();
			}
			function listUser(username){
				$.when(
					new AjaxUtil({
					  url: "/api/projects/" + $("#projectId").val() + "/members?username=" + (username == null ? "" : username),
					  type: "get",
					  errors: {
						403: ""
					  },
					  success: function(data, status, xhr){
					      return data || [];
					  }
					}).exec()
				).done(function(userList){
					listUserByProjectCallback(userList || []);
					$("#tblUser .glyphicon-pencil").on("click", function(e){
						var userId = $(this).attr("userid")
						getUserRoleCallback(userId);				
					});
					$("#tblUser .glyphicon-trash").on("click", function(){
						var userId = $(this).attr("userid");
						new AjaxUtil({
						  url: "/api/projects/" + $("#projectId").val() + "/members/" + userId,
						  type: "delete",
						  complete: function(jqXhr, status){
							  if(jqXhr && jqXhr.status == 200){
								  listUser(null);
							  }
						  }
						}).exec();
					});
				});
			}
			listUser(null);
			listOperationLogs();
		
			function listOperationLogs(){
				var projectId = $("#projectId").val();	
				
				$.when(
					new AjaxUtil({
						url : "/api/projects/" + projectId + "/logs/filter",
						data: {},
						type: "post",
						success: function(data){
							return data || [];
						}
					}).exec()
				).done(function(operationLogs){
					searchAccessLogCallback(operationLogs);
				});
			}
			
			$("#btnSearchUser").on("click", function(){
				var username = $("#txtSearchUser").val();
				if($.trim(username).length == 0){
					username = null;
				}
				listUser(username);
			});
			
			function toUTCSeconds(date, hour, min, sec) {
				var t = new Date(date);
				t.setHours(hour);
				t.setMinutes(min);
				t.setSeconds(sec);
				var utcTime = new Date(t.getUTCFullYear(),
								t.getUTCMonth(), 
								t.getUTCDate(),
								t.getUTCHours(),
								t.getUTCMinutes(),
								t.getUTCSeconds());
				return utcTime.getTime() / 1000;
			}
			
			$("#btnFilterLog").on("click", function(){

				var projectId = $("#projectId").val();	
				var username = $("#txtSearchUserName").val();
				
				var beginTimestamp = 0;
				var endTimestamp = 0;
				
				if($("#begindatepicker").val() != ""){
					beginTimestamp = toUTCSeconds($("#begindatepicker").val(), 0, 0, 0);	
				}
				if($("#enddatepicker").val() != ""){
					endTimestamp = toUTCSeconds($("#enddatepicker").val(), 23, 59, 59);	
				}
			
				new AjaxUtil({
					url: "/api/projects/" + projectId + "/logs/filter",
					data:{"username":username, "project_id" : projectId, "keywords" : getKeyWords() , "beginTimestamp" : beginTimestamp, "endTimestamp" : endTimestamp},
					type: "post",
					success: function(data, status, xhr){
						if(xhr && xhr.status == 200){
						  searchAccessLogCallback(data);
						}
					}
				}).exec();
			});
		
			$("#spnFilterOption input[name=chkAll]").on("click", function(){
				$("#spnFilterOption input[name=chkOperation]").prop("checked", $(this).prop("checked"));
			});
		
		    $("#spnFilterOption input[name=chkOperation]").on("click", function(){
				if(!$(this).prop("checked")){
					$("#spnFilterOption input[name=chkAll]").prop("checked", false);
				}
				
				var selectedAll = true;
	
				$("#spnFilterOption input[name=chkOperation]").each(function(i, e){
					if(!$(e).prop("checked")){
						selectedAll = false;
					}
				});
				
				if(selectedAll){
					$("#spnFilterOption input[name=chkAll]").prop("checked", true);
				}
			});
		
			function getKeyWords(){
				var keywords = "";
				var checkedItemList=$("#spnFilterOption input[name=chkOperation]:checked");
		        var keywords = [];
				$.each(checkedItemList, function(i, e){
					var itemValue = $(e).val();
					if(itemValue == "others" && $.trim($("#txtOthers").val()).length > 0){
						keywords.push($("#txtOthers").val());
					}else{
						keywords.push($(e).val());
					}
				});
				return keywords.join("/");
			}
				
			$('#datetimepicker1').datetimepicker({
				locale: i18n.getLocale(),
				ignoreReadonly: true,
				format: 'L',
				showClear: true
		    });
			$('#datetimepicker2').datetimepicker({
				locale: i18n.getLocale(),
				ignoreReadonly: true,
				format: 'L',
				showClear: true
		    });
		});
	}
	
	$(document).on("keydown", function(e){
		if(e.keyCode == 13){
		  e.preventDefault();
		  if($("#tabItemDetail li:eq(0)").is(":focus") || $("#txtRepoName").is(":focus")){
		    $("#btnSearchRepo").trigger("click");	
		  }else if($("#tabItemDetail li:eq(1)").is(":focus") || $("#txtSearchUser").is(":focus")){
		    $("#btnSearchUser").trigger("click");	
		  }else if($("#tabItemDetail li:eq(2)").is(":focus") || $("#txtSearchUserName").is(":focus")){
		    $("#btnFilterLog").trigger("click");	
		  }else if($("#txtUserName").is(":focus") || $("#lstRole :radio").is(":focus")){
		    $("#btnSave").trigger("click");	
		  }
		}
	});
})
