/*
WebHistoryView shows a visualization of URLs visited. This component was built
mainly to interact with the chrome extension.
*/

import { addView } from '../datatypes';
import { generateBarChart } from './components/BarChart';
import { generateLineChart } from './components/LineChart';

// http://stackoverflow.com/questions/5717093/check-if-a-javascript-string-is-a-url
var urlchecker = new RegExp('^(https?:\\/\\/)?' + // protocol
    '((([a-z\\d]([a-z\\d-]*[a-z\\d])*)\\.?)+[a-z]{2,}|' + // domain name
    '((\\d{1,3}\\.){3}\\d{1,3}))' + // OR ip (v4) address
    '(\\:\\d+)?(\\/[-a-z\\d%_.~+]*)*' + // port and path
    '(\\?[;&a-z\\d%_.~+=-]*)?' + // query string
    '(\\#[-a-z\\d_]*)?$', 'i'); // fragment locator

var urlView = {
    key: "urlView",
    title: "Most Common Domains",
    subtitle: ""
};

function showHistoryView(context) {
    if (context.data.length < 5 || context.pipescript === null) {
        return null;
    }

    var view = [];
    var d = context.data;

    var hasUrlProperty = true;
    var isUrl = true;
    var hasTitle = true;

    // First, we check if the data is directly of type URL. In this case,
    // we add a bar chart of the URLs. We only check 5 points on each side
    if (d.length < 10) {
        for (let i = 0; i < d.length; i++) {
            if (!("url" in d[i].d)) hasUrlProperty = false;
            if (!("title" in d[i].d)) hasTitle = false;
            if (typeof d[i].d !== "string" || !urlchecker.test(d[i].d)) isUrl = false;
        }
    } else {
        for (let i = 0; i < 5; i++) {
            if (!("url" in d[i].d)) hasUrlProperty = false;
            if (!("title" in d[i].d)) hasTitle = false;
            if (typeof d[i].d !== "string" || !urlchecker.test(d[i].d)) isUrl = false;
        }
        for (let i = d.length - 5; i < d.length; i++) {
            if (!("url" in d[i].d)) hasUrlProperty = false;
            if (!("title" in d[i].d)) hasTitle = false;
            if (typeof d[i].d !== "string" || !urlchecker.test(d[i].d)) isUrl = false;
        }
    }

    if (isUrl) {
        view.push({
            ...generateBarChart("map(domain,count) | top(20)", "Shows which domains were most frequenly visited"),
            ...urlView,
        });
    } else if (hasUrlProperty) {
        view.push({
            ...generateBarChart("$('url') | map(domain,count) | top(20)", "Shows which domains were most frequenly visited"),
            ...urlView,
        });
    }

    if (hasUrlProperty && hasTitle) {
        // If it has the URL property, check if it also has titlebar
        view.push({
            ...generateLineChart("$('title') | sentiment"),
            key: "urlTitleSentiment",
            title: "Website Titlebar Sentiment",
            subtitle: ""
        });
    }
    return view;
}


addView(showHistoryView);