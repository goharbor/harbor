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
    .module('harbor.search')
    .directive('search', search);
    
  SearchController.$inject = ['SearchService', '$scope'];
  
  function SearchController(SearchService, $scope) {
    var vm = this;
    vm.keywords = "";
    vm.search = searchByFilter;
    vm.filterBy = 'repository';
    
    searchByFilter();
    
    
    function searchByFilter() {
      SearchService(vm.keywords)
        .success(searchSuccess)
        .error(searchFailed);
    }
    
    function searchSuccess(data, status) {
      console.log('filterBy:' + vm.filterBy + ", data:" + data);
      vm.searchResult = data[vm.filterBy];
    }
    
    function searchFailed(data, status) {
      console.log('Failed to search:' + data);
    }
    
  }
  
  function search() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/search/search.directive.html',
      'scope': {
        'filterBy': '='
      },
      'controller': SearchController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    
    return directive; 
  }
  
})();