import{ax as o,ay as i,d as f,a9 as l,bG as r,aA as R,aB as d,cj as $,aD as B,g as O,aE as E,c as x,o as h,a as m}from"./index-BWGkTP3E.js";const T=o("statistic",[i("label",`
 font-weight: var(--n-label-font-weight);
 transition: .3s color var(--n-bezier);
 font-size: var(--n-label-font-size);
 color: var(--n-label-text-color);
 `),o("statistic-value",`
 margin-top: 4px;
 font-weight: var(--n-value-font-weight);
 `,[i("prefix",`
 margin: 0 4px 0 0;
 font-size: var(--n-value-font-size);
 transition: .3s color var(--n-bezier);
 color: var(--n-value-prefix-text-color);
 `,[o("icon",{verticalAlign:"-0.125em"})]),i("content",`
 font-size: var(--n-value-font-size);
 transition: .3s color var(--n-bezier);
 color: var(--n-value-text-color);
 `),i("suffix",`
 margin: 0 0 0 4px;
 font-size: var(--n-value-font-size);
 transition: .3s color var(--n-bezier);
 color: var(--n-value-suffix-text-color);
 `,[o("icon",{verticalAlign:"-0.125em"})])])]),j=Object.assign(Object.assign({},d.props),{tabularNums:Boolean,label:String,value:[String,Number]}),y=f({name:"Statistic",props:j,slots:Object,setup(n){const{mergedClsPrefixRef:e,inlineThemeDisabled:s,mergedRtlRef:c}=R(n),u=d("Statistic","-statistic",T,$,n,e),v=B("Statistic",c,e),t=O(()=>{const{self:{labelFontWeight:p,valueFontSize:b,valueFontWeight:g,valuePrefixTextColor:w,labelTextColor:C,valueSuffixTextColor:_,valueTextColor:z,labelFontSize:k},common:{cubicBezierEaseInOut:S}}=u.value;return{"--n-bezier":S,"--n-label-font-size":k,"--n-label-font-weight":p,"--n-label-text-color":C,"--n-value-font-weight":g,"--n-value-font-size":b,"--n-value-prefix-text-color":w,"--n-value-suffix-text-color":_,"--n-value-text-color":z}}),a=s?E("statistic",void 0,t,n):void 0;return{rtlEnabled:v,mergedClsPrefix:e,cssVars:s?void 0:t,themeClass:a==null?void 0:a.themeClass,onRender:a==null?void 0:a.onRender}},render(){var n;const{mergedClsPrefix:e,$slots:{default:s,label:c,prefix:u,suffix:v}}=this;return(n=this.onRender)===null||n===void 0||n.call(this),l("div",{class:[`${e}-statistic`,this.themeClass,this.rtlEnabled&&`${e}-statistic--rtl`],style:this.cssVars},r(c,t=>l("div",{class:`${e}-statistic__label`},this.label||t)),l("div",{class:`${e}-statistic-value`,style:{fontVariantNumeric:this.tabularNums?"tabular-nums":""}},r(u,t=>t&&l("span",{class:`${e}-statistic-value__prefix`},t)),this.value!==void 0?l("span",{class:`${e}-statistic-value__content`},this.value):r(s,t=>t&&l("span",{class:`${e}-statistic-value__content`},t)),r(v,t=>t&&l("span",{class:`${e}-statistic-value__suffix`},t))))}}),N={xmlns:"http://www.w3.org/2000/svg","xmlns:xlink":"http://www.w3.org/1999/xlink",viewBox:"0 0 512 512"},D=f({name:"ChevronDownOutline",render:function(e,s){return h(),x("svg",N,s[0]||(s[0]=[m("path",{fill:"none",stroke:"currentColor","stroke-linecap":"round","stroke-linejoin":"round","stroke-width":"48",d:"M112 184l144 144l144-144"},null,-1)]))}}),P={xmlns:"http://www.w3.org/2000/svg","xmlns:xlink":"http://www.w3.org/1999/xlink",viewBox:"0 0 512 512"},F=f({name:"ChevronUpOutline",render:function(e,s){return h(),x("svg",P,s[0]||(s[0]=[m("path",{fill:"none",stroke:"currentColor","stroke-linecap":"round","stroke-linejoin":"round","stroke-width":"48",d:"M112 328l144-144l144 144"},null,-1)]))}});export{F as C,y as _,D as a};
