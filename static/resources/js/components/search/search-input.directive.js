(function() {
  
  'use strict';
  
  angular
    .module('harbor.search')
    .directive('searchInput', searchInput);
    
  SearchInputController.$inject = ['$scope', '$location', '$window'];
  
  function SearchInputController($scope, $location, $window) {
    var vm = this;

    vm.searchFor = searchFor;
    
    function searchFor(searchContent) {
      $location
        .path('/search')
        .search({'q': searchContent});
      $window.location.href = $location.url();
    }
    
  }
  
  function searchInput() {
    
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/search/search-input.directive.html',
      'scope': {
        'searchInput': '=',
      },
      'link': link,
      'controller': SearchInputController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      element
        .find('input[type="text"]')
        .on('keydown', keydownHandler);
        
      function keydownHandler(e) {
        if(e.keyCode === 13) {
          ctrl.searchFor($(this).val());
        }
      }
      
    }
  }
  
})();