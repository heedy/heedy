// The Heedy serviceworker resets the cache on every server reboot, unless in dev mode.
// based on https://github.com/mdn/sw-test/blob/gh-pages/sw.js

const enableNavigationPreload = async () => {
    if (self.registration.navigationPreload) {
        await self.registration.navigationPreload.enable();
    }
};

const r = (old) => new Request(old, {cache: 'no-cache'});
const errResponse = (err) => new Response('{"error":"fetch_error","error_description":"Could not connect to the server","id":"?"}', {status: 408,headers: { 'Content-Type': 'application/json; charset=utf-8' }});

const getNetwork = async({ request, preloadResponsePromise}) => {
    const preloadResponse = await preloadResponsePromise;
    if (preloadResponse) return preloadResponse;
    try {
        return await fetch(r(request));
    } catch (error) {
        return errResponse(error);
    }
}

self.addEventListener('activate', (event) => {
    event.waitUntil(enableNavigationPreload());
});

{{if .DevMode}}
// If we are running in dev mode, we don't cache any results, so that every refresh
// always fetches from the server

self.addEventListener("install", function (event) {
    console.debug("sw: installing (dev mode)");
    event.waitUntil(
        caches.keys().then(function(keyList) {
            return Promise.all(
              keyList.map(function(key) {
                return caches.delete(key);
              })
            );
          }));

    self.skipWaiting();
});

self.addEventListener("fetch", function (event) {
    event.respondWith(getNetwork({
        request: event.request,
        preloadResponsePromise: event.preloadResponse,
    }));
});

{{else}}
// The cache name is reset on each heedy reboot, which will tell the browser to reinstall the serviceworker
// and therefore refresh the cache whenever plugins or config are changed.
let cache_name = {{ .RunID }};

const putInCache = async (request, response) => {
    const cache = await caches.open(cache_name);
    await cache.put(request, response);
};

const addResourcesToCache = async (resources) => {
    const cache = await caches.open(cache_name);
    await cache.addAll(resources);
};

const getCache = async ({ request, preloadResponsePromise}) => {
    const responseFromCache = await caches.match(request);
    if (responseFromCache) return responseFromCache;
    
    const preloadResponse = await preloadResponsePromise;
    if (preloadResponse) {
        putInCache(request, preloadResponse.clone());
        return preloadResponse;
    }

    try {
        const responseFromNetwork = await fetch(r(request));
        putInCache(request, responseFromNetwork.clone());
        return responseFromNetwork;
    } catch (error) {
        return errResponse(error);
    }
};

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
                r("/static/{{.}}"),{{end}}
                r("/favicon.ico"),
                r("/static/fonts/roboto-latin-400.woff2"),
                r("/static/fonts/roboto-latin-700.woff2"),
                r("/static/fonts/roboto-latin-500.woff2"),
                r("/static/fonts/MaterialIcons-Regular.woff2"),
                r("/static/fonts/fa-solid-900.woff2"),
                r("/static/fonts/fa-regular-400.woff2"),
                r("/manifest.json")
              ])
          }));

    self.skipWaiting();
});

const customCached = ["/manifest.json","/favicon.ico"];

self.addEventListener("fetch", function (event) {
    let path = (new URL(event.request.url)).pathname;

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
        // Logout is a special case, we clear the cache on logout
        caches.delete(cache_name);
    }

    event.respondWith(getNetwork({
        request: event.request,
        preloadResponsePromise: event.preloadResponse,
    }));
});
{{end}}