'use strict';

/**
 * @ngdoc function
 * @name openshiftConsole.controller:BuildsController
 * @description
 * # ProjectController
 * Controller of the openshiftConsole
 */
angular.module('openshiftConsole')
  .controller('BackingserviceInstancesController', function ($scope, AuthService, DataService) {
    $scope.emptyMessage = 'No instances to show';
    AuthService.withUser().then(function() {
      loadBackingServiceInstances();
    });

    var loadBackingServiceInstances = function() {
      DataService.list("backingserviceinstances", $scope, function(backingserviceinstances){
        $scope.backingserviceinstances = backingserviceinstances.by("metadata.name");
        console.log("backingserviceinstances", $scope.backingserviceinstances);
      });
    };

  });
