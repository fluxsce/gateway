const __vite__mapDeps=(i,m=__vite__mapDeps,d=(m.f||(m.f=["assets/main-B268JvO2.js","assets/pinia-BSSStNQg.js","assets/vue-D9N-mpE6.js","assets/niuma-ui-WcquJfTB.js","assets/axios-CzApALvg.js","assets/vue-i18n-u1sm6DuK.js","assets/boot-runtime-BnhxyVHI.js","assets/vue-router-HEnSPLnY.js","assets/niuma-ui-Dm198wG2.css","assets/xterm-6GBZ9nXN.css","assets/main-BLfNa4PJ.css"])))=>i.map(i=>d[i]);
import{_ as o}from"./boot-runtime-BnhxyVHI.js";function n(){var t;(t=document.getElementById("app-boot-splash"))==null||t.remove(),document.documentElement.classList.remove("app-booting")}function r(t){n();const e=document.getElementById("app");e&&(e.innerHTML=`
      <div style="padding: 20px; text-align: center; color: #666;">
        <h2>应用加载失败</h2>
        <p>请刷新页面或联系管理员</p>
        <p style="font-size: 12px; margin-top: 10px;">错误信息: ${t}</p>
      </div>
    `)}function a(){requestAnimationFrame(()=>{requestAnimationFrame(()=>{i()})})}async function i(){try{await(await o(()=>import("./main-B268JvO2.js").then(e=>e.B),__vite__mapDeps([0,1,2,3,4,5,6,7,8,9,10]))).startApp()}catch(t){console.error("应用启动失败:",t);const e=t instanceof Error?t.message:String(t);r(e||"未知错误")}}a();export{n as r};
