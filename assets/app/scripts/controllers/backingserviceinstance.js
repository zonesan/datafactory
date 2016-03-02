'use strict';

/**
 * @ngdoc function
 * @name openshiftConsole.controller:BuildsController
 * @description
 * # ProjectController
 * Controller of the openshiftConsole
 */
angular.module('openshiftConsole')
  .controller('backingserviceinstancecontroller', function ($scope, $routeParams, $filter, AuthService, DataService,ProjectsService) { 
    $scope.emptyMessage = 'No instances to show';
    $scope.alerts = {};
    AuthService.withUser().then(function() {
      loadbsi();
    });

    $scope.breadcrumbs = [
      {
        title: "Backing Service Instances",
        link: "project/" + $routeParams.project + "/browse/backingserviceinstances"
      }
    ];

    $scope.breadcrumbs.push({
      title: $routeParams.backingservice
    });

    var loadbsi = function(){
      ProjectsService
      .get($routeParams.project)
      .then(_.spread(function(project, context) {
        $scope.project = project;

            DataService.get("backingserviceinstances", $routeParams.backingserviceinstance, context).then(
              // success
              function(bsi) {
                $scope.loaded = true;
                $scope.backingserviceinstance = bsi;
                console.log('backingserviceinstance', bsi);
              },
              // failure
              function(e) {
                $scope.loaded = true;
                $scope.alerts["load"] = {
                  type: "error",
                  message: "The build details could not be loaded.",
                  details: "Reason: " + $filter('getErrorDetails')(e)
                };
              }
            );
         })   
        );   
      };
    });

       
      

