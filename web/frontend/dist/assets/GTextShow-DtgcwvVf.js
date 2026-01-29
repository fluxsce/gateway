import{ak as te,j as a,bI as ne,O as se,S as T,P as oe,Q as J,R as le,d as Z,T as V,V as ae,q as re,a7 as R,W as Y,Z as ie,r as I,a3 as D,m as ue,cR as z,c as N,o as x,a5 as G,n as ce,h as w,f as L,b as C,g as E,e as g,aO as X,w as b,K as H,t as B,i as K,J as U,_ as he}from"./index-DFGZIhYV.js";import{C as de}from"./CopyOutline-r53JwZwj.js";import{C as me}from"./CodeOutline-B_7H-bTU.js";import"./SearchForm-DCaazdSn.js";function fe(o,t){const n=te(ne,null);return a(()=>o.hljs||(n==null?void 0:n.mergedHljsRef.value))}function ge(o){const{textColor2:t,fontSize:n,fontWeightStrong:v,textColor3:u}=o;return{textColor:t,fontSize:n,fontWeightStrong:v,"mono-3":"#a0a1a7","hue-1":"#0184bb","hue-2":"#4078f2","hue-3":"#a626a4","hue-4":"#50a14f","hue-5":"#e45649","hue-5-2":"#c91243","hue-6":"#986801","hue-6-2":"#c18401",lineNumberTextColor:u}}const ve={common:se,self:ge},pe=T([oe("code",`
 font-size: var(--n-font-size);
 font-family: var(--n-font-family);
 `,[J("show-line-numbers",`
 display: flex;
 `),le("line-numbers",`
 user-select: none;
 padding-right: 12px;
 text-align: right;
 transition: color .3s var(--n-bezier);
 color: var(--n-line-number-text-color);
 `),J("word-wrap",[T("pre",`
 white-space: pre-wrap;
 word-break: break-all;
 `)]),T("pre",`
 margin: 0;
 line-height: inherit;
 font-size: inherit;
 font-family: inherit;
 `),T("[class^=hljs]",`
 color: var(--n-text-color);
 transition: 
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 `)]),({props:o})=>{const t=`${o.bPrefix}code`;return[`${t} .hljs-comment,
 ${t} .hljs-quote {
 color: var(--n-mono-3);
 font-style: italic;
 }`,`${t} .hljs-doctag,
 ${t} .hljs-keyword,
 ${t} .hljs-formula {
 color: var(--n-hue-3);
 }`,`${t} .hljs-section,
 ${t} .hljs-name,
 ${t} .hljs-selector-tag,
 ${t} .hljs-deletion,
 ${t} .hljs-subst {
 color: var(--n-hue-5);
 }`,`${t} .hljs-literal {
 color: var(--n-hue-1);
 }`,`${t} .hljs-string,
 ${t} .hljs-regexp,
 ${t} .hljs-addition,
 ${t} .hljs-attribute,
 ${t} .hljs-meta-string {
 color: var(--n-hue-4);
 }`,`${t} .hljs-built_in,
 ${t} .hljs-class .hljs-title {
 color: var(--n-hue-6-2);
 }`,`${t} .hljs-attr,
 ${t} .hljs-variable,
 ${t} .hljs-template-variable,
 ${t} .hljs-type,
 ${t} .hljs-selector-class,
 ${t} .hljs-selector-attr,
 ${t} .hljs-selector-pseudo,
 ${t} .hljs-number {
 color: var(--n-hue-6);
 }`,`${t} .hljs-symbol,
 ${t} .hljs-bullet,
 ${t} .hljs-link,
 ${t} .hljs-meta,
 ${t} .hljs-selector-id,
 ${t} .hljs-title {
 color: var(--n-hue-2);
 }`,`${t} .hljs-emphasis {
 font-style: italic;
 }`,`${t} .hljs-strong {
 font-weight: var(--n-font-weight-strong);
 }`,`${t} .hljs-link {
 text-decoration: underline;
 }`]}]),ye=Object.assign(Object.assign({},Y.props),{language:String,code:{type:String,default:""},trim:{type:Boolean,default:!0},hljs:Object,uri:Boolean,inline:Boolean,wordWrap:Boolean,showLineNumbers:Boolean,internalFontSize:Number,internalNoHighlight:Boolean}),xe=Z({name:"Code",props:ye,setup(o,{slots:t}){const{internalNoHighlight:n}=o,{mergedClsPrefixRef:v,inlineThemeDisabled:u}=ae(),m=I(null),c=n?{value:void 0}:fe(o),M=(l,h,r)=>{const{value:d}=c;return!d||!(l&&d.getLanguage(l))?null:d.highlight(r?h.trim():h,{language:l}).value},$=a(()=>o.inline||o.wordWrap?!1:o.showLineNumbers),j=()=>{if(t.default)return;const{value:l}=m;if(!l)return;const{language:h}=o,r=o.uri?window.decodeURIComponent(o.code):o.code;if(h){const f=M(h,r,o.trim);if(f!==null){if(o.inline)l.innerHTML=f;else{const _=l.querySelector(".__code__");_&&l.removeChild(_);const p=document.createElement("pre");p.className="__code__",p.innerHTML=f,l.appendChild(p)}return}}if(o.inline){l.textContent=r;return}const d=l.querySelector(".__code__");if(d)d.textContent=r;else{const f=document.createElement("pre");f.className="__code__",f.textContent=r,l.innerHTML="",l.appendChild(f)}};re(j),R(D(o,"language"),j),R(D(o,"code"),j),n||R(c,j);const k=Y("Code","-code",pe,ve,o,v),S=a(()=>{const{common:{cubicBezierEaseInOut:l,fontFamilyMono:h},self:{textColor:r,fontSize:d,fontWeightStrong:f,lineNumberTextColor:_,"mono-3":p,"hue-1":F,"hue-2":O,"hue-3":W,"hue-4":P,"hue-5":e,"hue-5-2":s,"hue-6":y,"hue-6-2":ee}}=k.value,{internalFontSize:A}=o;return{"--n-font-size":A?`${A}px`:d,"--n-font-family":h,"--n-font-weight-strong":f,"--n-bezier":l,"--n-text-color":r,"--n-mono-3":p,"--n-hue-1":F,"--n-hue-2":O,"--n-hue-3":W,"--n-hue-4":P,"--n-hue-5":e,"--n-hue-5-2":s,"--n-hue-6":y,"--n-hue-6-2":ee,"--n-line-number-text-color":_}}),i=u?ie("code",a(()=>`${o.internalFontSize||"a"}`),S,o):void 0;return{mergedClsPrefix:v,codeRef:m,mergedShowLineNumbers:$,lineNumbers:a(()=>{let l=1;const h=[];let r=!1;for(const d of o.code)d===`
`?(r=!0,h.push(l++)):r=!1;return r||h.push(l++),h.join(`
`)}),cssVars:u?void 0:S,themeClass:i==null?void 0:i.themeClass,onRender:i==null?void 0:i.onRender}},render(){var o,t;const{mergedClsPrefix:n,wordWrap:v,mergedShowLineNumbers:u,onRender:m}=this;return m==null||m(),V("code",{class:[`${n}-code`,this.themeClass,v&&`${n}-code--word-wrap`,u&&`${n}-code--show-line-numbers`],style:this.cssVars,ref:"codeRef"},u?V("pre",{class:`${n}-code__line-numbers`},this.lineNumbers):null,(t=(o=this.$slots).default)===null||t===void 0?void 0:t.call(o))}}),be={key:0,class:"g-text-show__toolbar"},je={class:"g-text-show__toolbar-left"},_e={key:1,class:"performance-tip"},$e={class:"g-text-show__toolbar-right"},we={key:1,class:"g-text-show__plain-text"},Q=500*1024,q=2*1024*1024,Ce=Z({name:"GTextShow",__name:"GTextShow",props:{content:{default:""},format:{default:"auto"},showLineNumbers:{type:Boolean,default:!1},showCopyButton:{type:Boolean,default:!0},autoFormat:{type:Boolean,default:!0},maxHeight:{},minHeight:{},class:{},style:{}},emits:["copy"],setup(o,{emit:t}){const n=o,v=t,u=ue(),m=I(!0),c=I(null),M={highlight:z.highlight.bind(z),getLanguage:z.getLanguage.bind(z)},$=a(()=>n.content?new Blob([n.content]).size:0),j=a(()=>$.value>Q),k=a(()=>{const e=$.value;return e<1024?`${e}B`:e<1024*1024?`${(e/1024).toFixed(2)}KB`:`${(e/1024/1024).toFixed(2)}MB`}),S=e=>{if(!e||!e.trim())return"txt";const s=e.trim();if(new Blob([s]).size>q)return s.startsWith("{")||s.startsWith("[")?"json":s.startsWith("<?xml")||s.startsWith("<")?s.includes("soap:Envelope")||s.includes("soapenv:Envelope")?"soap":"xml":"txt";if(s.startsWith("{")||s.startsWith("["))try{return JSON.parse(s),"json"}catch{}return s.startsWith("<?xml")||s.startsWith("<")?s.includes("soap:Envelope")||s.includes("soapenv:Envelope")?"soap":"xml":s.includes("---")||s.includes(":")&&s.split(`
`).length>1?"yaml":/^\s*(SELECT|INSERT|UPDATE|DELETE|CREATE|ALTER|DROP)\s+/i.test(s)?"sql":/<html[\s>]|<body[\s>]|<div[\s>]/i.test(s)?"html":s.includes("function")||s.includes("const ")||s.includes("let ")?s.includes(":")&&s.includes("interface")||s.includes("type ")?"typescript":"javascript":s.includes("{")&&s.includes("}")&&s.includes(":")?"css":"txt"},i=a(()=>n.format==="auto"?S(n.content):n.format),l=a(()=>{const e=i.value;return{json:"json",xml:"xml",soap:"xml",txt:"plaintext",yaml:"yaml",sql:"sql",javascript:"javascript",typescript:"typescript",css:"css",html:"html",auto:"plaintext"}[e]||"plaintext"}),h=e=>{if(new Blob([e]).size>q)return e;try{const y=JSON.parse(e);return JSON.stringify(y,null,2)}catch{return e}},r=a(()=>{if(!n.content)return"";const e=i.value,s=$.value,y=c.value!==null?c.value:n.autoFormat;return s>q&&!y?n.content:e==="json"&&y?h(n.content):n.content});R(()=>n.content,e=>{if(!e){m.value=!0,c.value=null;return}const s=new Blob([e]).size;m.value=s<=Q,c.value=null},{immediate:!0});const d=a(()=>{const e=i.value;return{json:"success",xml:"info",soap:"primary",txt:"default",yaml:"warning",sql:"info",javascript:"success",typescript:"primary",css:"info",html:"info",auto:"default"}[e]||"default"}),f=a(()=>{const e=i.value;return{json:"JSON",xml:"XML",soap:"SOAP",txt:"TXT",yaml:"YAML",sql:"SQL",javascript:"JavaScript",typescript:"TypeScript",css:"CSS",html:"HTML",auto:"AUTO"}[e]||"TXT"}),_=a(()=>n.showCopyButton||p.value&&!n.autoFormat),p=a(()=>i.value==="json"||i.value==="xml"||i.value==="soap"),F=a(()=>{const e={};if(n.style){if(typeof n.style=="string")return n.style;Object.assign(e,n.style)}return e}),O=a(()=>{const e={};return n.maxHeight&&(e.maxHeight=typeof n.maxHeight=="number"?`${n.maxHeight}px`:n.maxHeight,e.overflow="auto"),n.minHeight&&(e.minHeight=typeof n.minHeight=="number"?`${n.minHeight}px`:n.minHeight),e}),W=async()=>{try{const e=r.value||n.content;await navigator.clipboard.writeText(e),u.success("复制成功"),v("copy",e)}catch(e){u.error("复制失败"),console.error("复制失败:",e)}},P=()=>{const e=i.value;if(e!=="json"&&e!=="xml"&&e!=="soap"){u.warning("当前格式不支持格式化");return}const s=c.value!==null?c.value:n.autoFormat;c.value=!s,c.value?u.success("已格式化"):u.info("已取消格式化")};return(e,s)=>(x(),N("div",{class:ce(["g-text-show",n.class]),style:G(F.value)},[_.value?(x(),N("div",be,[L("div",je,[C(g(X),{size:"small",type:d.value},{default:b(()=>[H(B(f.value),1)]),_:1},8,["type"]),j.value?(x(),E(g(X),{key:0,size:"small",type:"warning"},{default:b(()=>[H(" 超大内容（"+B(k.value)+"） ",1)]),_:1})):w("",!0),j.value&&!m.value?(x(),N("span",_e," 已禁用语法高亮以提升性能 ")):w("",!0)]),L("div",$e,[o.showCopyButton?(x(),E(g(U),{key:0,size:"small",quaternary:"",onClick:W},{icon:b(()=>[C(g(K),null,{default:b(()=>[C(g(de))]),_:1})]),default:b(()=>[s[0]||(s[0]=H(" 复制 ",-1))]),_:1})):w("",!0),p.value?(x(),E(g(U),{key:1,size:"small",quaternary:"",onClick:P},{icon:b(()=>[C(g(K),null,{default:b(()=>[C(g(me))]),_:1})]),default:b(()=>[H(" "+B((c.value!==null?c.value:n.autoFormat)?"取消格式化":"格式化"),1)]),_:1})):w("",!0)])])):w("",!0),L("div",{class:"g-text-show__content",style:G(O.value)},[m.value?(x(),E(g(xe),{key:0,code:r.value,language:l.value,"show-line-numbers":o.showLineNumbers,hljs:M,class:"g-text-show__code"},null,8,["code","language","show-line-numbers"])):(x(),N("pre",we,[L("code",null,B(r.value),1)]))],4)],6))}}),Le=he(Ce,[["__scopeId","data-v-5cd1a2cd"]]);export{Le as G};
