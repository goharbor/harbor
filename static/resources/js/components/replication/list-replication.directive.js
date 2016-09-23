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
    .module('harbor.replication')
    .directive('listReplication', listReplication)
    .factory('jobStatus', jobStatus);

  jobStatus.inject = ['$filter', 'trFilter'];
  function jobStatus($filter, trFilter) {
    return function() {      
      return [
        {'key': 'all'    , 'value': $filter('tr')('all')},
        {'key': 'pending', 'value': $filter('tr')('pending')},
        {'key': 'running', 'value': $filter('tr')('running')},
        {'key': 'error'  , 'value': $filter('tr')('error')},
        {'key': 'retrying', 'value': $filter('tr')('retrying')},
        {'key': 'stopped', 'value': $filter('tr')('stopped')}, 
        {'key': 'finished', 'value':$filter('tr')('finished')},
        {'key': 'canceled', 'value': $filter('tr')('canceled')}
      ];
    };
  }
  
  ListReplicationController.$inject = ['$scope', 'getParameterByName', '$location', 'ListReplicationPolicyService', 'ToggleReplicationPolicyService', 'DeleteReplicationPolicyService', 'ListReplicationJobService', '$window', '$filter', 'trFilter', 'jobStatus'];
  
  function ListReplicationController($scope, getParameterByName, $location, ListReplicationPolicyService, ToggleReplicationPolicyService, DeleteReplicationPolicyService, ListReplicationJobService, $window, $filter, trFilter, jobStatus) {
    var vm = this;
    
    vm.sectionHeight = {'min-height': '1260px'};
      
    $scope.$on('retrieveData', function(e, val) {
      if(val) {
        vm.projectId = getParameterByName('project_id', $location.absUrl());
        vm.retrievePolicy();
      }
    });
    
    vm.addReplication = addReplication;
    vm.editReplication = editReplication;
    vm.deleteReplicationPolicy = deleteReplicationPolicy;
    
    vm.confirmToDelete = confirmToDelete;
    
    vm.searchReplicationPolicy = searchReplicationPolicy;
    vm.searchReplicationJob = searchReplicationJob;
    vm.refreshReplicationJob = refreshReplicationJob;
    
    vm.retrievePolicy = retrievePolicy;
    vm.retrieveJob = retrieveJob;
    
    vm.pageSize = 20;
    vm.page = 1;
    
    $scope.$watch('vm.page', function(current) {
      if(vm.lastPolicyId !== -1 && current) {
        vm.page = current;
        console.log('replication job: vm.page:' + current);
        vm.retrieveJob(vm.lastPolicyId, vm.page, vm.pageSize);
      }
    });
    
    vm.confirmToTogglePolicy = confirmToTogglePolicy;
    vm.togglePolicy = togglePolicy;
    
    vm.downloadLog = downloadLog;
      
    vm.last = false;
    
    vm.projectId = getParameterByName('project_id', $location.absUrl());
    vm.retrievePolicy();
       
    vm.jobStatus = jobStatus;
    vm.currentStatus = vm.jobStatus()[0];
   
    vm.pickUp = pickUp;
    
    vm.searchJobTIP = false;
    vm.refreshJobTIP = false;
        
    function searchReplicationPolicy() {
      vm.retrievePolicy();
    }   
    
    function searchReplicationJob() {
      if(vm.lastPolicyId !== -1) {
        vm.searchJobTIP = true;
        vm.retrieveJob(vm.lastPolicyId, vm.page, vm.pageSize);
      }
    }            
    
    function refreshReplicationJob() {
      if(vm.lastPolicyId !== -1) {
        vm.refreshJobTIP = true;
        vm.retrieveJob(vm.lastPolicyId, vm.page, vm.pageSize);
      }
    }
   
    function retrievePolicy() {
      ListReplicationPolicyService('', vm.projectId, vm.replicationPolicyName)
        .success(listReplicationPolicySuccess)
        .error(listReplicationPolicyFailed);
    }
    
    function retrieveJob(policyId, page, pageSize) {
      var status = (vm.currentStatus.key === 'all' ? '' : vm.currentStatus.key);
      ListReplicationJobService(policyId, vm.replicationJobName, status, toUTCSeconds(vm.fromDate, 0, 0, 0), toUTCSeconds(vm.toDate, 23, 59, 59), page, pageSize)
        .then(listReplicationJobSuccess, listReplicationJobFailed);
    }

    function listReplicationPolicySuccess(data, status) {
      vm.replicationJobs = [];
      vm.replicationPolicies = data || [];
    }
    
    function listReplicationPolicyFailed(data, status) {
      console.log('Failed to list replication policy:' + data);      
    }

    function listReplicationJobSuccess(response) {
      vm.replicationJobs = response.data || [];
      vm.totalCount = response.headers('X-Total-Count');
      var alertInfo = {
        'show': false,
        'message': ''
      };
      angular.forEach(vm.replicationJobs, function(item) {
        for(var key in item) {          
          var value = item[key];
          if(key === 'status' && (value === 'error' || value === 'retrying')) {
            alertInfo.show = true;
            alertInfo.message = $filter('tr')('alert_job_contains_error');
          }
          switch(key) {
          case 'operation':            
          case 'status':
            item[key] = $filter('tr')(value);
            break;
          default:
            break;
          }
        }
      });
     
      $scope.$emit('raiseAlert', alertInfo);
      vm.searchJobTIP = false;
      vm.refreshJobTIP = false;
    }
    
    function listReplicationJobFailed(response) {
      console.log('Failed to list replication job:' + response);
      vm.searchJobTIP = false;
      vm.refreshJobTIP = false;
    }

    function addReplication() {
      vm.modalTitle = $filter('tr')('add_new_policy', []);
      vm.action = 'ADD_NEW';
    }
    
    function editReplication(policyId) {
      vm.policyId = policyId;
      vm.modalTitle = $filter('tr')('edit_policy', []);
      vm.action = 'EDIT';
      
      console.log('Selected policy ID:' + vm.policyId);
    }

    function deleteReplicationPolicy() {
      DeleteReplicationPolicyService(vm.policyId)
        .success(deleteReplicationPolicySuccess)
        .error(deleteReplicationPolicyFailed);
    }
    
    function deleteReplicationPolicySuccess(data, status) {
      console.log('Successful delete replication policy.');
      vm.retrievePolicy();
    }
    
    function deleteReplicationPolicyFailed(data, status) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      if(status === 412) {
        $scope.$emit('modalMessage', $filter('tr')('failed_to_delete_replication_enabled')); 
      }else{
        $scope.$emit('modalMessage', $filter('tr')('failed_to_delete_replication_policy'));
      }
      $scope.$emit('raiseError', true);
      console.log('Failed to delete replication policy.');
    }
    
    function confirmToDelete(policyId, policyName) {
      vm.policyId = policyId;
     
      $scope.$emit('modalTitle', $filter('tr')('confirm_delete_policy_title'));
      $scope.$emit('modalMessage', $filter('tr')('confirm_delete_policy', [policyName]));
      
      var emitInfo = {
        'confirmOnly': false,
        'contentType': 'text/plain',
        'action': vm.deleteReplicationPolicy
      };
      
      $scope.$emit('raiseInfo', emitInfo);
    }


    function confirmToTogglePolicy(policyId, enabled, name) {
      vm.policyId = policyId;
      vm.enabled = enabled;

      var status = $filter('tr')(vm.enabled === 1 ? 'enable':'disable');

      var title;
      var message;
      if(enabled === 1){
        title = $filter('tr')('confirm_to_toggle_enabled_policy_title');
        message = $filter('tr')('confirm_to_toggle_enabled_policy');
      }else{
        title = $filter('tr')('confirm_to_toggle_disabled_policy_title');
        message = $filter('tr')('confirm_to_toggle_disabled_policy');
      }
      $scope.$emit('modalTitle', title);
      $scope.$emit('modalMessage', message);
            
      var emitInfo = {
        'contentType': 'text/html',
        'confirmOnly': false,
        'action': vm.togglePolicy
      };
      
      $scope.$emit('raiseInfo', emitInfo);
    }
     
    function togglePolicy() {      
      ToggleReplicationPolicyService(vm.policyId, vm.enabled)
        .success(toggleReplicationPolicySuccess)
        .error(toggleReplicationPolicyFailed);
    }
    
    function toggleReplicationPolicySuccess(data, status) {
      console.log('Successful toggle replication policy.');
      vm.retrievePolicy();
    }
    
    function toggleReplicationPolicyFailed(data, status) {
      console.log('Failed to toggle replication policy.');
    }
    
    function downloadLog(policyId) {
      $window.open('/api/jobs/replication/' + policyId + '/log', '_blank');
    }
    
    function pickUp(e) {
      switch(e.key){
      case 'fromDate':
        vm.fromDate = e.value;  
        break;
      case 'toDate':
        vm.toDate = e.value;
        break;
      }
      $scope.$apply();
    }
    
    function toUTCSeconds(date, hour, min, sec) {
      if(!angular.isDefined(date) || date === '') {
        return '';
      }
			var t = new Date(date);
			t.setHours(hour);
			t.setMinutes(min);
			t.setSeconds(sec);
			return t.getTime() / 1000;
		}
  
  }
  
  function listReplication($timeout, I18nService) {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/replication/list-replication.directive.html',
      'scope': {
        'sectionHeight': '='
      },
      'link': link,
      'controller': ListReplicationController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {


      var uponPaneHeight = element.find('#upon-pane').height();		
      var downPaneHeight = element.find('#down-pane').height();
      
      var uponTableHeight = element.find('#upon-pane .table-body-container').height();
      var downTableHeight = element.find('#down-pane .table-body-container').height();
      
      var handleHeight = element.find('.split-handle').height() + element.find('.split-handle').offset().top + element.find('.well').height() - 32;		
      console.log('handleHeight:' + handleHeight);
      var maxDownPaneHeight = 760;		
                      		
      element.find('.split-handle').on('mousedown', mousedownHandler);		
     		
      function mousedownHandler(e) {		
        e.preventDefault();		
        $(document).on('mousemove', mousemoveHandler);    		
        $(document).on('mouseup', mouseupHandler);		
      }		
      		
      function mousemoveHandler(e) {		
        
        var incrementHeight = $('.container-fluid').scrollTop() + e.pageY;
                
        if(element.find('#down-pane').height() <= maxDownPaneHeight) {		
          element.find('#upon-pane').css({'height' : (uponPaneHeight - (handleHeight - incrementHeight)) + 'px'});		
          element.find('#down-pane').css({'height' : (downPaneHeight + (handleHeight - incrementHeight)) + 'px'});  		
          element.find('#upon-pane .table-body-container').css({'height': (uponTableHeight - (handleHeight - incrementHeight)) + 'px'});
          element.find('#down-pane .table-body-container').css({'height': (downTableHeight + (handleHeight - incrementHeight)) + 'px'});
        }else{
          element.find('#down-pane').css({'height' : (maxDownPaneHeight) + 'px'});
          $(document).off('mousemove');		
        }			
      }		
      
      function mouseupHandler(e) {		
        $(document).off('mousedown');		
        $(document).off('mousemove');		
      }
   
    
      ctrl.lastPolicyId = -1;          
      
      scope.$watch('vm.replicationPolicies', function(current) { 
        $timeout(function(){
          if(current) {
            if(current.length > 0) {
              element.find('#upon-pane table>tbody>tr').on('click', trClickHandler);
              element.find('#upon-pane table>tbody>tr:eq(0)').trigger('click');
            }else{
               element
                .find('#upon-pane table>tbody>tr')  
                .css({'background-color': '#FFFFFF'})
                .css({'color': '#000'});
            }
          }
        });
      });
         
      function trClickHandler(e) {
        element
          .find('#upon-pane table>tbody>tr')  
          .css({'background-color': '#FFFFFF'})
          .css({'color': '#000'})
          .css({'cursor': 'default'});
        element
          .find('#upon-pane table>tbody>tr a')
          .css({'color': '#337ab7'});          
        $(this)
          .css({'background-color': '#057ac9'})
          .css({'color': '#fff'});
        $('a', this)
          .css({'color': '#fff'});
        ctrl.retrieveJob($(this).attr('policy_id'), ctrl.page, ctrl.pageSize);
        ctrl.lastPolicyId = $(this).attr('policy_id');
      }
      
      element.find('.datetimepicker').datetimepicker({
				locale: I18nService().getCurrentLanguage(),
				ignoreReadonly: true,
				format: 'L',
				showClear: true
		  });      
      element.find('#fromDatePicker').on('blur', function(){
        ctrl.pickUp({'key': 'fromDate', 'value': $(this).val()});
      });
      element.find('#toDatePicker').on('blur', function(){
        ctrl.pickUp({'key': 'toDate', 'value': $(this).val()});
      });
      
      element.find('#txtSearchPolicyInput').on('keydown', function(e) {
        if($(this).is(':focus') && e.keyCode === 13) {
          ctrl.searchReplicationPolicy();
        }
      });
      
      element.find('#txtSearchJobInput').on('keydown', function(e) {
        if($(this).is(':focus') && e.keyCode === 13) {
          ctrl.searchReplicationJob();
        }
      });
      
    }
  }
  
})();
