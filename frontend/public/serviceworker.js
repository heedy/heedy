const enableNavigationPreload = async () => {
    if (self.registration.navigationPreload) {
        await self.registration.navigationPreload.enable();
    }
};

const getNetwork = async({ request, preloadResponsePromise}) => {
    // Try to use the preloaded response, if it's there
    const preloadResponse = await preloadResponsePromise;
    if (preloadResponse) {
        console.debug('sw: (preload) ', request.url);
        return preloadResponse;
    }
    // Next try to get the resource from the network
    try {
        console.debug('sw: (net) ', request.url);
        return await fetch(request);
    } catch (error) {
        // there is nothing we can do, but we must always
        // return a Response object
        return new Response('Network Error', {
            status: 408,
            headers: { 'Content-Type': 'text/plain' },
        });
    }
}

{{if .DevMode}}
// If we are running in dev mode, we don't cache any results, so that every refresh
// always fetches from the server

self.addEventListener('activate', (event) => {
    event.waitUntil(enableNavigationPreload());
});

self.addEventListener("install", function (event) {
    console.debug("sw: installing (dev mode)");
    event.waitUntil(
        caches.keys().then(function(keyList) {
            return Promise.all(
              keyList.map(function(key) {
                return caches.delete(key);
              })
            );
          }).then(() => console.debug("sw: installed (dev mode)")));

    self.skipWaiting();
});

self.addEventListener("fetch", function (event) {
    event.respondWith(getNetwork({
        request: event.request,
        preloadResponsePromise: event.preloadResponse,
    }));
});

{{else}}
// If not in dev mode, then we want to cache the static resources
// The cache name is reset on each heedy reboot, which will tell the browser to reinstall the serviceworker
// and therefore refresh the cache whenever plugins or config are changed.
let cache_name = {{.RunID }};

// -----------------------------------------------------
// These are based on https://github.com/mdn/sw-test/blob/gh-pages/sw.js

const putInCache = async (request, response) => {
    const cache = await caches.open(cache_name);
    await cache.put(request, response);
};

const addResourcesToCache = async (resources) => {
    const cache = await caches.open(cache_name);
    await cache.addAll(resources);
};

const getCache = async ({ request, preloadResponsePromise}) => {
    // First try to get the resource from the cache
    const responseFromCache = await caches.match(request);
    if (responseFromCache) {
        console.debug("sw: (cached) ", request.url);
        return responseFromCache;
    }

    // Next try to use the preloaded response, if it's there
    const preloadResponse = await preloadResponsePromise;
    if (preloadResponse) {
        console.debug('sw: (preload->cache) ', request.url);
        putInCache(request, preloadResponse.clone());
        return preloadResponse;
    }

    // Next try to get the resource from the network
    try {
        console.debug('sw: (net->cache)', request.url);
        const responseFromNetwork = await fetch(request);
        // response may be used only once
        // we need to save clone to put one copy in cache
        // and serve second one
        putInCache(request, responseFromNetwork.clone());
        return responseFromNetwork;
    } catch (error) {
        // there is nothing we can do, but we must always
        // return a Response object
        return new Response('Network Error', {
            status: 408,
            headers: { 'Content-Type': 'text/plain' },
        });
    }
};


// -----------------------------------------------------

function getPath(request) {
    let u = new URL(request.url);
    return u.pathname;
}

self.addEventListener('activate', (event) => {
    event.waitUntil(enableNavigationPreload());
});

self.addEventListener("install", function (event) {
    console.debug("sw: installing",cache_name);
    event.waitUntil(
        caches.keys().then(function(keyList) {
            return Promise.all(
              keyList.map(function(key) {
                return caches.delete(key);
              })
            );
          }).then((result) => {
            return addResourcesToCache([{{range .Preload}}
                "/static/{{.}}",{{end}}
                "/favicon.ico",
                "/static/fonts/roboto-latin-400.woff2",
                "/static/fonts/roboto-latin-700.woff2",
                "/static/fonts/roboto-latin-500.woff2",
                "/static/fonts/MaterialIcons-Regular.woff2",
                "/static/fonts/fa-solid-900.woff2",
                "/static/fonts/fa-regular-400.woff2",
                "/manifest.json"
              ])
          }).then(() => console.debug("sw: installed",cache_name)));

    self.skipWaiting();
});

const customCached = ["/manifest.json","/favicon.ico"];

self.addEventListener("fetch", function (event) {
    let path = getPath(event.request);

    if (path.startsWith("/static/") || customCached.includes(path)) {
        // Static resources are always cached
        event.respondWith(getCache({
            request: event.request,
            preloadResponsePromise: event.preloadResponse,
        }));
        return;
    }
    // All others respond from the network

    if (path=="/auth/logout") {
        // Logout is a special case, we need to clear the cache
        console.debug("sw: clear cache");
        caches.delete(cache_name);
        // And then do the raw network request
    }

    event.respondWith(getNetwork({
        request: event.request,
        preloadResponsePromise: event.preloadResponse,
    }));

});
{{end}}