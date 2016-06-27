(function() {
  
  'use strict';
  
  angular
    .module('harbor.replication')
    .directive('listReplication', listReplication);
    
  ListReplicationController.$inject = ['$scope', 'getParameterByName', '$location', 'ListReplicationPolicyService', 'ToggleReplicationPolicyService', 'ListReplicationJobService', '$window', '$filter', 'trFilter'];
  
  function ListReplicationController($scope, getParameterByName, $location, ListReplicationPolicyService, ToggleReplicationPolicyService, ListReplicationJobService, $window, $filter, trFilter) {
    var vm = this;
    
    $scope.$on('$locationChangeSuccess', function() {
      vm.projectId = getParameterByName('project_id', $location.absUrl());
      vm.retrievePolicy();
    });
    
    vm.addReplication = addReplication;
    vm.editReplication = editReplication;
    
    vm.searchReplicationPolicy = searchReplicationPolicy;
    vm.searchReplicationJob = searchReplicationJob;
    
    vm.retrievePolicy = retrievePolicy;
    vm.retrieveJob = retrieveJob;
    vm.togglePolicy = togglePolicy;
    
    vm.downloadLog = downloadLog;
      
    vm.last = false;
    
    vm.projectId = getParameterByName('project_id', $location.absUrl());
    vm.retrievePolicy();
    
    function searchReplicationPolicy() {
      vm.retrievePolicy();
    }   
    
    function searchReplicationJob() {
      if(vm.lastPolicyId !== -1) {
        vm.retrieveJob(vm.lastPolicyId);
      }
    }            
   
    function retrievePolicy() {
      ListReplicationPolicyService('', vm.projectId, vm.replicationPolicyName)
        .success(listReplicationPolicySuccess)
        .error(listReplicationPolicyFailed);
    }
    
    function retrieveJob(policyId) {
      ListReplicationJobService(policyId, vm.replicationJobName)
        .success(listReplicationJobSuccess)
        .error(listReplicationJobFailed);
    }

    function listReplicationPolicySuccess(data, status) {
      vm.replicationJobs = [];
      vm.replicationPolicies = data || [];
    }
    
    function listReplicationPolicyFailed(data, status) {
      console.log('Failed list replication policy:' + data);
    }

    function listReplicationJobSuccess(data, status) {
      vm.replicationJobs = data || [];
    }
    
    function listReplicationJobFailed(data, status) {
      console.log('Failed list replication job:' + data);
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
     
    function togglePolicy(policyId, enabled) {
      ToggleReplicationPolicyService(policyId, enabled)
        .success(toggleReplicationPolicySuccess)
        .error(toggleReplicationPolicyFailed);
    }
    
    function toggleReplicationPolicySuccess(data, status) {
      console.log('Successful toggle replication policy.');
      vm.retrievePolicy();
    }
    
    function toggleReplicationPolicyFailed(data, status) {
      console.log('Failed toggle replication policy.');
    }
    
    function downloadLog(policyId) {
      $window.open('/api/jobs/replication/' + policyId + '/log', '_blank');
    }
  }
  
  function listReplication($timeout) {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/replication/list-replication.directive.html',
      'scope': true,
      'link': link,
      'controller': ListReplicationController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      var uponPaneHeight = element.find('#upon-pane').height();
      var handleHeight = element.find('.split-handle').height() + element.find('.split-handle').offset().top + element.find('.well').height() - 24;
      
      var maxDownPaneHeight = 245;
            
      element.find('.split-handle').on('mousedown', mousedownHandler);
      
      function mousedownHandler(e) {
        e.preventDefault();
        $(document).on('mousemove', mousemoveHandler);    
        $(document).on('mouseup', mouseupHandler);
      }
      
      function mousemoveHandler(e) {
        if(element.find('#down-pane').height() <= maxDownPaneHeight) {
          element.find('#upon-pane').css({'height' : (uponPaneHeight - (handleHeight - e.pageY)) + 'px'});
          element.find('#down-pane').css({'height' : (uponPaneHeight + (handleHeight - e.pageY - 196)) + 'px'});  
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
              if(ctrl.lastPolicyId === -1) {
                element.find('#upon-pane table>tbody>tr:eq(0)').trigger('click');  
              }else{
                element.find('#upon-pane table>tbody>tr').filter('[policy_id="' + ctrl.lastPolicyId + '"]').trigger('click');
              }
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
        ctrl.retrieveJob($(this).attr('policy_id'));
        ctrl.lastPolicyId = $(this).attr('policy_id');
      }
    }
  }
  
})();