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
    .module('harbor.layout.navigation')
    .directive('navigationAdminOptions', navigationAdminOptions);
  
  NavigationAdminOptions.$inject = ['$location'];
  
  function NavigationAdminOptions($location) {
    var vm = this;
    vm.path = $location.path();
  }
  
  function navigationAdminOptions() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/layout/navigation/navigation-admin-options.directive.html',
      'scope': {
        'target': '='
      },
      'link': link,
      'controller': NavigationAdminOptions,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      var visited = ctrl.path.substring(1);  
      console.log('visited:' + visited);

      if(visited) {
        element.find('a[tag="' + visited + '"]').addClass('active');
      }else{
        element.find('a:first').addClass('active');
      }
      
      element.find('a').on('click', click);
            
      function click(event) {
        element.find('a').removeClass('active');
        $(event.target).addClass('active');
        ctrl.target = $(this).attr('tag');
        scope.$apply();
      }
    }
  }
  
})();