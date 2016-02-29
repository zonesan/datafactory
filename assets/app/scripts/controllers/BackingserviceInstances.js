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

    AuthService.withUser().then(function() {
      loadBackingServices();
    });

    var loadBackingServices = function() {
      DataService.list("backingserviceinstances", $scope, function(backingservices){
        $scope.backingservices = backingservices.by("metadata.name");
        console.log("backingservices", $scope.backingservices);
      });
    };

  });
