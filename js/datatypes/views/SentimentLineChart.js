/*
This shows a line chart of the sentiment in text!

*/

import {addView} from '../datatypes';
import {generateLineChart} from './components/LineChart';
import {generateDropdownLineChart, generateTimeOptions} from './components/DropdownLineChart';
import dropdownTransformDisplay from './components/dropdownTransformDisplay';

// This makes sure that the string includes a space (don't want to match urls or identifiers)
function hasSentence(data) {
    return (typeof data) === "string" && data.includes(" ")
}

const SentimentView = [
    {
        ...generateLineChart("sentiment"),
        key: "sentimentView",
        title: "Plot of Sentiment",
        subtitle: "",
        dropdown: dropdownTransformDisplay("Sentiment is calculated by counting the words with happy/sad connotations using the AFINN-111 wordlist. Sentences deemed positive will have sentiment > 0, while negative sentences will generally have sentiment < 0.", "sentiment")
    }, {
        ...generateDropdownLineChart("This view averages the sentiment values over the chosen time period.", generateTimeOptions("Average", "sentiment", "mean"), 1),
        key: "averagedSentimentView",
        title: "Averaged Sentiment",
        subtitle: ""
    }
];

function showSentimentChart(context) {
    if (context.data.length > 1) {

        // We now check if the data is a string - and has spaces in it
        if (hasSentence(context.data[0].d) && hasSentence(context.data[context.data.length - 1].d)) {
            return SentimentView;
        }

    }

    return null;
}

addView(showSentimentChart);
