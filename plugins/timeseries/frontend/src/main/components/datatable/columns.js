import JSONCell from "./json.vue";
import TimestampCell from "./timestamp.vue";
import NumberCell from "./number.vue";
import StringCell from "./string.vue";
import BooleanCell from "./boolean.vue";
import DurationCell from "./duration.vue";

export const columnTypes = {
    timestamp: {
        component: TimestampCell,
        width: 180,
    },
    duration: {
        component: DurationCell,
        width: 80,
    },
    number: {
        component: NumberCell,
        width: 100,
    },
    string: {
        component: StringCell,
        width: 150,
    },
    boolean: {
        component: BooleanCell,
        width: 8 * 5,
    },
    enum: {
        component: StringCell,
        width: 100,
    },
    json: {
        component: JSONCell,
        width: 250,
    },
};

export default function (t) {
    return columnTypes[t] || columnTypes.json;
}