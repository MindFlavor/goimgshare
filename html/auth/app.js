'use strict';

// Declare app level module which depends on views, and components
angular.module('myApp', [
  'ngRoute',
  'myApp.version'
]).
config(['$routeProvider', function($routeProvider) {
  $routeProvider.otherwise({redirectTo: '/view1'});
}]).
controller('Main', ['$scope', '$http', '$window', function ($scope, $http, $window) {
      if ($scope.permafolders === undefined) {
        $scope.permafolders = null;
      }

        if($scope.selectedFolder === undefined) {
            $scope.selectedFolder = null;
        }

      $scope.retrievePermafolders = function () {
        console.log('Retrieving folders')
        $http.get('/folders').
            success(function (data) {
                console.log('Data retrieved from site')
                $scope.permafolders = data
            }).
            error(function (data, status, headers, config) {
                console.log('Data *not* retrieved')
                $scope.permafolders = null
                $window.location.href = '/';
            });
      };

      if ($scope.permafolders === null) {
        $scope.retrievePermafolders();
      }

        $scope.isSelected = function(p) {
            if(p === $scope.selectedFolder)
                return "active";
            else
                return "std";
        }

        $scope.selectFolder = function(b) {
            $scope.selectedFolder = b
        }
    }]);



