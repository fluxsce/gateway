const __vite__mapDeps=(i,m=__vite__mapDeps,d=(m.f||(m.f=["./main-DosCWgSE.js","./pinia-B2l23fkc.js","./vue-WBel5DBb.js","./ensis-ui-674SLu0K.js","./axios-CzApALvg.js","./vue-i18n-Bti0-2zo.js","./boot-runtime-CovLlAO2.js","./vue-router-NtmOMmJy.js","./ensis-ui-CJ3vafqS.css","./xterm-6GBZ9nXN.css","./main-BLfNa4PJ.css"])))=>i.map(i=>d[i]);
import{_ as o}from"./boot-runtime-CovLlAO2.js";function n(){var t;(t=document.getElementById("app-boot-splash"))==null||t.remove(),document.documentElement.classList.remove("app-booting")}function r(t){n();const e=document.getElementById("app");e&&(e.innerHTML=`
      <div style="padding: 20px; text-align: center; color: #666;">
        <h2>应用加载失败</h2>
        <p>请刷新页面或联系管理员</p>
        <p style="font-size: 12px; margin-top: 10px;">错误信息: ${t}</p>
      </div>
    `)}function a(){requestAnimationFrame(()=>{requestAnimationFrame(()=>{i()})})}async function i(){try{await(await o(()=>import("./main-DosCWgSE.js").then(e=>e.B),__vite__mapDeps([0,1,2,3,4,5,6,7,8,9,10]),import.meta.url)).startApp()}catch(t){console.error("应用启动失败:",t);const e=t instanceof Error?t.message:String(t);r(e||"未知错误")}}a();export{n as r};
