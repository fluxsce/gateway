import{P as $,Q as h,R as i,aA as N,S as b,bf as M,ak as O,b3 as H,r as p,X as K,Z as W,ai as D,V as _,a2 as B,d as G,T as v,aC as L,W as F,bJ as J,c as I,aI as P,aF as Q,$ as X}from"./index-DjxxKbEp.js";const Y=$("radio",`
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
`,[h("checked",[i("dot",`
 background-color: var(--n-color-active);
 `)]),i("dot-wrapper",`
 position: relative;
 flex-shrink: 0;
 flex-grow: 0;
 width: var(--n-radio-size);
 `),$("radio-input",`
 position: absolute;
 border: 0;
 border-radius: inherit;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 opacity: 0;
 z-index: 1;
 cursor: pointer;
 `),i("dot",`
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
 `)])]),i("label",`
 color: var(--n-text-color);
 padding: var(--n-label-padding);
 font-weight: var(--n-label-font-weight);
 display: inline-block;
 transition: color .3s var(--n-bezier);
 `),N("disabled",`
 cursor: pointer;
 `,[b("&:hover",[i("dot",{boxShadow:"var(--n-box-shadow-hover)"})]),h("focus",[b("&:not(:active)",[i("dot",{boxShadow:"var(--n-box-shadow-focus)"})])])]),h("disabled",`
 cursor: not-allowed;
 `,[i("dot",{boxShadow:"var(--n-box-shadow-disabled)",backgroundColor:"var(--n-color-disabled)"},[b("&::before",{backgroundColor:"var(--n-dot-color-disabled)"}),h("checked",`
 opacity: 1;
 `)]),i("label",{color:"var(--n-text-color-disabled)"}),$("radio-input",`
 cursor: not-allowed;
 `)])]),Z={name:String,value:{type:[String,Number,Boolean],default:"on"},checked:{type:Boolean,default:void 0},defaultChecked:Boolean,disabled:{type:Boolean,default:void 0},label:String,size:String,onUpdateChecked:[Function,Array],"onUpdate:checked":[Function,Array],checkedValue:{type:Boolean,default:void 0}},q=M("n-radio-group");function ee(o){const e=O(q,null),n=H(o,{mergedSize(a){const{size:r}=o;if(r!==void 0)return r;if(e){const{mergedSizeRef:{value:s}}=e;if(s!==void 0)return s}return a?a.mergedSize.value:"medium"},mergedDisabled(a){return!!(o.disabled||e!=null&&e.disabledRef.value||a!=null&&a.disabled.value)}}),{mergedSizeRef:l,mergedDisabledRef:d}=n,c=p(null),x=p(null),g=p(o.defaultChecked),t=K(o,"checked"),m=W(t,g),u=D(()=>e?e.valueRef.value===o.value:m.value),R=D(()=>{const{name:a}=o;if(a!==void 0)return a;if(e)return e.nameRef.value}),f=p(!1);function k(){if(e){const{doUpdateValue:a}=e,{value:r}=o;B(a,r)}else{const{onUpdateChecked:a,"onUpdate:checked":r}=o,{nTriggerFormInput:s,nTriggerFormChange:y}=n;a&&B(a,!0),r&&B(r,!0),s(),y(),g.value=!0}}function C(){d.value||u.value||k()}function w(){C(),c.value&&(c.value.checked=u.value)}function S(){f.value=!1}function z(){f.value=!0}return{mergedClsPrefix:e?e.mergedClsPrefixRef:_(o).mergedClsPrefixRef,inputRef:c,labelRef:x,mergedName:R,mergedDisabled:d,renderSafeChecked:u,focus:f,mergedSize:l,handleRadioInputChange:w,handleRadioInputBlur:S,handleRadioInputFocus:z}}const oe=Object.assign(Object.assign({},F.props),Z),re=G({name:"Radio",props:oe,setup(o){const e=ee(o),n=F("Radio","-radio",Y,J,o,e.mergedClsPrefix),l=I(()=>{const{mergedSize:{value:m}}=e,{common:{cubicBezierEaseInOut:u},self:{boxShadow:R,boxShadowActive:f,boxShadowDisabled:k,boxShadowFocus:C,boxShadowHover:w,color:S,colorDisabled:z,colorActive:a,textColor:r,textColorDisabled:s,dotColorActive:y,dotColorDisabled:A,labelPadding:U,labelLineHeight:j,labelFontWeight:T,[P("fontSize",m)]:V,[P("radioSize",m)]:E}}=n.value;return{"--n-bezier":u,"--n-label-line-height":j,"--n-label-font-weight":T,"--n-box-shadow":R,"--n-box-shadow-active":f,"--n-box-shadow-disabled":k,"--n-box-shadow-focus":C,"--n-box-shadow-hover":w,"--n-color":S,"--n-color-active":a,"--n-color-disabled":z,"--n-dot-color-active":y,"--n-dot-color-disabled":A,"--n-font-size":V,"--n-radio-size":E,"--n-text-color":r,"--n-text-color-disabled":s,"--n-label-padding":U}}),{inlineThemeDisabled:d,mergedClsPrefixRef:c,mergedRtlRef:x}=_(o),g=Q("Radio",x,c),t=d?X("radio",I(()=>e.mergedSize.value[0]),l,o):void 0;return Object.assign(e,{rtlEnabled:g,cssVars:d?void 0:l,themeClass:t==null?void 0:t.themeClass,onRender:t==null?void 0:t.onRender})},render(){const{$slots:o,mergedClsPrefix:e,onRender:n,label:l}=this;return n==null||n(),v("label",{class:[`${e}-radio`,this.themeClass,this.rtlEnabled&&`${e}-radio--rtl`,this.mergedDisabled&&`${e}-radio--disabled`,this.renderSafeChecked&&`${e}-radio--checked`,this.focus&&`${e}-radio--focus`],style:this.cssVars},v("input",{ref:"inputRef",type:"radio",class:`${e}-radio-input`,value:this.value,name:this.mergedName,checked:this.renderSafeChecked,disabled:this.mergedDisabled,onChange:this.handleRadioInputChange,onFocus:this.handleRadioInputFocus,onBlur:this.handleRadioInputBlur}),v("div",{class:`${e}-radio__dot-wrapper`}," ",v("div",{class:[`${e}-radio__dot`,this.renderSafeChecked&&`${e}-radio__dot--checked`]})),L(o.default,d=>!d&&!l?null:v("div",{ref:"labelRef",class:`${e}-radio__label`},d||l)))}});export{re as N,Z as a,q as r,ee as s};
