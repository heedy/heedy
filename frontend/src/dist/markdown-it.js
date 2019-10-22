import MarkdownIt from 'markdown-it';

let md = new MarkdownIt({
    html: false
});

export {
    md
};
export default MarkdownIt;