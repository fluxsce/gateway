import{d as E,a8 as n,aq as f,as as x,at as z,ar as s,au as U,c2 as O,aA as _,r as V,g as N,a$ as W,aC as D,d9 as q,aB as T,aH as K,b8 as X,a6 as Y,b7 as S,c3 as Z,aX as G,da as J,ag as B,aY as Q,bk as k,db as ee,bv as ae,a5 as re,bl as te,a7 as le,aI as A,ao as oe,bI as se,c as ne,o as ie,a as de}from"./index-DKB7UIvv.js";const ce=E({name:"ChevronLeft",render(){return n("svg",{viewBox:"0 0 16 16",fill:"none",xmlns:"http://www.w3.org/2000/svg"},n("path",{d:"M10.3536 3.14645C10.5488 3.34171 10.5488 3.65829 10.3536 3.85355L6.20711 8L10.3536 12.1464C10.5488 12.3417 10.5488 12.6583 10.3536 12.8536C10.1583 13.0488 9.84171 13.0488 9.64645 12.8536L5.14645 8.35355C4.95118 8.15829 4.95118 7.84171 5.14645 7.64645L9.64645 3.14645C9.84171 2.95118 10.1583 2.95118 10.3536 3.14645Z",fill:"currentColor"}))}}),pe=f("collapse","width: 100%;",[f("collapse-item",`
 font-size: var(--n-font-size);
 color: var(--n-text-color);
 transition:
 color .3s var(--n-bezier),
 border-color .3s var(--n-bezier);
 margin: var(--n-item-margin);
 `,[x("disabled",[s("header","cursor: not-allowed;",[s("header-main",`
 color: var(--n-title-text-color-disabled);
 `),f("collapse-item-arrow",`
 color: var(--n-arrow-color-disabled);
 `)])]),f("collapse-item","margin-left: 32px;"),z("&:first-child","margin-top: 0;"),z("&:first-child >",[s("header","padding-top: 0;")]),x("left-arrow-placement",[s("header",[f("collapse-item-arrow","margin-right: 4px;")])]),x("right-arrow-placement",[s("header",[f("collapse-item-arrow","margin-left: 4px;")])]),s("content-wrapper",[s("content-inner","padding-top: 16px;"),O({duration:"0.15s"})]),x("active",[s("header",[x("active",[f("collapse-item-arrow","transform: rotate(90deg);")])])]),z("&:not(:first-child)","border-top: 1px solid var(--n-divider-color);"),U("disabled",[x("trigger-area-main",[s("header",[s("header-main","cursor: pointer;"),f("collapse-item-arrow","cursor: default;")])]),x("trigger-area-arrow",[s("header",[f("collapse-item-arrow","cursor: pointer;")])]),x("trigger-area-extra",[s("header",[s("header-extra","cursor: pointer;")])])]),s("header",`
 font-size: var(--n-title-font-size);
 display: flex;
 flex-wrap: nowrap;
 align-items: center;
 transition: color .3s var(--n-bezier);
 position: relative;
 padding: var(--n-title-padding);
 color: var(--n-title-text-color);
 `,[s("header-main",`
 display: flex;
 flex-wrap: nowrap;
 align-items: center;
 font-weight: var(--n-title-font-weight);
 transition: color .3s var(--n-bezier);
 flex: 1;
 color: var(--n-title-text-color);
 `),s("header-extra",`
 display: flex;
 align-items: center;
 transition: color .3s var(--n-bezier);
 color: var(--n-text-color);
 `),f("collapse-item-arrow",`
 display: flex;
 transition:
 transform .15s var(--n-bezier),
 color .3s var(--n-bezier);
 font-size: 18px;
 color: var(--n-arrow-color);
 `)])])]),me=Object.assign(Object.assign({},D.props),{defaultExpandedNames:{type:[Array,String],default:null},expandedNames:[Array,String],arrowPlacement:{type:String,default:"left"},accordion:{type:Boolean,default:!1},displayDirective:{type:String,default:"if"},triggerAreas:{type:Array,default:()=>["main","extra","arrow"]},onItemHeaderClick:[Function,Array],"onUpdate:expandedNames":[Function,Array],onUpdateExpandedNames:[Function,Array],onExpandedNamesChange:{type:[Function,Array],validator:()=>!0,default:void 0}}),H=X("n-collapse"),xe=E({name:"Collapse",props:me,slots:Object,setup(e,{slots:i}){const{mergedClsPrefixRef:l,inlineThemeDisabled:o,mergedRtlRef:d}=_(e),r=V(e.defaultExpandedNames),h=N(()=>e.expandedNames),v=W(h,r),w=D("Collapse","-collapse",pe,q,e,l);function c(p){const{"onUpdate:expandedNames":t,onUpdateExpandedNames:m,onExpandedNamesChange:y}=e;m&&S(m,p),t&&S(t,p),y&&S(y,p),r.value=p}function g(p){const{onItemHeaderClick:t}=e;t&&S(t,p)}function a(p,t,m){const{accordion:y}=e,{value:I}=v;if(y)p?(c([t]),g({name:t,expanded:!0,event:m})):(c([]),g({name:t,expanded:!1,event:m}));else if(!Array.isArray(I))c([t]),g({name:t,expanded:!0,event:m});else{const C=I.slice(),P=C.findIndex($=>t===$);~P?(C.splice(P,1),c(C),g({name:t,expanded:!1,event:m})):(C.push(t),c(C),g({name:t,expanded:!0,event:m}))}}Y(H,{props:e,mergedClsPrefixRef:l,expandedNamesRef:v,slots:i,toggleItem:a});const u=T("Collapse",d,l),R=N(()=>{const{common:{cubicBezierEaseInOut:p},self:{titleFontWeight:t,dividerColor:m,titlePadding:y,titleTextColor:I,titleTextColorDisabled:C,textColor:P,arrowColor:$,fontSize:F,titleFontSize:j,arrowColorDisabled:L,itemMargin:M}}=w.value;return{"--n-font-size":F,"--n-bezier":p,"--n-text-color":P,"--n-divider-color":m,"--n-title-padding":y,"--n-title-font-size":j,"--n-title-text-color":I,"--n-title-text-color-disabled":C,"--n-title-font-weight":t,"--n-arrow-color":$,"--n-arrow-color-disabled":L,"--n-item-margin":M}}),b=o?K("collapse",void 0,R,e):void 0;return{rtlEnabled:u,mergedTheme:w,mergedClsPrefix:l,cssVars:o?void 0:R,themeClass:b==null?void 0:b.themeClass,onRender:b==null?void 0:b.onRender}},render(){var e;return(e=this.onRender)===null||e===void 0||e.call(this),n("div",{class:[`${this.mergedClsPrefix}-collapse`,this.rtlEnabled&&`${this.mergedClsPrefix}-collapse--rtl`,this.themeClass],style:this.cssVars},this.$slots)}}),fe=E({name:"CollapseItemContent",props:{displayDirective:{type:String,required:!0},show:Boolean,clsPrefix:{type:String,required:!0}},setup(e){return{onceTrue:J(B(e,"show"))}},render(){return n(Z,null,{default:()=>{const{show:e,displayDirective:i,onceTrue:l,clsPrefix:o}=this,d=i==="show"&&l,r=n("div",{class:`${o}-collapse-item__content-wrapper`},n("div",{class:`${o}-collapse-item__content-inner`},this.$slots));return d?G(r,[[Q,e]]):e?r:null}})}}),ue={title:String,name:[String,Number],disabled:Boolean,displayDirective:String},ve=E({name:"CollapseItem",props:ue,setup(e){const{mergedRtlRef:i}=_(e),l=ae(),o=re(()=>{var a;return(a=e.name)!==null&&a!==void 0?a:l}),d=le(H);d||te("collapse-item","`n-collapse-item` must be placed inside `n-collapse`.");const{expandedNamesRef:r,props:h,mergedClsPrefixRef:v,slots:w}=d,c=N(()=>{const{value:a}=r;if(Array.isArray(a)){const{value:u}=o;return!~a.findIndex(R=>R===u)}else if(a){const{value:u}=o;return u!==a}return!0});return{rtlEnabled:T("Collapse",i,v),collapseSlots:w,randomName:l,mergedClsPrefix:v,collapsed:c,triggerAreas:B(h,"triggerAreas"),mergedDisplayDirective:N(()=>{const{displayDirective:a}=e;return a||h.displayDirective}),arrowPlacement:N(()=>h.arrowPlacement),handleClick(a){let u="main";A(a,"arrow")&&(u="arrow"),A(a,"extra")&&(u="extra"),h.triggerAreas.includes(u)&&d&&!e.disabled&&d.toggleItem(c.value,o.value,a)}}},render(){const{collapseSlots:e,$slots:i,arrowPlacement:l,collapsed:o,mergedDisplayDirective:d,mergedClsPrefix:r,disabled:h,triggerAreas:v}=this,w=k(i.header,{collapsed:o},()=>[this.title]),c=i["header-extra"]||e["header-extra"],g=i.arrow||e.arrow;return n("div",{class:[`${r}-collapse-item`,`${r}-collapse-item--${l}-arrow-placement`,h&&`${r}-collapse-item--disabled`,!o&&`${r}-collapse-item--active`,v.map(a=>`${r}-collapse-item--trigger-area-${a}`)]},n("div",{class:[`${r}-collapse-item__header`,!o&&`${r}-collapse-item__header--active`]},n("div",{class:`${r}-collapse-item__header-main`,onClick:this.handleClick},l==="right"&&w,n("div",{class:`${r}-collapse-item-arrow`,key:this.rtlEnabled?0:1,"data-arrow":!0},k(g,{collapsed:o},()=>[n(oe,{clsPrefix:r},{default:()=>this.rtlEnabled?n(ce,null):n(se,null)})])),l==="left"&&w),ee(c,{collapsed:o},a=>n("div",{class:`${r}-collapse-item__header-extra`,onClick:this.handleClick,"data-extra":!0},a))),n(fe,{clsPrefix:r,displayDirective:d,show:!o},i))}}),he={xmlns:"http://www.w3.org/2000/svg","xmlns:xlink":"http://www.w3.org/1999/xlink",viewBox:"0 0 512 512"},we=E({name:"HeartOutline",render:function(i,l){return ie(),ne("svg",he,l[0]||(l[0]=[de("path",{d:"M352.92 80C288 80 256 144 256 144s-32-64-96.92-64c-52.76 0-94.54 44.14-95.08 96.81c-1.1 109.33 86.73 187.08 183 252.42a16 16 0 0 0 18 0c96.26-65.34 184.09-143.09 183-252.42c-.54-52.67-42.32-96.81-95.08-96.81z",fill:"none",stroke:"currentColor","stroke-linecap":"round","stroke-linejoin":"round","stroke-width":"32"},null,-1)]))}});export{we as H,xe as N,ve as a};
