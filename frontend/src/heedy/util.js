var cssLinks = {};
var jsScripts = {};

export function addCSS(linkurl, integrity = "", crossorigin = "anonymous") {
  if (linkurl in cssLinks) return;
  cssLinks[linkurl] = true;
  var link = document.createElement("link");
  link.setAttribute("rel", "stylesheet");
  link.setAttribute("type", "text/css");
  link.setAttribute("href", linkurl);
  if (integrity != "") {
    link.setAttribute("integrity", integrity);
  }
  link.setAttribute("crossorigin", crossorigin);
  document.getElementsByTagName("head")[0].appendChild(link);
}

export function addJS(srcurl, integrity = "", crossorigin = "anonymous") {
  if (srcurl in jsScripts) return;
  jsScripts[srcurl] = true;
  var link = document.createElement("script");
  link.setAttribute("type", "application/javascript");
  link.setAttribute("src", srcurl);
  if (integrity != "") {
    link.setAttribute("integrity", integrity);
  }
  link.setAttribute("crossorigin", crossorigin);
  document.getElementsByTagName("head")[0].appendChild(link);
}
