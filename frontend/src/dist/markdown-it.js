import MarkdownIt from "markdown-it";

let md = new MarkdownIt({
  html: false,
});

window.markdownit = MarkdownIt;

export { md };
export default MarkdownIt;
