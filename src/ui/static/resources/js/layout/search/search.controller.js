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
    .module('harbor.layout.search')
    .controller('SearchController', SearchController);
   
  SearchController.$inject = ['$location', 'SearchService', '$scope', '$filter', 'trFilter', 'getParameterByName'];
  
  function SearchController($location, SearchService, $scope, $filter, trFilter, getParameterByName) {
    var vm = this;
    
    vm.q = getParameterByName('q', $location.absUrl());
    console.log('vm.q:' + vm.q);
    SearchService(vm.q)
      .success(searchSuccess)
      .error(searchFailed);
  
    //Error message dialog handler for search.
    $scope.$on('modalTitle', function(e, val) {
      vm.modalTitle = val;
    });
    
    $scope.$on('modalMessage', function(e, val) {
      vm.modalMessage = val;
    });
       
    $scope.$on('raiseError', function(e, val) {
      if(val) {   
        vm.action = function() {
          $scope.$broadcast('showDialog', false);
        };
        vm.contentType = 'text/plain';
        vm.confirmOnly = true;      
        $scope.$broadcast('showDialog', true);
      }
    });
    
    function searchSuccess(data, status) {
      vm.repository = data['repository'];
      vm.project = data['project'];
    }
    
    function searchFailed(data, status) {
      
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_in_search'));
      $scope.$emit('raiseError', true);
      
      console.log('Failed to search:' + data);
    }
  }
  
})();