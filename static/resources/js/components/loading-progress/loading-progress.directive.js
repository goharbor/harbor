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
    .module('harbor.loading.progress')
    .directive('loadingProgress', loadingProgress);
  
  function loadingProgress() {
    var directive = {
      'restrict': 'EA',
      'scope': {
        'toggleInProgress': '=',
        'hideTarget': '@'
      },
      'link': link
    };
    
    return directive;
    
    function link(scope, element, attrs) {
      var spinner = $('<span class="loading-progress">');

      function convertToBoolean(val) {
        return val === 'true' ? true : false;
      }
      
      var hideTarget = convertToBoolean(scope.hideTarget);
      
      console.log('loading-progress, toggleInProgress:' + scope.toggleInProgress + ', hideTarget:' + hideTarget);
      
      var pristine = element.html();
                 
      scope.$watch('toggleInProgress', function(current) {
        if(scope.toggleInProgress) {
          element.attr('disabled', 'disabled');
          if(hideTarget) {
            element.html(spinner);
          }else{
            spinner = spinner.css({'margin-left': '5px'});
            element.append(spinner);
          }
        }else{
          if(hideTarget) {
            element.html(pristine);
          }else{
            element.find('.loading-progress').remove();
          }
          element.removeAttr('disabled');
        }
      });

    }
  }
  
})();