import{g as d,aX as k,ch as M,az as g,ax as O,bb as w,ay as W,d as F,a9 as C,aA as I,r as q,h as V,O as j,aT as x,aB as y,ci as A,aE as D}from"./index-BWGkTP3E.js";function K(n,e){const o=k(M,null);return d(()=>n.hljs||(o==null?void 0:o.mergedHljsRef.value))}const U=g([O("code",`
 font-size: var(--n-font-size);
 font-family: var(--n-font-family);
 `,[w("show-line-numbers",`
 display: flex;
 `),W("line-numbers",`
 user-select: none;
 padding-right: 12px;
 text-align: right;
 transition: color .3s var(--n-bezier);
 color: var(--n-line-number-text-color);
 `),w("word-wrap",[g("pre",`
 white-space: pre-wrap;
 word-break: break-all;
 `)]),g("pre",`
 margin: 0;
 line-height: inherit;
 font-size: inherit;
 font-family: inherit;
 `),g("[class^=hljs]",`
 color: var(--n-text-color);
 transition: 
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 `)]),({props:n})=>{const e=`${n.bPrefix}code`;return[`${e} .hljs-comment,
 ${e} .hljs-quote {
 color: var(--n-mono-3);
 font-style: italic;
 }`,`${e} .hljs-doctag,
 ${e} .hljs-keyword,
 ${e} .hljs-formula {
 color: var(--n-hue-3);
 }`,`${e} .hljs-section,
 ${e} .hljs-name,
 ${e} .hljs-selector-tag,
 ${e} .hljs-deletion,
 ${e} .hljs-subst {
 color: var(--n-hue-5);
 }`,`${e} .hljs-literal {
 color: var(--n-hue-1);
 }`,`${e} .hljs-string,
 ${e} .hljs-regexp,
 ${e} .hljs-addition,
 ${e} .hljs-attribute,
 ${e} .hljs-meta-string {
 color: var(--n-hue-4);
 }`,`${e} .hljs-built_in,
 ${e} .hljs-class .hljs-title {
 color: var(--n-hue-6-2);
 }`,`${e} .hljs-attr,
 ${e} .hljs-variable,
 ${e} .hljs-template-variable,
 ${e} .hljs-type,
 ${e} .hljs-selector-class,
 ${e} .hljs-selector-attr,
 ${e} .hljs-selector-pseudo,
 ${e} .hljs-number {
 color: var(--n-hue-6);
 }`,`${e} .hljs-symbol,
 ${e} .hljs-bullet,
 ${e} .hljs-link,
 ${e} .hljs-meta,
 ${e} .hljs-selector-id,
 ${e} .hljs-title {
 color: var(--n-hue-2);
 }`,`${e} .hljs-emphasis {
 font-style: italic;
 }`,`${e} .hljs-strong {
 font-weight: var(--n-font-weight-strong);
 }`,`${e} .hljs-link {
 text-decoration: underline;
 }`]}]),X=Object.assign(Object.assign({},y.props),{language:String,code:{type:String,default:""},trim:{type:Boolean,default:!0},hljs:Object,uri:Boolean,inline:Boolean,wordWrap:Boolean,showLineNumbers:Boolean,internalFontSize:Number,internalNoHighlight:Boolean}),J=F({name:"Code",props:X,setup(n,{slots:e}){const{internalNoHighlight:o}=n,{mergedClsPrefixRef:m,inlineThemeDisabled:h}=I(),a=q(null),b=o?{value:void 0}:K(n),z=(t,s,l)=>{const{value:r}=b;return!r||!(t&&r.getLanguage(t))?null:r.highlight(l?s.trim():s,{language:t}).value},N=d(()=>n.inline||n.wordWrap?!1:n.showLineNumbers),f=()=>{if(e.default)return;const{value:t}=a;if(!t)return;const{language:s}=n,l=n.uri?window.decodeURIComponent(n.code):n.code;if(s){const i=z(s,l,n.trim);if(i!==null){if(n.inline)t.innerHTML=i;else{const $=t.querySelector(".__code__");$&&t.removeChild($);const u=document.createElement("pre");u.className="__code__",u.innerHTML=i,t.appendChild(u)}return}}if(n.inline){t.textContent=l;return}const r=t.querySelector(".__code__");if(r)r.textContent=l;else{const i=document.createElement("pre");i.className="__code__",i.textContent=l,t.innerHTML="",t.appendChild(i)}};V(f),j(x(n,"language"),f),j(x(n,"code"),f),o||j(b,f);const R=y("Code","-code",U,A,n,m),v=d(()=>{const{common:{cubicBezierEaseInOut:t,fontFamilyMono:s},self:{textColor:l,fontSize:r,fontWeightStrong:i,lineNumberTextColor:$,"mono-3":u,"hue-1":S,"hue-2":p,"hue-3":L,"hue-4":H,"hue-5":B,"hue-5-2":E,"hue-6":P,"hue-6-2":T}}=R.value,{internalFontSize:_}=n;return{"--n-font-size":_?`${_}px`:r,"--n-font-family":s,"--n-font-weight-strong":i,"--n-bezier":t,"--n-text-color":l,"--n-mono-3":u,"--n-hue-1":S,"--n-hue-2":p,"--n-hue-3":L,"--n-hue-4":H,"--n-hue-5":B,"--n-hue-5-2":E,"--n-hue-6":P,"--n-hue-6-2":T,"--n-line-number-text-color":$}}),c=h?D("code",d(()=>`${n.internalFontSize||"a"}`),v,n):void 0;return{mergedClsPrefix:m,codeRef:a,mergedShowLineNumbers:N,lineNumbers:d(()=>{let t=1;const s=[];let l=!1;for(const r of n.code)r===`
`?(l=!0,s.push(t++)):l=!1;return l||s.push(t++),s.join(`
`)}),cssVars:h?void 0:v,themeClass:c==null?void 0:c.themeClass,onRender:c==null?void 0:c.onRender}},render(){var n,e;const{mergedClsPrefix:o,wordWrap:m,mergedShowLineNumbers:h,onRender:a}=this;return a==null||a(),C("code",{class:[`${o}-code`,this.themeClass,m&&`${o}-code--word-wrap`,h&&`${o}-code--show-line-numbers`],style:this.cssVars,ref:"codeRef"},h?C("pre",{class:`${o}-code__line-numbers`},this.lineNumbers):null,(e=(n=this.$slots).default)===null||e===void 0?void 0:e.call(n))}});export{J as _};
