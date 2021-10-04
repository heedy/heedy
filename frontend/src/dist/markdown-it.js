import MarkdownIt from "markdown-it";

import mila from "markdown-it-link-attributes";

import MdiHljs from "markdown-it-highlightjs";
import hljs from "highlight.js/lib/core"
import js_lang from "highlight.js/lib/languages/javascript";
import py_lang from "highlight.js/lib/languages/python";
import bash_lang from "highlight.js/lib/languages/bash";
import 'highlight.js/styles/github.css';

//import texmath from "markdown-it-texmath";
//import katex from "katex";
//import 'markdown-it-texmath/css/texmath.css';
//import 'katex/dist/katex.min.css';


hljs.registerLanguage(
  'javascript', js_lang
);
hljs.registerLanguage(
  'python', py_lang
);
hljs.registerLanguage(
  'bash', bash_lang
);

let md = new MarkdownIt({
  html: false,
}).use(MdiHljs, { hljs }).use(mila, {
  attrs: {
    target: '_blank',
    rel: 'noopener'
  }
}); //.use(texmath, { engine: katex, delimeters: 'dollars' })

window.markdownit = MarkdownIt;

export { md, hljs, MdiHljs } //, texmath, katex };
export default MarkdownIt;
