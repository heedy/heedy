import moment from "../../../dist/moment.mjs";
import { transform } from "../../analysis.mjs";
function preprocessor(qd, visualization) {
  return {
    ...visualization,
    visualization: "table",
    config: visualization.config.map((c, i) => {
      let darray = qd.dataset_array[i];
      if (c.transform !== undefined) {
        darray = transform(darray, c.transform);
      }
      return ({
        ...c,
        data: darray.map((dp) => {
          let obj = {
            t: new Date(dp.t * 1000).toLocaleString(),
            d: JSON.stringify(dp.d),
            dt:
              dp.dt === undefined
                ? ""
                : moment.duration(dp.dt, "seconds").humanize(),
          };
          if (typeof dp.d == "object") {
            Object.keys(dp.d).map((k) => {
              obj["d_" + k] =
                typeof dp.d[k] !== "string" ? JSON.stringify(dp.d[k]) : dp.d[k];
            });
          }
          return obj;
        }),
      });
    }),
  };
}

export default preprocessor;
