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
(function() {
  
  'use strict';
  
  angular
    .module('harbor.repository')
    .directive('popupDetails', popupDetails);
  
  PopupDetailsController.$inject = [];
  
  function PopupDetailsController() {
       
  }
  
  popupDetails.$inject = ['ListManifestService', '$filter', 'dateLFilter'];
  
  function popupDetails(ListManifestService, $filter, dateLFilter) {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/repository/popup-details.directive.html',
      'scope': {
        'repoName': '@',
        'tag': '@',
        'index': '@'
      },
      'replace': true,
      'link': link,
      'controller': PopupDetailsController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {  
      element
        .popover({
          'template': '<div class="popover" role="tooltip"><div class="arrow"></div><div class="popover-title"></div><div class="popover-content"></div></div>',
          'title': '<div class="pull-right clearfix"><a href="javascript:void(0);"><span class="glyphicon glyphicon-remove-circle"></span></a></div>',
          'content': generateContent,
          'html': true
        })
        .on('show.bs.popover', function(e) {
           var data = ListManifestService(ctrl.repoName, ctrl.tag)
            .success(function(data) {
              return data;
            })
            .error(function(data, status) {
              console.log('Failed to get manifest of repo :' + ctrl.repoName);
              return null;
            });      
           ctrl.manifest = {};
           if(data && angular.isDefined(data.responseJSON)) {
             ctrl.manifest = angular.fromJson(data.responseJSON.manifest.history[0].v1Compatibility);
             ctrl.manifest['created'] = $filter('dateL')(ctrl.manifest['created'], 'YYYY-MM-DD HH:mm:ss');
           }
        })
        .on('inserted.bs.popover', function(e){      
          var self = jQuery(this);                 
          $('[type="text"]:input', self.parent())
            .on('click', function() {
              $(this).select();
            });
          self.parent().find('.glyphicon.glyphicon-remove-circle').on('click', function() {
            element.trigger('click');
          });
        });      
  
      function generateContent() {
        var content =  '<form class="form-horizontal" width="100%">' +
          '<div class="form-group">' +
          '<label class="col-sm-3 control-label">Id</label>' +
          '<div class="col-sm-9"><p class="form-control-static long-line long-line-margin-right"><input type="text" id="txtImageId" value="' + check_output(ctrl.manifest['id']) + '" readonly size="32"></p></div></div>' +
          '<div class="form-group"><label class="col-sm-3 control-label">Parent</label>' +
          '<div class="col-sm-9"><p class="form-control-static long-line long-line-margin-right"><input type="text" id="txtImageId" value="' + check_output(ctrl.manifest['parent']) + '" readonly size="32"></p></div></div>' +
          '<div class="form-group"><label class="col-sm-3 control-label">Created</label>' +
          '<div class="col-sm-9"><p class="form-control-static long-line-margin-right">' + check_output(ctrl.manifest['created']) + '</p></div></div>' +
          '<div class="form-group"><label class="col-sm-3 control-label">Author</label>' +
          '<div class="col-sm-9"><p class="form-control-static long-line-margin-right">' + check_output(ctrl.manifest['author'])  + '</p></div></div>' +
          '<div class="form-group"><label class="col-sm-3 control-label">Architecture</label>' + 
          '<div class="col-sm-9"><p class="form-control-static long-line-margin-right">' + check_output(ctrl.manifest['architecture']) + '</p></div></div>' +
          '<div class="form-group"><label class="col-sm-3 control-label">Docker Version</label>' +
          '<div class="col-sm-9"><p class="form-control-static long-line-margin-right">' + check_output(ctrl.manifest['docker_version']) + '</p></div></div>' +
          '<div class="form-group"><label class="col-sm-3 control-label">OS</label>' +
          '<div class="col-sm-9"><p class="form-control-static long-line-margin-right">' + check_output(ctrl.manifest['os']) + '</p></div></div>' +
        '</form>';
        return content;        
      }
      function check_output(s) {
        return (angular.isUndefined(s) || s === '') ? 'N/A' : s;
      }
    }
  }
  
})();