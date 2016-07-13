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
    .directive('pullCommand', pullCommand);
  
  function PullCommandController() {
    
  }
  
  function pullCommand() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/repository/pull-command.directive.html',
      'scope': {
        'repoName': '@',
        'tag': '@'
      },
      'link': link,
      'controller': PullCommandController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
       
      ctrl.harborRegUrl = $('#HarborRegUrl').val() + '/';
      
      element.find('input[type="text"]').on('click', function() {
        $(this).select();
      });
      
      element.find('a').on('click', clickHandler);
      
      function clickHandler(e) {
        element.find('input[type="text"]').select();
      }
  
    }
    
  }
  
})();