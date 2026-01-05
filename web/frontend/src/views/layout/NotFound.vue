<template>
    <div class="not-found">
        <div class="container">
            <div class="error-code">404</div>
            <h1>页面未找到</h1>
            <p>抱歉，您访问的页面不存在或已被移除</p>
            <div class="actions">
                <router-link to="/" class="btn-primary">返回首页</router-link>
                <button @click="goBack" class="btn-secondary">返回上一页</button>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

const goBack = () => {
    router.back()
}
</script>

<style scoped>
.not-found {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    padding: 20px;
    text-align: center;
    background-color: #f8f9fa;
}

.container {
    max-width: 600px;
}

.error-code {
    font-size: 120px;
    font-weight: 700;
    color: #1976d2;
    line-height: 1;
    margin-bottom: 20px;
}

h1 {
    font-size: 32px;
    margin-bottom: 16px;
    color: #333;
}

p {
    font-size: 18px;
    color: #666;
    margin-bottom: 30px;
}

.actions {
    display: flex;
    justify-content: center;
    gap: 16px;
}

.btn-primary,
.btn-secondary {
    padding: 12px 24px;
    border-radius: 4px;
    font-size: 16px;
    cursor: pointer;
    transition: all 0.2s;
    text-decoration: none;
}

.btn-primary {
    background-color: #1976d2;
    color: white;
    border: none;
}

.btn-primary:hover {
    background-color: #1565c0;
}

.btn-secondary {
    background-color: transparent;
    border: 1px solid #1976d2;
    color: #1976d2;
}

.btn-secondary:hover {
    background-color: rgba(25, 118, 210, 0.08);
}
</style>