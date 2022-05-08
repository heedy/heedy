function preprocessor(qd, visualization) {

  return {
    ...visualization,
    data: visualization.config.map((c, i) => {
      const darray = qd.dataset_array[i];
      return { ...c, data: darray };
    }),
  };
}

export default preprocessor;
