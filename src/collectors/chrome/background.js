
function LogDatapoint(pgurl) {
    if (pgurl != "chrome://newtab/") {
        console.log("PageActive:",pgurl)
    }
}

chrome.tabs.onUpdated.addListener(function (tabId, changeInfo, tab) {
    if (changeInfo.url !=null ) {
        LogDatapoint(tab.url)
    }
    
});

chrome.tabs.onActivated.addListener(function (activeInfo) {
    chrome.tabs.get(activeInfo.tabId, function (tab) {
        LogDatapoint(tab.url)
    });
});

chrome.windows.onFocusChanged.addListener(function (windowId) {
    if (windowId == chrome.windows.WINDOW_ID_NONE) {
        LogDatapoint("")
    } else {
        chrome.tabs.query({ "active": true, "windowId": windowId }, function (tabarr) {
            if (tabarr.length==1) {
                LogDatapoint(tabarr[0].url)
            }
        });
    }

})