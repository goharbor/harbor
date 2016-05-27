(function() {
  
  'use strict';
  
  angular
    .module('harbor.replication')
    .directive('listReplication', listReplication);
    
  ListReplicationController.$inject = ['ListReplicationPolicyService', 'ListReplicationJobService'];
  
  function ListReplicationController(ListReplicationPolicyService, ListReplicationJobService) {
    var vm = this;
    
    vm.addReplication = addReplication;
    vm.retrievePolicy = retrievePolicy;
    vm.retrieveJob = retrieveJob;
    vm.last = false;
    
    vm.retrievePolicy();
   
    function retrievePolicy() {
      ListReplicationPolicyService()
        .then(listReplicationPolicySuccess, listReplicationPolicyFailed);
    }
    
    function retrieveJob(policyId) {
      ListReplicationJobService(policyId)
        .then(listReplicationJobSuccess, listReplicationJobFailed);
    }

    function listReplicationPolicySuccess(data, status) {
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
      vm.modalTitle = 'Create New Policy';
      vm.modalMessage = '';
    }
    
  }
  
  function listReplication($timeout) {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/replication/list-replication.directive.html',
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
      
      scope.$watch('vm.last', function(current) { 
        $timeout(function(){
          if(current) {
            element.find('#upon-pane table>tbody>tr').on('click', trClickHandler);
            element.find('#upon-pane table>tbody>tr:eq(0)').trigger('click');
          }
        });
      });
      
      function trClickHandler(e) {
        element
          .find('#upon-pane table>tbody>tr')  
          .css({'background-color': '#FFFFFF'})
          .css({'color': '#000'});
        $(this)
          .css({'background-color': '#057ac9'})
          .css({'color': '#fff'});
        ctrl.retrieveJob($(this).attr('policy_id'));
      }
    }
  }
  
})();