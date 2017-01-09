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
    .module('harbor.paginator')
    .directive('paginator', paginator);
    
  PaginatorController.$inject = [];
  
  function PaginatorController() {
    var vm = this;   
  }
  
  paginator.$inject = [];
  
  function paginator() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/paginator/paginator.directive.html',
      'scope': {
        'totalCount': '@',
        'pageSize': '@',
        'page': '=',
        'displayCount': '@'
      },
      'link': link,
      'controller': PaginatorController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      scope.$watch('vm.page', function(current) {
        if(current) { 
          ctrl.page = current;
          togglePageButton();
        }
      });
      
      var tc;
                        
      scope.$watch('vm.totalCount', function(current) {
        if(current) {
          var totalCount   = current;   
                                                                  
          tc = new TimeCounter();
          
          console.log('Total Count:' + totalCount + ', Page Size:' + ctrl.pageSize + ', Display Count:' + ctrl.displayCount + ', Page:' + ctrl.page);

          ctrl.buttonCount = Math.ceil(totalCount / ctrl.pageSize);
          
          if(ctrl.buttonCount <= ctrl.displayCount) {
            tc.setMaximum(1);
            ctrl.visible = false;
          }else{
            tc.setMaximum(Math.ceil(ctrl.buttonCount / ctrl.displayCount));
            ctrl.visible = true;
          }
                   
          ctrl.gotoFirst = gotoFirst;
          ctrl.gotoLast = gotoLast;

          if(ctrl.buttonCount < ctrl.page) {
            ctrl.page = ctrl.buttonCount;
          }                   
                    
          ctrl.previous = previous;    
          ctrl.next = next;
                    
          drawButtons(tc.getTime());    

          togglePrevious(tc.canDecrement());
          toggleNext(tc.canIncrement());
                    
          toggleFirst();
          toggleLast();
          
          togglePageButton();
          
        }
      });

      var TimeCounter = function() {
        this.time = 0;
        this.minimum = 0;
        this.maximum = 0;
      };
      
      TimeCounter.prototype.setMaximum = function(maximum) {
        this.maximum = maximum;
      };
      
      TimeCounter.prototype.getMaximum = function() {
        return this.maximum;
      };
      
      TimeCounter.prototype.increment = function() {
        if(this.time < this.maximum) {
          ++this.time;
          if((ctrl.page % ctrl.displayCount) != 0) {
            ctrl.page = this.time * ctrl.displayCount;
          }
          ++ctrl.page;
        }
      };
      
      TimeCounter.prototype.canIncrement = function() {
        if(this.time + 1 < this.maximum) {
          return true;
        }
        return false;
      };
      
      TimeCounter.prototype.decrement = function() {
        if(this.time > this.minimum) {         
          if(this.time === 0) {
            ctrl.page = ctrl.displayCount;
          }else{
            ctrl.page =  this.time * ctrl.displayCount;
          }
          --this.time;
        }
      };
      
      TimeCounter.prototype.canDecrement = function() {
        if(this.time > this.minimum) {
          return true;
        }
        return false;
      };
      
      TimeCounter.prototype.getTime = function() {
        return this.time;
      };
      
      TimeCounter.prototype.setTime = function(time) {
        this.time = time;
      };
                 
      function drawButtons(time) {
        element.find('li[tag="pagination-button"]').remove();
        var buttons = [];
        for(var i = 1; i <= ctrl.displayCount; i++) {
          var displayNumber = ctrl.displayCount * time + i;
          if(displayNumber <= ctrl.buttonCount) {
            buttons.push('<li tag="pagination-button"><a href="javascript:void(0)" page="' + displayNumber + '">' + displayNumber + '<span class="sr-only"></span></a></li>');
          }
        }
        
        $(buttons.join(''))
          .insertAfter(element.find('ul li:eq(' + (ctrl.visible ? 1 : 0) + ')')).end()
          .on('click', buttonClickHandler); 
      }
      
      function togglePrevious(status) {
        ctrl.disabledPrevious = status ? '' : 'disabled';    
        toggleFirst();
        toggleLast();    
      }          
                   
      function toggleNext(status) {
        ctrl.disabledNext = status ? '' : 'disabled';        
        toggleFirst();
        toggleLast();
      }
      
      function toggleFirst() {
        ctrl.disabledFirst = (ctrl.page > 1) ? '' : 'disabled';
      }
      
      function toggleLast() {
        ctrl.disabledLast = (ctrl.page < ctrl.buttonCount) ? '' : 'disabled';
      }
          
      function buttonClickHandler(e) {
        ctrl.page = $(e.target).attr('page');              
        togglePageButton();                                      
        togglePrevious(tc.canDecrement());
        toggleNext(tc.canIncrement());
       
        scope.$apply();
      }  
          
      function togglePageButton() {
        element.find('li[tag="pagination-button"]').removeClass('active');
        element.find('li[tag="pagination-button"] a[page="' + ctrl.page + '"]').parent().addClass('active');
      }          
         
      function previous() {
        if(tc.canDecrement()) {
          tc.decrement();
          drawButtons(tc.getTime());
          togglePageButton();
          togglePrevious(tc.canDecrement());
          toggleNext(tc.canIncrement());
        }
      }      
      
      function gotoFirst() {     
        ctrl.page = 1;
        tc.setTime(0);
        drawButtons(0);
        
        toggleFirst();
        toggleLast();
        
        togglePageButton();
        togglePrevious(tc.canDecrement());
        toggleNext(tc.canIncrement());
      }
          
      function next() {
        if(tc.canIncrement()) {    
          tc.increment();
          drawButtons(tc.getTime());     
          togglePageButton();
          togglePrevious(tc.canDecrement());
          toggleNext(tc.canIncrement());
        }
      }
      
      function gotoLast() {
        ctrl.page = ctrl.buttonCount;
        tc.setTime(Math.ceil(ctrl.buttonCount / ctrl.displayCount) - 1);
        drawButtons(tc.getTime()); 
        
        toggleFirst();
        toggleLast();
        
        togglePageButton();
        togglePrevious(tc.canDecrement());
        toggleNext(tc.canIncrement());
      }
    }
  }
  
})();