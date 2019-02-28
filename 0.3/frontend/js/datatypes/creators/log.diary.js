import { addCreator } from "../datatypes";

export const diarySchema = {
  type: "string",
  minLength: 1
};

addCreator("log.diary", {
  name: "diary",
  required: null,
  optional: null,
  description: "A log (or diary) can be used to write about events in your life. Analysis of the text might reveal general trends in your thoughts or what events are associated with certain ratings.",
  default: {
    schema: JSON.stringify(diarySchema),
    datatype: "log.diary",
    icon: "material:library_books",
    ephemeral: false,
    downlink: false
  }
});
