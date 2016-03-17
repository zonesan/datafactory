'use strict';

/**
 * @ngdoc function
 * @name openshiftConsole.controller:BackingservicesController
 * @description
 * # ProjectController
 * Controller of the openshiftConsole
 */
angular.module('openshiftConsole')
  .controller('BackingservicesController', function ($scope, $routeParams, AuthService, DataService, ProjectsService) {

    $scope.emptyMessage = '现在没有数据...';

    var watches = [];

    ProjectsService
      .get($routeParams.project)
      .then(_.spread(function(project, context) {
        $scope.project = project;
        watches.push(DataService.watch("backingservices", context, function(backingservices) {
          $scope.backingservices = backingservices.by("metadata.name");
          console.log("backingservices", $scope.backingservices);
        }));
      })
    );

    $scope.$on('$destroy', function(){
      DataService.unwatchAll(watches);
    });
  });
