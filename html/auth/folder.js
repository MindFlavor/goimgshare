/**
 * Created by mindflavor on 15/10/2015.
 */
'use strict';

// Declare app level module which depends on views, and components
angular.module('folder', [
    'ngRoute'
]).controller('Main', ['$scope', '$http', '$window', '$location', function ($scope, $http, $window, $location) {
    var idx = $location.absUrl().indexOf("?id=")
    var id = $location.absUrl().substring(idx + 4)
    console.log(id)

    if ($scope.remoteImages === undefined) {
        $scope.remoteImages = null;
    }
    if ($scope.remoteVideos === undefined) {
        $scope.remoteVideos = null;
    }
    if ($scope.remoteExtra === undefined) {
        $scope.remoteExtra = null;
    }

    $scope.getItemURL = function(item) {
        return '/file/' + id + '/' + item.Name;
    };

    $scope.getSmallThumbnailURL = function(item) {
        var s = '/smallthumb/' + id + '/' + item.Name
        console.log("Called smallthumb: " + s)
        return s;
    };

    $scope.getAvgThumbnailURL = function(item) {
        return '/avgthumb/' + id + '/' + item.Name;
    };

    $scope.retrieveImages = function () {
        console.log('Retrieving images')
        $http.get('/images/' + id).
            success(function (data) {
                console.log('Images retrieved from site')
                $scope.remoteImages = data
            }).
            error(function (data, status, headers, config) {
                console.log('Images *not* retrieved')
                $scope.remoteImages = null
                $window.location.href = '/';
            });
    };

    $scope.retrieveVideos = function () {
        console.log('Retrieving videos')
        $http.get('/videos/' + id).
            success(function (data) {
                console.log('Videos retrieved from site')
                $scope.remoteVideos = data
            }).
            error(function (data, status, headers, config) {
                console.log('Videos *not* retrieved')
                $scope.remoteVideos = null
                $window.location.href = '/';
            });
    };

    $scope.retrieveExtra = function () {
        console.log('Retrieving extra')
        $http.get('/extra/' + id).
            success(function (data) {
                console.log('extra retrieved from site')
                $scope.remoteExtra = data
            }).
            error(function (data, status, headers, config) {
                console.log('extra *not* retrieved')
                $scope.remoteExtra = null
                $window.location.href = '/';
            });
    };

    if ($scope.remoteImages === null) {
        $scope.retrieveImages();
    }
    if ($scope.remoteVideos === null) {
        $scope.retrieveVideos();
    }
    if ($scope.remoteExtra === null) {
        $scope.retrieveExtra();
    }
}]);
