/**
 * Created by frcogno on 19/10/2015.
 */
'use strict';

// Declare app level module which depends on views, and components
angular.module('auth', [
    'ngRoute'
]).controller('Main', ['$scope', '$http', function ($scope, $http) {

    if ($scope.supportedAuths === undefined) {
        $scope.supportedAuths = null;
    }

    $scope.getSupportedAuths = function () {
        console.log('Retrieving supportedAuths')
        $http.get('/supportedAuths').
            success(function (data) {
                console.log('supportedAuths retrieved from site')
                $scope.supportedAuths = data
            }).
            error(function (data, status, headers, config) {
                console.log('supportedAuths *not* retrieved')
                $scope.supportedAuths = null
            });
    };

    if ($scope.supportedAuths === null) {
        $scope.getSupportedAuths()
    }

    $scope.IsSupportedAuth = function(auth) {
        if ($scope.supportedAuths === null) return false;
        for (var i=0; i<$scope.supportedAuths.length; i++) {
            if ($scope.supportedAuths[i] === auth) {
                return true;
            }
        }
        return false;
    }
}]);