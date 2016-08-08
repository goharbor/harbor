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
    .module('harbor.details')
    .directive('nodeCustom', nodeCustom);

  NodeCustomController.$inject = ['$scope', 'ListCustomService', '$location', '$window', '$filter', 'trFilter', 'getParameterByName'];

  function NodeCustomController($scope, ListCustomService, $location, $window, $filter, trFilter, getParameterByName) {
    var vm = this;
    vm.retrieve = retrieve;

    $scope.$watch('vm.projectId', function(current, origin) {
      if(current) {
        console.log('customs:' + current);
        vm.retrieve();
      }
    });

    $scope.$watch('vm.selectedCustom', function(current, origin) {
      if(current) {
        vm.selectedId = current.id;
      }
    });

    $scope.$on('selectedCustomNull', function(e, val) {
        vm.selectedCustom = null;
    });

    function retrieve(){
      ListCustomService(vm.projectId)
        .success(getCustomComplete)
        .error(getCustomFailed);
    }

    function selectItem(custom) {
      vm.selectedCustom = item;
    }

    function getCustomComplete(data, status) {
      vm.customs = data || [];
    }

    function getCustomFailed(response) {
      console.log('Failed to list repositories:' + response);
    }
  }

  function nodeCustom() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/details/node-custom.directive.html',
      'scope': {
        'projectId': '=',
        'selectedCustom': '=',
      },
      'replace': true,
      'controller': NodeCustomController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }

})();
