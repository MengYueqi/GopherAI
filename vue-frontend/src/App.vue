<template>
  <div id="app">
    <router-view v-slot="{ Component }">
      <transition name="page" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>
  </div>
</template>

<script>
export default {
  name: 'App'
}
</script>

<style>
:root {
  --bg-color: #eef1ff;
  --bg-gradient: radial-gradient(circle at 20% 20%, rgba(104, 117, 245, 0.18), transparent 45%),
    radial-gradient(circle at 80% 0%, rgba(236, 72, 153, 0.14), transparent 40%),
    radial-gradient(circle at 0% 80%, rgba(16, 185, 129, 0.13), transparent 35%);
  --card-bg: rgba(255, 255, 255, 0.88);
  --card-border: rgba(99, 102, 241, 0.14);
  --text-color: #0f172a;
  --text-muted: #0f172a;
  --primary: #7064ff;
  --primary-dark: #594bff;
  --secondary: #2dd4bf;
  --accent: #ff8fb3;
  --radius-lg: 28px;
  --radius-md: 18px;
  --radius-sm: 12px;
  --shadow-lg: 0 35px 80px rgba(15, 23, 42, 0.15);
  --shadow-md: 0 20px 45px rgba(15, 23, 42, 0.12);
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html,
body {
  height: 100%;
  font-family: 'Google Sans', 'Roboto', 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', '微软雅黑', Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: var(--text-color);
  background: var(--bg-color);
}

body {
  position: relative;
}

body::before {
  content: '';
  position: fixed;
  inset: 0;
  background: var(--bg-gradient);
  z-index: -2;
}

body::after {
  content: '';
  position: fixed;
  inset: 0;
  background-image: linear-gradient(rgba(255, 255, 255, 0.08) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.08) 1px, transparent 1px);
  background-size: 120px 120px;
  opacity: 0.3;
  z-index: -1;
  pointer-events: none;
}

#app {
  min-height: 100%;
}

::selection {
  background: rgba(112, 100, 255, 0.2);
}

/* 页面切换动画 */
.page-enter-active,
.page-leave-active {
  transition: all 0.45s cubic-bezier(0.6, 0, 0.18, 1);
}

.page-enter-from {
  opacity: 0;
  transform: translateY(16px) scale(0.98);
}

.page-leave-to {
  opacity: 0;
  transform: translateY(-12px) scale(0.98);
}

/* 全局滚动条样式 */
::-webkit-scrollbar {
  width: 9px;
  height: 9px;
}

::-webkit-scrollbar-track {
  background: rgba(15, 23, 42, 0.08);
  border-radius: 100px;
}

::-webkit-scrollbar-thumb {
  background: linear-gradient(135deg, rgba(112, 100, 255, 0.85), rgba(45, 212, 191, 0.85));
  border-radius: 100px;
  border: 2px solid transparent;
}

::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(135deg, rgba(89, 75, 255, 0.95), rgba(16, 185, 129, 0.95));
}

/* 基础排版 */
p {
  color: var(--text-muted);
  line-height: 1.7;
}

a {
  color: var(--primary);
  text-decoration: none;
}

a:hover {
  color: var(--primary-dark);
}

.glass-panel {
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  backdrop-filter: blur(32px);
}

.pill {
  display: inline-flex;
  align-items: center;
  padding: 6px 12px;
  border-radius: 999px;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  background: rgba(112, 100, 255, 0.12);
  color: var(--primary-dark);
  font-weight: 600;
}

/* Element Plus 组件样式覆盖 */
.el-button {
  font-weight: 600;
  border-radius: var(--radius-sm);
  transition: all 0.2s ease;
  border: none;
  box-shadow: 0 8px 20px rgba(112, 100, 255, 0.2);
}

.el-button + .el-button {
  margin-left: 8px;
}

.el-button--primary {
  background-image: linear-gradient(120deg, var(--primary), var(--secondary));
  color: #fff;
}

.el-button--primary:focus,
.el-button--primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 16px 30px rgba(112, 100, 255, 0.3);
}

.el-button--text {
  color: var(--primary);
  text-align: center;
}

.el-card {
  border-radius: var(--radius-lg);
  border: 1px solid var(--card-border);
  box-shadow: var(--shadow-md);
  background: var(--card-bg);
  backdrop-filter: blur(24px);
}

.el-input {
  border-radius: var(--radius-sm);
}

.el-message {
  border-radius: var(--radius-sm);
  background: var(--card-bg);
  border-color: transparent;
  box-shadow: 0 14px 30px rgba(15, 23, 42, 0.15);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .page-enter-from,
  .page-leave-to {
    transform: translateY(0) scale(1);
    opacity: 0;
  }

  .page-enter-active,
  .page-leave-active {
    transition: opacity 0.3s ease;
  }
}
</style>
