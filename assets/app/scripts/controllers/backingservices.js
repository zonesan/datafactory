'use strict';

/**
 * @ngdoc function
 * @name openshiftConsole.controller:BackingservicesController
 * @description
 * # ProjectController
 * Controller of the openshiftConsole
 */
angular.module('openshiftConsole')
  .controller('BackingservicesController', function ($scope, AuthService, DataService) {

    AuthService.withUser().then(function() {
      loadBackingServices();
    });

    var loadBackingServices = function() {
      DataService.list("backingservices", $scope, function(backingservices){
        $scope.backingservices = backingservices.by("metadata.name");
        console.log("backingservices", $scope.backingservices);
      });
    };

  });
