/**
Copyright 2015 Joseph Lewis <joseph@josephlewis.net>

All rights reserved.
**/

var connectordbApp = angular.module('connectordbApp', ['ngRoute']);

// configure our routes
connectordbApp.config(function($routeProvider) {
    $routeProvider

        .when('/', {
            templateUrl : '/app/pages/explore.html',
            controller  : 'exploreController'
        })

        .when('/calculate', {
            templateUrl : '/app/pages/calculate.html',
            controller  : 'calculateController'
        })

        .when('/data', {
            templateUrl : '/app/pages/data.html',
            controller  : 'dataController'
        })

        .when('/data/:devicename/:streamname', {
            templateUrl : '/app/pages/stream.html',
            controller  : 'streamController'
        })

        .when('/data/:devicename', {
            templateUrl : '/app/pages/device.html',
            controller  : 'deviceController'
        })

        .when('/settings', {
            templateUrl : '/app/pages/settings.html',
            controller  : 'settingsController'
        })

        .when('/track', {
            templateUrl : '/app/pages/track.html',
            controller  : 'trackController'
        });
});

// create the controller and inject Angular's $scope
connectordbApp.controller('exploreController', function($scope) {
    // create a message to display in our view
    $scope.message = 'More to come here!';
    $scope.title = 'Explore';
    $scope.devices = [{"name":"user","nickname":"","apikey":"joseph-cec7fbee-3d55-4664-67f6-3e866472ff17","enabled":true,"admin":true,"canwrite":true,"user":true,"visible":true},{"name":"foo","nickname":"","apikey":"bd234c95-089c-4c81-5cda-b74bb5536149","enabled":true,"canwrite":true,"visible":true},{"name":"bar","nickname":"","apikey":"e666c018-e3f2-46e1-6acc-1cfd9a3d1635","enabled":true,"canwrite":true,"visible":true},{"name":"Nexus_4","nickname":"","apikey":"5c00d98c-2aa9-4f41-71b1-177c7e5f4971","enabled":true,"canwrite":true,"visible":true}];
});

// create the controller and inject Angular's $scope
connectordbApp.controller('streamController', function($rootScope, $scope, $routeParams, $route) {
    // create a message to display in our view
    $scope.devicename = $routeParams.devicename;
    $scope.streamname = $routeParams.streamname
    $scope.title = $scope.message;
    $scope.stream = {"name":"activity","schema":{"type":"string"}};
    $scope.devices = [{"name":"user","nickname":"","apikey":"joseph-cec7fbee-3d55-4664-67f6-3e866472ff17","enabled":true,"admin":true,"canwrite":true,"user":true,"visible":true},{"name":"foo","nickname":"","apikey":"bd234c95-089c-4c81-5cda-b74bb5536149","enabled":true,"canwrite":true,"visible":true},{"name":"bar","nickname":"","apikey":"e666c018-e3f2-46e1-6acc-1cfd9a3d1635","enabled":true,"canwrite":true,"visible":true},{"name":"Nexus_4","nickname":"","apikey":"5c00d98c-2aa9-4f41-71b1-177c7e5f4971","enabled":true,"canwrite":true,"visible":true}];
});

// create the controller and inject Angular's $scope
connectordbApp.controller('deviceController', function($rootScope, $scope, $routeParams, $route) {
    // create a message to display in our view
    $scope.device = {
        "name":"user",
        "nickname":"",
        "apikey":"joseph-cec7fbee-3d55-4664-67f6-3e866472ff17",
        "enabled":true,
        "admin":true,
        "canwrite":true,
        "user":true,
        "visible":true};


    $scope.devicename = $scope.device.nickname || $routeParams.devicename;
    $scope.title = $routeParams.devicename;

    $scope.streams = [{"name":"activity","schema":{"type":"string"}}, {"name":"location","schema":{"type":"string"}}];
});


// create the controller and inject Angular's $scope
connectordbApp.controller('calculateController', function($scope) {
    // create a message to display in our view
    $scope.message = 'More to come here!';
});


// create the controller and inject Angular's $scope
connectordbApp.controller('dataController', function($scope) {
    // create a message to display in our view
    $scope.devices = [{"name":"user","nickname":"","apikey":"joseph-cec7fbee-3d55-4664-67f6-3e866472ff17","enabled":true,"admin":true,"canwrite":true,"user":true,"visible":true},{"name":"foo","nickname":"","apikey":"bd234c95-089c-4c81-5cda-b74bb5536149","enabled":true,"canwrite":true,"visible":true},{"name":"bar","nickname":"","apikey":"e666c018-e3f2-46e1-6acc-1cfd9a3d1635","enabled":true,"canwrite":true,"visible":true},{"name":"Nexus_4","nickname":"","apikey":"5c00d98c-2aa9-4f41-71b1-177c7e5f4971","enabled":true,"canwrite":true,"visible":true}];
});

function getStreams(deviceName) {
        return [{"name":"location","schema":{"properties":{"accuracy":{"type":"number"},"altitude":{"type":"number"},"bearing":{"type":"number"},"latitude":{"type":"number"},"longitude":{"type":"number"},"speed":{"type":"number"}},"required":["latitude","longitude"],"type":"object"}},{"name":"plugged_in","schema":{"type":"boolean"}},{"name":"screen_on","schema":{"type":"boolean"}},{"name":"steps","schema":{"type":"number"}},{"name":"heart_rate","schema":{"type":"number"}},{"name":"activity","schema":{"type":"string"}}];
}


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
