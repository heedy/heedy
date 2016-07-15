// http://www.html5rocks.com/en/tutorials/service-worker/introduction/

// Increment the version if there were changes
var CACHE_NAME = 'v2';

// getPath returns the server path of the resource being requested
function getPath(request) {
    u = new URL(request.url);
    if (request.referrer === "") {
        return u.pathname;
    }
    uref = new URL(request.referrer)
    if (u.host === uref.host) {
        return u.pathname;
    }
    return request.url;
}

self.addEventListener('install', function(event) {
    console.log("Installing ServiceWorker...");
});

// On activate, we reset the entire cache
self.addEventListener('activate', function(event) {
    event.waitUntil(caches.keys().then(function(keyList) {
        return Promise.all(keyList.map(function(key) {
            return caches.delete(key);
        }));
    }));
});

self.addEventListener('fetch', function(event) {
    var rpath = getPath(event.request);

    if (rpath == "/logout" || rpath == "/login" || rpath == "/api/v1/login") {
        console.log("uncache everything...");
        // On logout, uncache everything
        event.respondWith(caches.keys().then(function(keyList) {
            return Promise.all(keyList.map(function(key) {
                return caches.delete(key);
            })).then(() => fetch(event.request));
        }));
    } else if (rpath.startsWith("/api/")) {
        event.respondWith(fetch(event.request));
    } else {
        // We're not logging out

        event.respondWith(caches.match(event.request).then(function(response) {
            // Cache hit - return response
            if (response) {
                //console.log("Using Cached:", rpath);
                return response;
            }

            // IMPORTANT: Clone the request. A request is a stream and
            // can only be consumed once. Since we are consuming this
            // once by cache and once by the browser for fetch, we need
            // to clone the response
            var fetchRequest = event.request.clone();

            // TODO: Add special logic for /user/device/stream
            // For now, we just cache all of the visited pages - which is suboptimal
            if (rpath.startsWith("/app/") || rpath.startsWith("/www/") || rpath.startsWith("/")) {
                console.log("Adding to cache:", rpath);
                return fetch(fetchRequest).then(function(response) {
                    // Check if we received a valid responnp
                    if (!response || response.status !== 200 || response.type !== 'basic') {
                        return response;
                    }

                    // IMPORTANT: Clone the response. A response is a stream
                    // and because we want the browser to consume the response
                    // as well as the cache consuming the response, we need
                    // to clone it so we have 2 stream.
                    var responseToCache = response.clone();

                    caches.open(CACHE_NAME).then(function(cache) {
                        cache.put(event.request, responseToCache);
                    });

                    return response;
                });
            }

            //console.log("Not Cached:", rpath);
            return fetch(fetchRequest);
        }));
    }
});
