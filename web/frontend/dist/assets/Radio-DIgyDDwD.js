import{ak as B,am as h,al as i,ao as T,an as b,ba as M,a1 as O,aZ as G,r as p,aa as H,aV as K,$ as D,au as _,b1 as $,d as L,a2 as v,aq as W,aw as F,bG as q,c as I,az as P,av as Y,aB as Z}from"./index-5VJHBj6L.js";const J=B("radio",`
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
 `),B("radio-input",`
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
 `),T("disabled",`
 cursor: pointer;
 `,[b("&:hover",[i("dot",{boxShadow:"var(--n-box-shadow-hover)"})]),h("focus",[b("&:not(:active)",[i("dot",{boxShadow:"var(--n-box-shadow-focus)"})])])]),h("disabled",`
 cursor: not-allowed;
 `,[i("dot",{boxShadow:"var(--n-box-shadow-disabled)",backgroundColor:"var(--n-color-disabled)"},[b("&::before",{backgroundColor:"var(--n-dot-color-disabled)"}),h("checked",`
 opacity: 1;
 `)]),i("label",{color:"var(--n-text-color-disabled)"}),B("radio-input",`
 cursor: not-allowed;
 `)])]),Q={name:String,value:{type:[String,Number,Boolean],default:"on"},checked:{type:Boolean,default:void 0},defaultChecked:Boolean,disabled:{type:Boolean,default:void 0},label:String,size:String,onUpdateChecked:[Function,Array],"onUpdate:checked":[Function,Array],checkedValue:{type:Boolean,default:void 0}},X=M("n-radio-group");function ee(o){const e=O(X,null),n=G(o,{mergedSize(a){const{size:r}=o;if(r!==void 0)return r;if(e){const{mergedSizeRef:{value:s}}=e;if(s!==void 0)return s}return a?a.mergedSize.value:"medium"},mergedDisabled(a){return!!(o.disabled||e!=null&&e.disabledRef.value||a!=null&&a.disabled.value)}}),{mergedSizeRef:l,mergedDisabledRef:d}=n,c=p(null),x=p(null),g=p(o.defaultChecked),t=H(o,"checked"),m=K(t,g),u=D(()=>e?e.valueRef.value===o.value:m.value),R=D(()=>{const{name:a}=o;if(a!==void 0)return a;if(e)return e.nameRef.value}),f=p(!1);function k(){if(e){const{doUpdateValue:a}=e,{value:r}=o;$(a,r)}else{const{onUpdateChecked:a,"onUpdate:checked":r}=o,{nTriggerFormInput:s,nTriggerFormChange:y}=n;a&&$(a,!0),r&&$(r,!0),s(),y(),g.value=!0}}function w(){d.value||u.value||k()}function C(){w(),c.value&&(c.value.checked=u.value)}function z(){f.value=!1}function S(){f.value=!0}return{mergedClsPrefix:e?e.mergedClsPrefixRef:_(o).mergedClsPrefixRef,inputRef:c,labelRef:x,mergedName:R,mergedDisabled:d,renderSafeChecked:u,focus:f,mergedSize:l,handleRadioInputChange:C,handleRadioInputBlur:z,handleRadioInputFocus:S}}const oe=Object.assign(Object.assign({},F.props),Q),re=L({name:"Radio",props:oe,setup(o){const e=ee(o),n=F("Radio","-radio",J,q,o,e.mergedClsPrefix),l=I(()=>{const{mergedSize:{value:m}}=e,{common:{cubicBezierEaseInOut:u},self:{boxShadow:R,boxShadowActive:f,boxShadowDisabled:k,boxShadowFocus:w,boxShadowHover:C,color:z,colorDisabled:S,colorActive:a,textColor:r,textColorDisabled:s,dotColorActive:y,dotColorDisabled:U,labelPadding:j,labelLineHeight:A,labelFontWeight:V,[P("fontSize",m)]:E,[P("radioSize",m)]:N}}=n.value;return{"--n-bezier":u,"--n-label-line-height":A,"--n-label-font-weight":V,"--n-box-shadow":R,"--n-box-shadow-active":f,"--n-box-shadow-disabled":k,"--n-box-shadow-focus":w,"--n-box-shadow-hover":C,"--n-color":z,"--n-color-active":a,"--n-color-disabled":S,"--n-dot-color-active":y,"--n-dot-color-disabled":U,"--n-font-size":E,"--n-radio-size":N,"--n-text-color":r,"--n-text-color-disabled":s,"--n-label-padding":j}}),{inlineThemeDisabled:d,mergedClsPrefixRef:c,mergedRtlRef:x}=_(o),g=Y("Radio",x,c),t=d?Z("radio",I(()=>e.mergedSize.value[0]),l,o):void 0;return Object.assign(e,{rtlEnabled:g,cssVars:d?void 0:l,themeClass:t==null?void 0:t.themeClass,onRender:t==null?void 0:t.onRender})},render(){const{$slots:o,mergedClsPrefix:e,onRender:n,label:l}=this;return n==null||n(),v("label",{class:[`${e}-radio`,this.themeClass,this.rtlEnabled&&`${e}-radio--rtl`,this.mergedDisabled&&`${e}-radio--disabled`,this.renderSafeChecked&&`${e}-radio--checked`,this.focus&&`${e}-radio--focus`],style:this.cssVars},v("input",{ref:"inputRef",type:"radio",class:`${e}-radio-input`,value:this.value,name:this.mergedName,checked:this.renderSafeChecked,disabled:this.mergedDisabled,onChange:this.handleRadioInputChange,onFocus:this.handleRadioInputFocus,onBlur:this.handleRadioInputBlur}),v("div",{class:`${e}-radio__dot-wrapper`}," ",v("div",{class:[`${e}-radio__dot`,this.renderSafeChecked&&`${e}-radio__dot--checked`]})),W(o.default,d=>!d&&!l?null:v("div",{ref:"labelRef",class:`${e}-radio__label`},d||l)))}});export{re as N,Q as a,X as r,ee as s};
