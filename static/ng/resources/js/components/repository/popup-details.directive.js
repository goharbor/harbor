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
        'tag': '@',
        'index': '@'
      },
      'replace': true,
      'link': link,
      'controller': PopupDetailsController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      ctrl.retrieve();
      scope.$watch('vm.manifest', function(current) {
        if(current) {
          
          element
            .popover({
              'template': '<div class="popover" role="tooltip"><div class="arrow"></div><div class="popover-title"></div><div class="popover-content"></div></div>',
              'title': '<div class="pull-right clearfix"><a href="javascript:void(0);"><span class="glyphicon glyphicon-remove-circle"></span></a></div>',
              'content': generateContent,
              'html': true
            })
            .on('shown.bs.popover', function(e){      
              var self = jQuery(this);                 
              $('[type="text"]:input', self.parent())
                .select()
                .end()
                .on('click', function() {
                  $(this).select();
                });
              self.parent().find('.glyphicon.glyphicon-remove-circle').on('click', function() {
                element.trigger('click');
              });
            });
        }
      });
      function generateContent() {
        var content =  '<form class="form-horizontal">' +
          '<div class="form-group">' +
          '<label class="col-sm-3 control-label">Id</label>' +
          '<div class="col-sm-9"><p class="form-control-static long-line"><input type="text" id="txtImageId" value="' + ctrl.manifest['Id'] + '" readonly size="40"></p></div></div>' +
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