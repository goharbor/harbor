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
    .module('harbor.layout.element.height')
    .directive('elementHeight', elementHeight);
    
  function elementHeight($window) {
    var directive = {
      'restrict': 'A',
      'link': link
    };
    
    return directive;
    
    function link(scope, element, attrs) {

      var w = angular.element($window);

      scope.getDimension = function() {
        return {'h' : w.height()};
      };
      
      if(!angular.isDefined(scope.subsHeight))  {scope.subsHeight = 110;}
      if(!angular.isDefined(scope.subsSection)) {scope.subsSection = 32;}
      if(!angular.isDefined(scope.subsSubPane)) {scope.subsSubPane = 226;}
      if(!angular.isDefined(scope.subsTblBody)) {scope.subsTblBody = 40;}
    
      scope.$watch(scope.getDimension, function(current) {
        if(current) {
          var h = current.h; 
          element.css({'height': (h - scope.subsHeight) + 'px'});
          element.find('.section').css({'height': (h - scope.subsHeight - scope.subsSection) + 'px'});        
          element.find('.sub-pane').css({'height': (h - scope.subsHeight - scope.subsSubPane) + 'px'});
          element.find('.tab-pane').css({'height': (h - scope.subsHeight - scope.subsSubPane - scope.subsSection -100) + 'px'});
//          var subPaneHeight = element.find('.sub-pane').height();
//          element.find('.table-body-container').css({'height': (subPaneHeight - scope.subsTblBody) + 'px'});
        }
      }, true);
      
      w.on('pageshow, resize', function() {       
        scope.$apply();
      });
    }
  }
  
})();
