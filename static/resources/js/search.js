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
	
	$(document).on("keydown", function(e){
		if(e.keyCode == 13){
			e.preventDefault();
			if($("#txtCommonSearch").is(":focus")){
				search($("#txtCommonSearch").val());
			}
		}
	});
	
	var queryParam = $("#queryParam").val();
	$("#txtCommonSearch").val(queryParam);
	
	search(queryParam);

	function search(keyword){
		keyword = $.trim(keyword)
		if(keyword.length > 0){
			$.ajax({
				url: "/api/search",
				data: {"q": keyword},
				type: "get",
				dataType: "json",
				success: function(data, status, xhr){
					if(xhr && xhr.status == 200){
						$("#panelCommonSearchProjectsHeader").text(i18n.getMessage("projects") + " (" + data.project.length + ")");
						$("#panelCommonSearchRepositoriesHeader").text(i18n.getMessage("repositories") +" (" + data.repository.length + ")");		
						
						render($("#panelCommonSearchProjectsBody"), data.project, "project");
						render($("#panelCommonSearchRepositoriesBody"), data.repository, "repository");
					}
				}
			});
		}
	}	

	var Project = function(id, name, public){
		this.id = id;
		this.name = name;
		this.public = public;
	}
	
	function render(element, data, discriminator){
		$(element).children().remove();
		$.each(data, function(i, e){
			var project, description, repoName;
			switch(discriminator){
			case "project":
				project = new Project(e.id, e.name, e.public);
				description = project.name;
				repoName = "";
				break;
			case "repository":
				project = new Project(e.project_id, e.project_name, e.project_public);
				description = e.repository_name;
				repoName = e.repository_name.substring(e.repository_name.lastIndexOf("/") + 1);
				break;
			} 		
			if(project){
				$(element).append('<div><a href="/registry/detail?project_id=' + project.id + (repoName != "" ? '&repo_name=' + repoName : "") + '">' + description + '</a></div>');
			}			
		});
	}
});