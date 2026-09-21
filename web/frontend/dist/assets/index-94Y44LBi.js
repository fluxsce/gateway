const __vite__mapDeps=(i,m=__vite__mapDeps,d=(m.f||(m.f=["assets/main-Cmcd73zE.js","assets/pinia-DSvvG8AT.js","assets/vue-YjpWV8up.js","assets/niuma-ui-C48iqLFa.js","assets/axios-CzApALvg.js","assets/vue-i18n-D1Q_ct_M.js","assets/boot-runtime-BnhxyVHI.js","assets/vue-router-BaudgDyd.js","assets/niuma-ui-ByVExN4i.css","assets/monaco-CP7YX7lX.css","assets/xterm-6GBZ9nXN.css","assets/main-DzxPXDYD.css"])))=>i.map(i=>d[i]);
import{_ as o}from"./boot-runtime-BnhxyVHI.js";function n(){var t;(t=document.getElementById("app-boot-splash"))==null||t.remove(),document.documentElement.classList.remove("app-booting")}function r(t){n();const e=document.getElementById("app");e&&(e.innerHTML=`
      <div style="padding: 20px; text-align: center; color: #666;">
        <h2>应用加载失败</h2>
        <p>请刷新页面或联系管理员</p>
        <p style="font-size: 12px; margin-top: 10px;">错误信息: ${t}</p>
      </div>
    `)}function a(){requestAnimationFrame(()=>{requestAnimationFrame(()=>{i()})})}async function i(){try{await(await o(()=>import("./main-Cmcd73zE.js").then(e=>e.B),__vite__mapDeps([0,1,2,3,4,5,6,7,8,9,10,11]))).startApp()}catch(t){console.error("应用启动失败:",t);const e=t instanceof Error?t.message:String(t);r(e||"未知错误")}}a();export{n as r};
