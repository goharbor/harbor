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
    .directive('navigationDetails', navigationDetails);
  
  NavigationDetailsController.$inject = ['$window', '$location', '$scope', 'getParameterByName'];
  
  function NavigationDetailsController($window, $location, $scope, getParameterByName) {
    var vm = this;    
    
     
    vm.projectId = getParameterByName('project_id', $location.absUrl());

    $scope.$on('$locationChangeSuccess', function() {
      vm.projectId = getParameterByName('project_id', $location.absUrl());
    });
    
    vm.path = $location.path();
  }
  
  function navigationDetails() {
    var directive = {
      restrict: 'E',
      templateUrl: '/navigation_detail?timestamp=' + new Date().getTime(),
      link: link,
      scope: {
        'target': '='
      },
      replace: true,
      controller: NavigationDetailsController,
      controllerAs: 'vm',
      bindToController: true
    };
    
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      
      var visited = ctrl.path.substring(1);  
      
      if(visited) {
        element.find('a[tag="' + visited + '"]').addClass('active');
      }else{
        element.find('a:first').addClass('active');
      }
      
      scope.$watch('vm.target', function(current) {
        if(current) {
          ctrl.target = current;
          element.find('a').removeClass('active');
          element.find('a[tag="' + ctrl.target + '"]').addClass('active');
        }
      });
      
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