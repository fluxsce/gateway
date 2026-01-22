import{P as B,Q as h,R as t,aA as N,S as b,ak as M,b0 as O,r as p,Y as H,ai as D,V as _,a1 as $,bm as K,a3 as W,d as G,T as v,aC as L,W as F,bQ as Q,aF as Y,Z as J,j as P,aJ as I}from"./index-B5iwLEhr.js";const Z=B("radio",`
 line-height: var(--n-label-line-height);
 outline: none;
 position: relative;
 user-select: none;
 -webkit-user-select: none;
 display: inline-flex;
 align-items: flex-start;
 flex-wrap: nowrap;
 font-size: var(--n-font-size);
 word-break: break-word;
`,[h("checked",[t("dot",`
 background-color: var(--n-color-active);
 `)]),t("dot-wrapper",`
 position: relative;
 flex-shrink: 0;
 flex-grow: 0;
 width: var(--n-radio-size);
 `),B("radio-input",`
 position: absolute;
 border: 0;
 width: 0;
 height: 0;
 opacity: 0;
 margin: 0;
 `),t("dot",`
 position: absolute;
 top: 50%;
 left: 0;
 transform: translateY(-50%);
 height: var(--n-radio-size);
 width: var(--n-radio-size);
 background: var(--n-color);
 box-shadow: var(--n-box-shadow);
 border-radius: 50%;
 transition:
 background-color .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier);
 `,[b("&::before",`
 content: "";
 opacity: 0;
 position: absolute;
 left: 4px;
 top: 4px;
 height: calc(100% - 8px);
 width: calc(100% - 8px);
 border-radius: 50%;
 transform: scale(.8);
 background: var(--n-dot-color-active);
 transition: 
 opacity .3s var(--n-bezier),
 background-color .3s var(--n-bezier),
 transform .3s var(--n-bezier);
 `),h("checked",{boxShadow:"var(--n-box-shadow-active)"},[b("&::before",`
 opacity: 1;
 transform: scale(1);
 `)])]),t("label",`
 color: var(--n-text-color);
 padding: var(--n-label-padding);
 font-weight: var(--n-label-font-weight);
 display: inline-block;
 transition: color .3s var(--n-bezier);
 `),N("disabled",`
 cursor: pointer;
 `,[b("&:hover",[t("dot",{boxShadow:"var(--n-box-shadow-hover)"})]),h("focus",[b("&:not(:active)",[t("dot",{boxShadow:"var(--n-box-shadow-focus)"})])])]),h("disabled",`
 cursor: not-allowed;
 `,[t("dot",{boxShadow:"var(--n-box-shadow-disabled)",backgroundColor:"var(--n-color-disabled)"},[b("&::before",{backgroundColor:"var(--n-dot-color-disabled)"}),h("checked",`
 opacity: 1;
 `)]),t("label",{color:"var(--n-text-color-disabled)"}),B("radio-input",`
 cursor: not-allowed;
 `)])]),q={name:String,value:{type:[String,Number,Boolean],default:"on"},checked:{type:Boolean,default:void 0},defaultChecked:Boolean,disabled:{type:Boolean,default:void 0},label:String,size:String,onUpdateChecked:[Function,Array],"onUpdate:checked":[Function,Array],checkedValue:{type:Boolean,default:void 0}},X=K("n-radio-group");function ee(o){const e=M(X,null),n=O(o,{mergedSize(a){const{size:r}=o;if(r!==void 0)return r;if(e){const{mergedSizeRef:{value:s}}=e;if(s!==void 0)return s}return a?a.mergedSize.value:"medium"},mergedDisabled(a){return!!(o.disabled||e!=null&&e.disabledRef.value||a!=null&&a.disabled.value)}}),{mergedSizeRef:l,mergedDisabledRef:d}=n,c=p(null),x=p(null),g=p(o.defaultChecked),i=W(o,"checked"),m=H(i,g),u=D(()=>e?e.valueRef.value===o.value:m.value),R=D(()=>{const{name:a}=o;if(a!==void 0)return a;if(e)return e.nameRef.value}),f=p(!1);function k(){if(e){const{doUpdateValue:a}=e,{value:r}=o;$(a,r)}else{const{onUpdateChecked:a,"onUpdate:checked":r}=o,{nTriggerFormInput:s,nTriggerFormChange:y}=n;a&&$(a,!0),r&&$(r,!0),s(),y(),g.value=!0}}function w(){d.value||u.value||k()}function C(){w(),c.value&&(c.value.checked=u.value)}function S(){f.value=!1}function z(){f.value=!0}return{mergedClsPrefix:e?e.mergedClsPrefixRef:_(o).mergedClsPrefixRef,inputRef:c,labelRef:x,mergedName:R,mergedDisabled:d,renderSafeChecked:u,focus:f,mergedSize:l,handleRadioInputChange:C,handleRadioInputBlur:S,handleRadioInputFocus:z}}const oe=Object.assign(Object.assign({},F.props),q),re=G({name:"Radio",props:oe,setup(o){const e=ee(o),n=F("Radio","-radio",Z,Q,o,e.mergedClsPrefix),l=P(()=>{const{mergedSize:{value:m}}=e,{common:{cubicBezierEaseInOut:u},self:{boxShadow:R,boxShadowActive:f,boxShadowDisabled:k,boxShadowFocus:w,boxShadowHover:C,color:S,colorDisabled:z,colorActive:a,textColor:r,textColorDisabled:s,dotColorActive:y,dotColorDisabled:j,labelPadding:A,labelLineHeight:U,labelFontWeight:T,[I("fontSize",m)]:V,[I("radioSize",m)]:E}}=n.value;return{"--n-bezier":u,"--n-label-line-height":U,"--n-label-font-weight":T,"--n-box-shadow":R,"--n-box-shadow-active":f,"--n-box-shadow-disabled":k,"--n-box-shadow-focus":w,"--n-box-shadow-hover":C,"--n-color":S,"--n-color-active":a,"--n-color-disabled":z,"--n-dot-color-active":y,"--n-dot-color-disabled":j,"--n-font-size":V,"--n-radio-size":E,"--n-text-color":r,"--n-text-color-disabled":s,"--n-label-padding":A}}),{inlineThemeDisabled:d,mergedClsPrefixRef:c,mergedRtlRef:x}=_(o),g=Y("Radio",x,c),i=d?J("radio",P(()=>e.mergedSize.value[0]),l,o):void 0;return Object.assign(e,{rtlEnabled:g,cssVars:d?void 0:l,themeClass:i==null?void 0:i.themeClass,onRender:i==null?void 0:i.onRender})},render(){const{$slots:o,mergedClsPrefix:e,onRender:n,label:l}=this;return n==null||n(),v("label",{class:[`${e}-radio`,this.themeClass,this.rtlEnabled&&`${e}-radio--rtl`,this.mergedDisabled&&`${e}-radio--disabled`,this.renderSafeChecked&&`${e}-radio--checked`,this.focus&&`${e}-radio--focus`],style:this.cssVars},v("div",{class:`${e}-radio__dot-wrapper`}," ",v("div",{class:[`${e}-radio__dot`,this.renderSafeChecked&&`${e}-radio__dot--checked`]}),v("input",{ref:"inputRef",type:"radio",class:`${e}-radio-input`,value:this.value,name:this.mergedName,checked:this.renderSafeChecked,disabled:this.mergedDisabled,onChange:this.handleRadioInputChange,onFocus:this.handleRadioInputFocus,onBlur:this.handleRadioInputBlur})),L(o.default,d=>!d&&!l?null:v("div",{ref:"labelRef",class:`${e}-radio__label`},d||l)))}});export{re as N,q as a,X as r,ee as s};
