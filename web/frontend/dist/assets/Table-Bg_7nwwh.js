import{at as r,aq as n,bp as O,bq as V,as as t,au as j,d as q,a8 as F,aA as A,aC as b,cG as D,aB as G,g as c,aF as a,aH as I}from"./index-Bm8WTrgo.js";const K=r([n("table",`
 font-size: var(--n-font-size);
 font-variant-numeric: tabular-nums;
 line-height: var(--n-line-height);
 width: 100%;
 border-radius: var(--n-border-radius) var(--n-border-radius) 0 0;
 text-align: left;
 border-collapse: separate;
 border-spacing: 0;
 overflow: hidden;
 background-color: var(--n-td-color);
 border-color: var(--n-merged-border-color);
 transition:
 background-color .3s var(--n-bezier),
 border-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 --n-merged-border-color: var(--n-border-color);
 `,[r("th",`
 white-space: nowrap;
 transition:
 background-color .3s var(--n-bezier),
 border-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 text-align: inherit;
 padding: var(--n-th-padding);
 vertical-align: inherit;
 text-transform: none;
 border: 0px solid var(--n-merged-border-color);
 font-weight: var(--n-th-font-weight);
 color: var(--n-th-text-color);
 background-color: var(--n-th-color);
 border-bottom: 1px solid var(--n-merged-border-color);
 border-right: 1px solid var(--n-merged-border-color);
 `,[r("&:last-child",`
 border-right: 0px solid var(--n-merged-border-color);
 `)]),r("td",`
 transition:
 background-color .3s var(--n-bezier),
 border-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 padding: var(--n-td-padding);
 color: var(--n-td-text-color);
 background-color: var(--n-td-color);
 border: 0px solid var(--n-merged-border-color);
 border-right: 1px solid var(--n-merged-border-color);
 border-bottom: 1px solid var(--n-merged-border-color);
 `,[r("&:last-child",`
 border-right: 0px solid var(--n-merged-border-color);
 `)]),t("bordered",`
 border: 1px solid var(--n-merged-border-color);
 border-radius: var(--n-border-radius);
 `,[r("tr",[r("&:last-child",[r("td",`
 border-bottom: 0 solid var(--n-merged-border-color);
 `)])])]),t("single-line",[r("th",`
 border-right: 0px solid var(--n-merged-border-color);
 `),r("td",`
 border-right: 0px solid var(--n-merged-border-color);
 `)]),t("single-column",[r("tr",[r("&:not(:last-child)",[r("td",`
 border-bottom: 0px solid var(--n-merged-border-color);
 `)])])]),t("striped",[r("tr:nth-of-type(even)",[r("td","background-color: var(--n-td-color-striped)")])]),j("bottom-bordered",[r("tr",[r("&:last-child",[r("td",`
 border-bottom: 0px solid var(--n-merged-border-color);
 `)])])])]),O(n("table",`
 background-color: var(--n-td-color-modal);
 --n-merged-border-color: var(--n-border-color-modal);
 `,[r("th",`
 background-color: var(--n-th-color-modal);
 `),r("td",`
 background-color: var(--n-td-color-modal);
 `)])),V(n("table",`
 background-color: var(--n-td-color-popover);
 --n-merged-border-color: var(--n-border-color-popover);
 `,[r("th",`
 background-color: var(--n-th-color-popover);
 `),r("td",`
 background-color: var(--n-td-color-popover);
 `)]))]),N=Object.assign(Object.assign({},b.props),{bordered:{type:Boolean,default:!0},bottomBordered:{type:Boolean,default:!0},singleLine:{type:Boolean,default:!0},striped:Boolean,singleColumn:Boolean,size:{type:String,default:"medium"}}),J=q({name:"Table",props:N,setup(e){const{mergedClsPrefixRef:o,inlineThemeDisabled:i,mergedRtlRef:g}=A(e),h=b("Table","-table",K,D,e,o),v=G("Table",g,o),s=c(()=>{const{size:l}=e,{self:{borderColor:p,tdColor:m,tdColorModal:u,tdColorPopover:f,thColor:C,thColorModal:x,thColorPopover:z,thTextColor:k,tdTextColor:P,borderRadius:R,thFontWeight:B,lineHeight:y,borderColorModal:T,borderColorPopover:$,tdColorStriped:w,tdColorStripedModal:M,tdColorStripedPopover:S,[a("fontSize",l)]:_,[a("tdPadding",l)]:E,[a("thPadding",l)]:H},common:{cubicBezierEaseInOut:L}}=h.value;return{"--n-bezier":L,"--n-td-color":m,"--n-td-color-modal":u,"--n-td-color-popover":f,"--n-td-text-color":P,"--n-border-color":p,"--n-border-color-modal":T,"--n-border-color-popover":$,"--n-border-radius":R,"--n-font-size":_,"--n-th-color":C,"--n-th-color-modal":x,"--n-th-color-popover":z,"--n-th-font-weight":B,"--n-th-text-color":k,"--n-line-height":y,"--n-td-padding":E,"--n-th-padding":H,"--n-td-color-striped":w,"--n-td-color-striped-modal":M,"--n-td-color-striped-popover":S}}),d=i?I("table",c(()=>e.size[0]),s,e):void 0;return{rtlEnabled:v,mergedClsPrefix:o,cssVars:i?void 0:s,themeClass:d==null?void 0:d.themeClass,onRender:d==null?void 0:d.onRender}},render(){var e;const{mergedClsPrefix:o}=this;return(e=this.onRender)===null||e===void 0||e.call(this),F("table",{class:[`${o}-table`,this.themeClass,{[`${o}-table--rtl`]:this.rtlEnabled,[`${o}-table--bottom-bordered`]:this.bottomBordered,[`${o}-table--bordered`]:this.bordered,[`${o}-table--single-line`]:this.singleLine,[`${o}-table--single-column`]:this.singleColumn,[`${o}-table--striped`]:this.striped}],style:this.cssVars},this.$slots)}});export{J as _};
