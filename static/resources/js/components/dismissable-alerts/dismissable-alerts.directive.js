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
    .module('harbor.dismissable.alerts')
    .directive('dismissableAlerts', dismissableAlerts);

  function dismissableAlerts() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/dismissable-alerts/dismissable-alerts.directive.html',
      'link': link
    };
    return directive;
    function link(scope, element, attrs, ctrl) {
      
      scope.close = function() {
        scope.toggleAlert = false;
      };
      scope.$on('raiseAlert', function(e, val) {
        console.log('received raiseAlert:' + angular.toJson(val));
        if(val.show) {
          scope.message = val.message;
          scope.toggleAlert = true;
        }else{
          scope.message = '';
          scope.toggleAlert = false;
        }
      });
    }
  }
  
})();
