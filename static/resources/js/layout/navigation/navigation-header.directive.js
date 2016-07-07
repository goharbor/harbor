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
    .directive('navigationHeader', navigationHeader);
  
  NavigationHeaderController.$inject = ['$window', '$scope', 'currentUser', '$timeout'];
    
  function NavigationHeaderController($window, $scope, currentUser, $timeout) {
    var vm = this;
    vm.url = $window.location.pathname;    
  }
  
  function navigationHeader() {
    var directive = {
      restrict: 'E',
      templateUrl: '/navigation_header?timestamp=' + new Date().getTime(),
      link: link,
      scope: true,
      controller: NavigationHeaderController,
      controllerAs: 'vm',
      bindToController: true
    };
    
    return directive;
   
    function link(scope, element, attrs, ctrl) {     
      var visited = ctrl.url;
      console.log('visited:' + visited);
      if (visited !== '' && visited !== '/') {
         element.find('a[href*="' + visited + '"]').addClass('active'); 
      }      
      element.find('a').on('click', click);
      function click(event) {
        element.find('a').removeClass('active');
        $(event.target).not('span').addClass('active');
      }     
    }
   
  }
  
})();