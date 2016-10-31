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
    .module('harbor.layout.index')
    .controller('IndexController', IndexController);
    
  IndexController.$inject = ['$scope', '$filter', 'trFilter', '$timeout'];
    
  function IndexController($scope, $filter, trFilter, $timeout) {
    
    $scope.subsHeight = 110;
    $scope.subsSection = 32;
    $scope.subsSubPane = 226;
        
    var vm = this;
       
    vm.customBodyHeight = {'height': '180px'};
    vm.viewAll = viewAll;

    function viewAll() {
      var indexDesc = $filter('tr')('index_desc', []);
      var indexDesc1 = $filter('tr')('index_desc_1', []);
      var indexDesc2 = $filter('tr')('index_desc_2', []);
      var indexDesc3 = $filter('tr')('index_desc_3', []);
      var indexDesc4 = $filter('tr')('index_desc_4', []);
      var indexDesc5 = $filter('tr')('index_desc_5', []);
      var indexDesc6 = $filter('tr')('index_desc_6', []);
      
      $scope.$emit('modalTitle', $filter('tr')('harbor_intro_title'));
      $scope.$emit('modalMessage', '<p class="page-content text-justify">'+
        indexDesc + 
  			'</p>' +
        '<ul>' +
          '<li class="long-line">▪︎ ' + indexDesc1 + '</li>' +
          '<li class="long-line">▪︎ ' + indexDesc2 + '</li>' +
          '<li class="long-line">▪︎ ' + indexDesc3 + '</li>' +
          '<li class="long-line">▪︎ ' + indexDesc4 + '</li>' +
          '<li class="long-line">▪︎ ' + indexDesc5 + '</li>' +
          '<li class="long-line">▪︎ ' + indexDesc6 + '</li>' +
  			'</ul>');
      var emitInfo = {
        'contentType': 'text/html',
        'confirmOnly': true,
        'action': function() {
          $scope.$broadcast('showDialog', false);
        }
      };
      $scope.$emit('raiseInfo', emitInfo);
    }
    
    //Message dialog handler for index.
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
        
        $timeout(function() {    
          $scope.$broadcast('showDialog', true);
        }, 350);
      }
    });
   
   $scope.$on('raiseInfo', function(e, val) {
      if(val) {
        vm.action = function() {
          val.action();
          $scope.$broadcast('showDialog', false);
        };
        vm.contentType = val.contentType;
        vm.confirmOnly = val.confirmOnly;
       
        $scope.$broadcast('showDialog', true);
      }
    }); 
    
  }
        
})();
