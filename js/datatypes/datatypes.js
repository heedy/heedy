// datatypes references the types of data that can be input/viewed by the frontend.
// An example of an input is the star ratings. When a stream with the star rating datatype
// is queried, the getInput function returns a component where stars can be clicked to input ratings
// In the case of views, given a stream, it returns an array of components: the visualizations/plots/tables
// that will show cool stuff.

// combine takes the given array, and combines it with dots
function combine(arr) {
    let res = "";
    for (let i = 0; i < arr.length; i++) {
        if (i != 0) {
            res += ".";
        }
        res += arr[i];
    }
    return res;
}

// getFromDict is a helper function, used later in the file to correctly separate
// datatypes by dots.
function getFromDict(dict, datatype, plugintype) {
    let datapath = datatype.split(".");
    let currpath = "";
    for (let i = datapath.length; i >= 0; i--) {
        currpath = combine(datapath.slice(0, i));
        if (dict[currpath] !== undefined) {
            // console.log("Using " + plugintype + " plugin '" + currpath + "' for datatype '" + datatype + "'");
            return dict[currpath];
        }

    }
}

// add and get input - there can only be a single input component per stream. The inputs are set up by datatype.
// The search for valid inputs goes by dots - if we have a registered input for rating.stars, and the stream
// has datatype rating.stars.supercool, and supercool is not given, then rating.stars is returned. If no datatype
// can be found, it returns a JSONSchema input component, which will allow directly inputting the data.
var inputdict = {};

export function addInput(datatype, input) {
    inputdict[datatype] = input;
}

export function getInput(datatype) {
    return getFromDict(inputdict, datatype, "input");
}

// add and get creators are similar to inputs, as they are per datatype. The creators are
// a structure which gives the necessary information to create a new stream with the given
// datatype, and includes the components that should be included.

var createdict = {};
export function addCreator(datatype, creator) {
    createdict[datatype] = creator;
}

export function getCreator(datatype) {
    return getFromDict(createdict, datatype, "create");
}

// add and get views for a stream. Each view consists of a function that given a stream either returns a Component
// for the stream, or null (if the component is not built for the given datatype). This allows getViews to return an
// array of components for showing data tables and plots, etc. There will always be at least one view: the data table.
// Other than that, views are dependent on streams.

var viewarray = [];

export function addView(view) {
    viewarray.append(view);
}

export function getViews(payload) {
    var resultarray = [];
    for (let i = 0; i < viewarray.length; i++) {
        let res = viewarray[i](payload);
        if (res !== null) {
            resultarray.append(res);
        }
    }
    return resultarray;
}
