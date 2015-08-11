/**
Copyright 2015 Joseph Lewis <joseph@josephlewis.net>

All rights reserved.
**/

var connectordbApp = angular.module('connectordbApp', ['ngRoute']);

// configure our routes
connectordbApp.config(function($routeProvider) {
    $routeProvider

        .when('/', {
            templateUrl : 'pages/explore.html',
            controller  : 'exploreController'
        })

        .when('/calculate', {
            templateUrl : 'pages/calculate.html',
            controller  : 'calculateController'
        })

        .when('/data', {
            templateUrl : 'pages/data.html',
            controller  : 'dataController'
        })

        .when('/settings', {
            templateUrl : 'pages/settings.html',
            controller  : 'settingsController'
        })

        .when('/track', {
            templateUrl : 'pages/track.html',
            controller  : 'trackController'
        });
});

// create the controller and inject Angular's $scope
connectordbApp.controller('exploreController', function($scope) {
    // create a message to display in our view
    $scope.message = 'More to come here!';
});


// create the controller and inject Angular's $scope
connectordbApp.controller('calculateController', function($scope) {
    // create a message to display in our view
    $scope.message = 'More to come here!';
});


// create the controller and inject Angular's $scope
connectordbApp.controller('dataController', function($scope) {
    // create a message to display in our view
    $scope.message = 'More to come here!';
});


// create the controller and inject Angular's $scope
connectordbApp.controller('settingsController', function($scope) {
    // create a message to display in our view
    $scope.message = 'More to come here!';
});


// create the controller and inject Angular's $scope
connectordbApp.controller('trackController', function($scope) {
    // create a message to display in our view
    $scope.message = 'More to come here!';
});
