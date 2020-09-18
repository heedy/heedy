import moment from "../../../dist/moment.mjs";

function preprocessor(qd, visualization) {
  return {
    ...visualization,
    config: visualization.config.map((c, i) => ({
      ...c,
      data: qd.dataset[i].map((dp) => {
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
            obj[k] =
              typeof dp.d[k] !== "string" ? JSON.stringify(dp.d[k]) : dp.d[k];
          });
        }
        return obj;
      }),
    })),
  };
}

export default preprocessor;
