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
    .module('harbor.log')
    .directive('listLog', listLog);
    
  ListLogController.$inject  = ['$scope','ListLogService', 'getParameterByName', '$location', '$filter', 'trFilter'];
  
  function ListLogController($scope, ListLogService, getParameterByName, $location, $filter, trFilter) {
    
    $scope.subsTabPane = 30;
    
    var vm = this;
    
    vm.sectionHeight = {'min-height': '579px'};
    
    vm.isOpen = false;
       
    vm.beginTimestamp = 0;
    vm.endTimestamp = 0;
    vm.keywords = '';
    
    vm.username = $location.hash() || '';
        
    vm.op = [];
    vm.opOthers = true;
    
    vm.search = search;
    vm.showAdvancedSearch = showAdvancedSearch;
  
    vm.projectId = getParameterByName('project_id', $location.absUrl());
    vm.queryParams = {
      'beginTimestamp' : vm.beginTimestamp,
      'endTimestamp'   : vm.endTimestamp,
      'keywords' : vm.keywords,
      'projectId': vm.projectId,
      'username' : vm.username
    };

    vm.page = 1;
    vm.pageSize = 15;            

    $scope.$watch('vm.page', function(current, origin) {
      if(current) {
        vm.page = current;
        retrieve(vm.queryParams, vm.page, vm.pageSize);
      }
    }); 
      
    $scope.$on('retrieveData', function(e, val) {
      if(val) {
        vm.projectId = getParameterByName('project_id', $location.absUrl());
        vm.queryParams = {
          'beginTimestamp' : vm.beginTimestamp,
          'endTimestamp'   : vm.endTimestamp,
          'keywords' : vm.keywords,
          'projectId': vm.projectId,
          'username' : vm.username
        };
        vm.username = '';
        retrieve(vm.queryParams, vm.page, vm.pageSize);
      }
    });
            
    function search(e) {
      
      vm.page = 1;
      
      if(e.op[0] === 'all') {
        e.op = ['create', 'pull', 'push', 'delete'];
      }      
      if(vm.opOthers && $.trim(vm.others) !== '') {
        e.op.push(vm.others);
      }               
      
      vm.queryParams.keywords = e.op.join('/');
      vm.queryParams.username = e.username;
            
      vm.queryParams.beginTimestamp = toUTCSeconds(vm.fromDate, 0, 0, 0);
      vm.queryParams.endTimestamp = toUTCSeconds(vm.toDate, 23, 59, 59);
      
      retrieve(vm.queryParams, vm.page, vm.pageSize);

    }
    
    function showAdvancedSearch() {
      if(vm.isOpen){
        vm.isOpen = false;
      }else{
        vm.isOpen = true;
      }
    }
    
    function retrieve(queryParams, page, pageSize) {
      ListLogService(queryParams, page, pageSize)
        .then(listLogComplete)
        .catch(listLogFailed);
    }

    function listLogComplete(response) {
      vm.logs = response.data;
      vm.totalCount = response.headers('X-Total-Count');
      
      console.log('Total Count in logs:' + vm.totalCount + ', page:' + vm.page);
      
      vm.isOpen = false;
    }
    function listLogFailed(response){
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_get_log') + response);
      $scope.$emit('raiseError', true);
      console.log('Failed to get log:' + response);
    }
    
    function toUTCSeconds(date, hour, min, sec) {
      if(!angular.isDefined(date) || date === '') {
        return 0;
      }
      
			var t = new Date(date);
			t.setHours(hour);
			t.setMinutes(min);
			t.setSeconds(sec);
			
			return t.getTime() / 1000;
		}
    
  }
  
  function listLog() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/log/list-log.directive.html',
      'scope': {
        'sectionHeight': '='
      },
      'link': link,
      'controller': ListLogController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      element.find('#txtSearchInput').on('keydown', function(e) {
        if($(this).is(':focus') && e.keyCode === 13) {
          ctrl.search({'op': ctrl.op, 'username': ctrl.username});
        }
      });
    }
  }
  
})();