(function() {
  
  'use strict';
  
  angular
    .module('harbor.repository')
    .directive('popupDetails', popupDetails);
  
  PopupDetailsController.$inject = ['ListManifestService', '$filter', 'dateLFilter'];
  
  function PopupDetailsController(ListManifestService, $filter, dateLFilter) {
    var vm = this;

    vm.retrieve = retrieve;
    function retrieve() {
      ListManifestService(vm.repoName, vm.tag)
        .success(getManifestSuccess)
        .error(getManifestFailed);
    }
    
    function getManifestSuccess(data, status) {
      console.log('Successful get manifest:' + data);
      vm.manifest = data;      
      vm.manifest['Created'] = $filter('dateL')(vm.manifest['Created'], 'YYYY-MM-DD HH:mm:ss');
    }
    
    function getManifestFailed(data, status) {
      console.log('Failed get manifest:' + data);
    }
  }
  
  function popupDetails() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/repository/popup-details.directive.html',
      'scope': {
        'repoName': '@',
        'tag': '@'
      },
      'link': link,
      'controller': PopupDetailsController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      ctrl.retrieve();
      scope.$watch('vm.manifest', function(current, origin) {
        if(current) {
          element.find('span').popover({
            'content': generateContent,
            'html': true
          });  
        }
      });
      
      function generateContent() {
        var content = 
        '<form class="form-horizontal">' +
          '<div class="form-group">' +
          '<label class="col-sm-3 control-label">Id</label>' +
          '<div class="col-sm-9"><p class="form-control-static long-line">' + ctrl.manifest['Id'] + '</p></div></div>' +
          '<div class="form-group"><label class="col-sm-3 control-label">Parent</label>' +
          '<div class="col-sm-9"><p class="form-control-static long-line">' + ctrl.manifest['Parent'] + '</p></div></div>' +
          '<div class="form-group"><label class="col-sm-3 control-label">Created</label>' +
          '<div class="col-sm-9"><p class="form-control-static">' + ctrl.manifest['Created'] + '</p></div></div>' +
          '<div class="form-group"><label class="col-sm-3 control-label">Duration Days</label>' +
          '<div class="col-sm-9"><p class="form-control-static">' + (ctrl.manifest['Duration Days'] === '' ? 'N/A' : ctrl.manifest['Duration Days']) + ' days</p></div></div>' +
          '<div class="form-group"><label class="col-sm-3 control-label">Author</label>' +
          '<div class="col-sm-9"><p class="form-control-static">' + (ctrl.manifest['Author'] === '' ? 'N/A' : ctrl.manifest['Author']) + '</p></div></div>' +
          '<div class="form-group"><label class="col-sm-3 control-label">Architecture</label>' + 
          '<div class="col-sm-9"><p class="form-control-static">' + (ctrl.manifest['Architecture'] === '' ? 'N/A' : ctrl.manifest['Architecture']) + '</p></div></div>' +
          '<div class="form-group"><label class="col-sm-3 control-label">Docker Version</label>' +
          '<div class="col-sm-9"><p class="form-control-static">' + (ctrl.manifest['Docker Version'] === '' ? 'N/A' : ctrl.manifest['Docker Version']) + '</p></div></div>' +
          '<div class="form-group"><label class="col-sm-3 control-label">OS</label>' +
          '<div class="col-sm-9"><p class="form-control-static">' + (ctrl.manifest['OS']  === '' ? 'N/A' : ctrl.manifest['OS']) + '</p></div></div>' +
        '</form>';
        return content;
      }
    }
  }
  
})();