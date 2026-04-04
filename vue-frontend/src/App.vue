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
  /* 中国传统色彩体系 */
  --cinnabar: #9D2933; /* 朱砂红 */
  --indigo: #1A365D; /* 石青 */
  --ochre: #8C4320; /* 赭石 */
  --ivory: #F5EFE2; /* 牙白 */
  --ink: #1C1C1C; /* 墨黑 */
  --jade: #4A7059; /* 玉色 */
  --gold: #D4AF37; /* 金色 */
  --rice-paper: #FAF6ED; /* 宣纸色 */
  --pine: #2C3E2E; /* 松色 */

  /* 功能色彩 */
  --bg-color: var(--rice-paper);
  --card-bg: rgba(255, 253, 245, 0.95);
  --card-border: rgba(157, 41, 51, 0.15);
  --text-color: var(--ink);
  --text-muted: #5C5C5C;
  --primary: var(--cinnabar);
  --primary-dark: #822029;
  --secondary: var(--indigo);
  --accent: var(--gold);
  --success: var(--jade);

  /* 圆角 - 中式圆角偏方圆 */
  --radius-lg: 12px;
  --radius-md: 8px;
  --radius-sm: 4px;

  /* 阴影 - 水墨阴影效果 */
  --shadow-lg: 0 20px 60px rgba(28, 28, 28, 0.12);
  --shadow-md: 0 12px 30px rgba(28, 28, 28, 0.08);
  --shadow-sm: 0 4px 12px rgba(28, 28, 28, 0.06);
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html,
body {
  height: 100%;
  font-family: 'Noto Serif SC', 'Source Han Serif SC', 'SimSun', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: var(--text-color);
  background: var(--bg-color);
  line-height: 1.8;
}

body {
  position: relative;
}

body::before {
  content: '';
  position: fixed;
  inset: 0;
  background-image:
    radial-gradient(circle at 10% 20%, rgba(157, 41, 51, 0.05), transparent 50%),
    radial-gradient(circle at 90% 30%, rgba(26, 54, 93, 0.04), transparent 50%),
    radial-gradient(circle at 30% 80%, rgba(74, 112, 89, 0.03), transparent 50%);
  z-index: -2;
}

body::after {
  content: '';
  position: fixed;
  inset: 0;
  background-image:
    url("data:image/svg+xml,%3Csvg width='100' height='100' viewBox='0 0 100 100' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M11 18c3.866 0 7-3.134 7-7s-3.134-7-7-7-7 3.134-7 7 3.134 7 7 7zm48 25c3.866 0 7-3.134 7-7s-3.134-7-7-7-7 3.134-7 7 3.134 7 7 7zm-43-7c1.657 0 3-1.343 3-3s-1.343-3-3-3-3 1.343-3 3 1.343 3 3 3zm63 31c1.657 0 3-1.343 3-3s-1.343-3-3-3-3 1.343-3 3 1.343 3 3 3zM34 90c1.657 0 3-1.343 3-3s-1.343-3-3-3-3 1.343-3 3 1.343 3 3 3zm56-76c1.657 0 3-1.343 3-3s-1.343-3-3-3-3 1.343-3 3 1.343 3 3 3zM12 86c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm28-65c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm23-11c2.76 0 5-2.24 5-5s-2.24-5-5-5-5 2.24-5 5 2.24 5 5 5zm-6 60c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm28 11c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4z' fill='%239D2933' fill-opacity='0.03' fill-rule='evenodd'/%3E%3C/svg%3E");
  opacity: 0.4;
  z-index: -1;
  pointer-events: none;
}

#app {
  min-height: 100%;
}

::selection {
  background: rgba(157, 41, 51, 0.15);
}

/* 页面切换动画 - 更优雅的淡入淡出 */
.page-enter-active,
.page-leave-active {
  transition: all 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.page-enter-from {
  opacity: 0;
  transform: translateY(20px);
}

.page-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

/* 全局滚动条样式 - 中式卷轴风格 */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: rgba(28, 28, 28, 0.04);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb {
  background: linear-gradient(180deg, var(--cinnabar), var(--indigo));
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(180deg, var(--primary-dark), var(--secondary));
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

/* 中式面板样式 */
.chinese-panel {
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  position: relative;
}

.chinese-panel::before {
  content: '';
  position: absolute;
  inset: 4px;
  border: 1px solid rgba(157, 41, 51, 0.1);
  border-radius: 8px;
  pointer-events: none;
}

/* 中式标签 */
.chinese-tag {
  display: inline-flex;
  align-items: center;
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  background: rgba(255, 255, 255, 0.95);
  color: var(--primary-dark);
  font-weight: 600;
  border-left: 3px solid var(--gold);
  box-shadow: var(--shadow-sm);
}

/* Element Plus 组件样式覆盖 - 中国风 */
.el-button {
  font-weight: 500;
  border-radius: var(--radius-sm);
  transition: all 0.3s ease;
  border: 1px solid transparent;
  font-family: inherit;
  letter-spacing: 0.05em;
}

.el-button + .el-button {
  margin-left: 12px;
}

.el-button--primary {
  background: var(--primary);
  color: #fff;
  border-color: var(--primary);
  box-shadow: var(--shadow-sm);
}

.el-button--primary:focus,
.el-button--primary:hover {
  background: var(--primary-dark);
  border-color: var(--primary-dark);
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.el-button--default {
  background: transparent;
  border-color: var(--card-border);
  color: var(--text-color);
}

.el-button--default:hover {
  border-color: var(--primary);
  color: var(--primary);
  background: rgba(157, 41, 51, 0.05);
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
}

.el-card__header {
  border-bottom: 1px solid rgba(157, 41, 51, 0.1);
  font-weight: 600;
  color: var(--secondary);
}

.el-input__wrapper {
  border-radius: var(--radius-sm);
  border: 1px solid rgba(28, 28, 28, 0.15);
  background: rgba(255, 255, 255, 0.8);
}

.el-input__wrapper:hover {
  border-color: var(--primary);
}

.el-input__wrapper.is-focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(157, 41, 51, 0.1);
}

.el-message {
  border-radius: var(--radius-sm);
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  box-shadow: var(--shadow-md);
}

.el-message--success {
  border-left: 4px solid var(--success);
}

.el-message--error {
  border-left: 4px solid var(--primary);
}

.el-message--warning {
  border-left: 4px solid var(--gold);
}

.el-message--info {
  border-left: 4px solid var(--secondary);
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
