'use strict';

/**
 * @ngdoc function
 * @name openshiftConsole.controller:BuildsController
 * @description
 * # ProjectController
 * Controller of the openshiftConsole
 */
angular.module('openshiftConsole')
  .controller('BackingServiceInstancesController', function ($scope, AuthService, DataService) {
    $scope.emptyMessage = 'No instances to show';
    AuthService.withUser().then(function() {
      loadBackingServiceInstances();
    });

    var loadBackingServiceInstances = function() {
      DataService.list("backingserviceinstances", $scope, function(backingserviceinstances){
        $scope.backingserviceinstances = backingserviceinstances.by("metadata.name");
        console.log("backingserviceinstances", $scope.backingserviceinstances);
        
        if ($scope.backingserviceinstances) {
          loadBackingServices($scope.backingserviceinstances);
        }
      });
    };

    var matchBs = function(bss, guid){
      for(var key in bss){
        var plans = bss[key].spec.plans;
        for(var k in plans){
          if (plans[k].id == guid) {
            return key;
          }
        }
      }
      return null;
    }

    var loadBackingServices = function(bsis) {
      DataService.list("backingservices", {}, function(bss){
        var bss = bss.by("metadata.name");
        console.log("bss", bss, "bsis", bsis);

        for(var key in bsis) {
          var bsName = matchBs(bss, bsis[key].spec.provisioning.backingserviceinstance_plan_guid);
          if(bsName){
            $scope.backingserviceinstances[key].bsName = bsName;
          }
        }
      });

    };

  });
