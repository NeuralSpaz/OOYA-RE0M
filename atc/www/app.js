var morfapp = new angular.module("MorfUI", []);

morfapp.controller("MainCtl", ["$scope", function($scope){


}]);
// function createWebSocket() {
// 	var loc = window.location, websocket_uri;
// 	if (loc.protocol === "https:") {
// 		websocket_uri = "wss:";
// 	} else {
// 		websocket_uri = "ws:";
// 	}
// 	websocket_uri += "//" + loc.host + "/ws";

// 	try {
// 		console.log("Opening Websocket to " + websocket_uri);
// 		ws = new WebSocket(websocket_uri);
// 	} catch ( e ) {
// 		console.warn("No websockets! failing back to xhr");
// 		fallbackLoadData();
// 		return;
// 	}			
	
// 	ws.onopen = function() {
// 		console.log("WebScoket Open");
// 		var newMsg = {
// 			cmd: "PING" 
// 		};
// 		ws.send(JSON.stringify(newMsg));
// 	}
// 	ws.onmessage = function( e ) {
// 		console.log("New Message");
// 		dostuffwithdata(JSON.parse(e));
// 	}
// 	ws.onclose = function( e ) {
// 		console.log("WebScoket Closed");
// 		time = setTimeout(function(){
// 			ws=createWebSocket();
// 		},100);
// 	}
// 	return ws
// }

// function fallbackLoadData(){
// 	console.log("Opening xhr connection");
// }

// ws = createWebSocket();

// var myVar = setInterval(function(){myTimer()},10);

// function myTimer() {
// 	var newMsg = {
// 		cmd: "PING" 
// 	};
// 	ws.send(JSON.stringify(newMsg));
// }

// function dostuffwithdata ( msg ) {
// 	console.log(msg);
// }




// // sender helper (socket, request, callback)
// angular.module('MorfUI').factory('MyService', ['$q', '$rootScope', function($q, $rootScope) {
//     // We return this object to anything injecting our service
//     var Service = {};
//     // Keep all pending requests here until they get responses
//     var callbacks = {};
//     // Create a unique callback ID to map requests to responses
//     var currentCallbackId = 0;
//     // Create our websocket object with the address to the websocket
//     var ws = new WebSocket("ws://localhost:8000/socket/");
    
//     ws.onopen = function(){  
//         console.log("Socket has been opened!");  
//     };
    
//     ws.onmessage = function(message) {
//         listener(JSON.parse(message.data));
//     };

//     function sendRequest(request) {
//       var defer = $q.defer();
//       var callbackId = getCallbackId();
//       callbacks[callbackId] = {
//         time: new Date(),
//         cb:defer
//       };
//       request.callback_id = callbackId;
//       console.log('Sending request', request);
//       ws.send(JSON.stringify(request));
//       return defer.promise;
//     }

//     function listener(data) {
//       var messageObj = data;
//       console.log("Received data from websocket: ", messageObj);
//       // If an object exists with callback_id in our callbacks object, resolve it
//       if(callbacks.hasOwnProperty(messageObj.callback_id)) {
//         console.log(callbacks[messageObj.callback_id]);
//         $rootScope.$apply(callbacks[messageObj.callback_id].cb.resolve(messageObj.data));
//         delete callbacks[messageObj.callbackID];
//       }
//     }
//     // This creates a new callback ID for a request
//     function getCallbackId() {
//       currentCallbackId += 1;
//       if(currentCallbackId > 10000) {
//         currentCallbackId = 0;
//       }
//       return currentCallbackId;
//     }

//     // Define a "getter" for getting customer data
//     Service.getCustomers = function() {
//       var request = {
//         type: "get_customers"
//       }
//       // Storing in a variable for clarity on what sendRequest returns
//       var promise = sendRequest(request); 
//       return promise;
//     }

//     return Service;
// }])
