
export const setPart1 = (s) => ({ type: "UPLOADER_PART1", value: s });
export const setPart2 = (s) => ({ type: "UPLOADER_PART2", value: s });
export const setPart3 = (s) => ({ type: "UPLOADER_PART3", value: s });
export const setState = (s) => ({ type: "UPLOADER_SET", value: s });

export const process = () => ({ type: "UPLOADER_PROCESS" });
export const upload = () => ({ type: "UPLOADER_UPLOAD" });