/*
Copyright (c) 2008-2015 Pivotal Labs

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/
function getJasmineRequireObj(){return typeof module!="undefined"&&module.exports?exports:(window.jasmineRequire=window.jasmineRequire||{},window.jasmineRequire)}getJasmineRequireObj().console=function(n,t){t.ConsoleReporter=n.ConsoleReporter()};getJasmineRequireObj().ConsoleReporter=function(){function t(t){function r(){i("\n")}function o(n,t){return y?a[n]+t+a.none:t}function s(n,t){return t==1?n:n+"s"}function w(n,t){for(var i=[],r=0;r<t;r++)i.push(n);return i}function v(n,t){for(var r=(n||"").split("\n"),u=[],i=0;i<r.length;i++)u.push(w(" ",t).join("")+r[i]);return u.join("\n")}function b(n){var t,u;for(r(),i(n.fullName),t=0;t<n.failedExpectations.length;t++)u=n.failedExpectations[t],r(),i(v(u.message,2)),i(v(u.stack,2));r()}function k(n){for(var t=0;t<n.failedExpectations.length;t++)r(),i(o("red","An error was thrown in an afterAll")),r(),i(o("red","AfterAll "+n.failedExpectations[t].message));r()}var i=t.print,y=t.showColors||!1,p=t.onComplete||function(){},l=t.timer||n,f,u,h=[],e,a={green:"\x1b[32m",red:"\x1b[31m",yellow:"\x1b[33m",none:"\x1b[0m"},c=[];return i("ConsoleReporter is deprecated and will be removed in a future version."),this.jasmineStarted=function(){f=0;u=0;e=0;i("Started");r();l.start()},this.jasmineDone=function(){var n,t,o;for(r(),n=0;n<h.length;n++)b(h[n]);for(f>0?(r(),t=f+" "+s("spec",f)+", "+u+" "+s("failure",u),e&&(t+=", "+e+" pending "+s("spec",e)),i(t)):i("No specs found"),r(),o=l.elapsed()/1e3,i("Finished in "+o+" "+s("second",o)),r(),n=0;n<c.length;n++)k(c[n]);p(u===0)},this.specDone=function(n){if(f++,n.status=="pending"){e++;i(o("yellow","*"));return}if(n.status=="passed"){i(o("green","."));return}n.status=="failed"&&(u++,h.push(n),i(o("red","F")))},this.suiteDone=function(n){n.failedExpectations&&n.failedExpectations.length>0&&(u++,c.push(n))},this}var n={start:function(){},elapsed:function(){return 0}};return t};