import{ak as w,al as u,am as h,an as y,ao as F,d as J,bH as Q,bI as W,a2 as $,r as I,aZ as X,au as Y,aw as G,bG as oo,aa as V,aV as to,a0 as eo,av as ro,c as S,az as T,aB as no,b1 as E}from"./index-MyDFJZSo.js";import{r as ao}from"./Radio-BDKhEig0.js";const io=w("radio-group",`
 display: inline-block;
 font-size: var(--n-font-size);
`,[u("splitor",`
 display: inline-block;
 vertical-align: bottom;
 width: 1px;
 transition:
 background-color .3s var(--n-bezier),
 opacity .3s var(--n-bezier);
 background: var(--n-button-border-color);
 `,[h("checked",{backgroundColor:"var(--n-button-border-color-active)"}),h("disabled",{opacity:"var(--n-opacity-disabled)"})]),h("button-group",`
 white-space: nowrap;
 height: var(--n-height);
 line-height: var(--n-height);
 `,[w("radio-button",{height:"var(--n-height)",lineHeight:"var(--n-height)"}),u("splitor",{height:"var(--n-height)"})]),w("radio-button",`
 vertical-align: bottom;
 outline: none;
 position: relative;
 user-select: none;
 -webkit-user-select: none;
 display: inline-block;
 box-sizing: border-box;
 padding-left: 14px;
 padding-right: 14px;
 white-space: nowrap;
 transition:
 background-color .3s var(--n-bezier),
 opacity .3s var(--n-bezier),
 border-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 background: var(--n-button-color);
 color: var(--n-button-text-color);
 border-top: 1px solid var(--n-button-border-color);
 border-bottom: 1px solid var(--n-button-border-color);
 `,[w("radio-input",`
 pointer-events: none;
 position: absolute;
 border: 0;
 border-radius: inherit;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 opacity: 0;
 z-index: 1;
 `),u("state-border",`
 z-index: 1;
 pointer-events: none;
 position: absolute;
 box-shadow: var(--n-button-box-shadow);
 transition: box-shadow .3s var(--n-bezier);
 left: -1px;
 bottom: -1px;
 right: -1px;
 top: -1px;
 `),y("&:first-child",`
 border-top-left-radius: var(--n-button-border-radius);
 border-bottom-left-radius: var(--n-button-border-radius);
 border-left: 1px solid var(--n-button-border-color);
 `,[u("state-border",`
 border-top-left-radius: var(--n-button-border-radius);
 border-bottom-left-radius: var(--n-button-border-radius);
 `)]),y("&:last-child",`
 border-top-right-radius: var(--n-button-border-radius);
 border-bottom-right-radius: var(--n-button-border-radius);
 border-right: 1px solid var(--n-button-border-color);
 `,[u("state-border",`
 border-top-right-radius: var(--n-button-border-radius);
 border-bottom-right-radius: var(--n-button-border-radius);
 `)]),F("disabled",`
 cursor: pointer;
 `,[y("&:hover",[u("state-border",`
 transition: box-shadow .3s var(--n-bezier);
 box-shadow: var(--n-button-box-shadow-hover);
 `),F("checked",{color:"var(--n-button-text-color-hover)"})]),h("focus",[y("&:not(:active)",[u("state-border",{boxShadow:"var(--n-button-box-shadow-focus)"})])])]),h("checked",`
 background: var(--n-button-color-active);
 color: var(--n-button-text-color-active);
 border-color: var(--n-button-border-color-active);
 `),h("disabled",`
 cursor: not-allowed;
 opacity: var(--n-opacity-disabled);
 `)])]);function so(o,a,t){var s;const e=[];let c=!1;for(let i=0;i<o.length;++i){const d=o[i],l=(s=d.type)===null||s===void 0?void 0:s.name;l==="RadioButton"&&(c=!0);const p=d.props;if(l!=="RadioButton"){e.push(d);continue}if(i===0)e.push(d);else{const f=e[e.length-1].props,m=a===f.value,v=f.disabled,x=a===p.value,g=p.disabled,z=(m?2:0)+(v?0:1),k=(x?2:0)+(g?0:1),_={[`${t}-radio-group__splitor--disabled`]:v,[`${t}-radio-group__splitor--checked`]:m},B={[`${t}-radio-group__splitor--disabled`]:g,[`${t}-radio-group__splitor--checked`]:x},R=z<k?B:_;e.push($("div",{class:[`${t}-radio-group__splitor`,R]}),d)}}return{children:e,isButtonGroup:c}}const lo=Object.assign(Object.assign({},G.props),{name:String,value:[String,Number,Boolean],defaultValue:{type:[String,Number,Boolean],default:null},size:String,disabled:{type:Boolean,default:void 0},"onUpdate:value":[Function,Array],onUpdateValue:[Function,Array]}),bo=J({name:"RadioGroup",props:lo,setup(o){const a=I(null),{mergedSizeRef:t,mergedDisabledRef:s,nTriggerFormChange:e,nTriggerFormInput:c,nTriggerFormBlur:i,nTriggerFormFocus:d}=X(o),{mergedClsPrefixRef:l,inlineThemeDisabled:p,mergedRtlRef:f}=Y(o),m=G("Radio","-radio-group",io,oo,o,l),v=I(o.defaultValue),x=V(o,"value"),g=to(x,v);function z(r){const{onUpdateValue:n,"onUpdate:value":C}=o;n&&E(n,r),C&&E(C,r),v.value=r,e(),c()}function k(r){const{value:n}=a;n&&(n.contains(r.relatedTarget)||d())}function _(r){const{value:n}=a;n&&(n.contains(r.relatedTarget)||i())}eo(ao,{mergedClsPrefixRef:l,nameRef:V(o,"name"),valueRef:g,disabledRef:s,mergedSizeRef:t,doUpdateValue:z});const B=ro("Radio",f,l),R=S(()=>{const{value:r}=t,{common:{cubicBezierEaseInOut:n},self:{buttonBorderColor:C,buttonBorderColorActive:H,buttonBorderRadius:U,buttonBoxShadow:A,buttonBoxShadowFocus:D,buttonBoxShadowHover:P,buttonColor:j,buttonColorActive:M,buttonTextColor:N,buttonTextColorActive:O,buttonTextColorHover:K,opacityDisabled:L,[T("buttonHeight",r)]:Z,[T("fontSize",r)]:q}}=m.value;return{"--n-font-size":q,"--n-bezier":n,"--n-button-border-color":C,"--n-button-border-color-active":H,"--n-button-border-radius":U,"--n-button-box-shadow":A,"--n-button-box-shadow-focus":D,"--n-button-box-shadow-hover":P,"--n-button-color":j,"--n-button-color-active":M,"--n-button-text-color":N,"--n-button-text-color-hover":K,"--n-button-text-color-active":O,"--n-height":Z,"--n-opacity-disabled":L}}),b=p?no("radio-group",S(()=>t.value[0]),R,o):void 0;return{selfElRef:a,rtlEnabled:B,mergedClsPrefix:l,mergedValue:g,handleFocusout:_,handleFocusin:k,cssVars:p?void 0:R,themeClass:b==null?void 0:b.themeClass,onRender:b==null?void 0:b.onRender}},render(){var o;const{mergedValue:a,mergedClsPrefix:t,handleFocusin:s,handleFocusout:e}=this,{children:c,isButtonGroup:i}=so(Q(W(this)),a,t);return(o=this.onRender)===null||o===void 0||o.call(this),$("div",{onFocusin:s,onFocusout:e,ref:"selfElRef",class:[`${t}-radio-group`,this.rtlEnabled&&`${t}-radio-group--rtl`,this.themeClass,i&&`${t}-radio-group--button-group`],style:this.cssVars},c)}});export{bo as _};
