<template>
  <div class="menu-container">
    <header class="menu-hero glass-panel">
      <div class="hero-text">
        <span class="pill">GopherAI Orchestrator</span>
        <h1>集中管理对话、图像与知识</h1>
        <p>
          一处即可访问所有 AI 工具。我们重构了布局与层次，提供更具呼吸感的导航体验，
          让你在不同模块间切换时更顺滑。
        </p>
        <div class="hero-meta">
          <div>
            <strong>02</strong>
            <span>个应用可用</span>
          </div>
          <div>
            <strong>∞</strong>
            <span>对话字数</span>
          </div>
          <div>
            <strong>24/7</strong>
            <span>图像处理</span>
          </div>
        </div>
      </div>
      <div class="hero-actions">
        <div class="status-pill">
          <span class="status-dot"></span>
          <span>登录中</span>
        </div>
        <el-button type="danger" plain @click="handleLogout">退出登录</el-button>
      </div>
    </header>

    <section class="menu-subgrid">
      <article class="summary-card glass-panel">
        <h3>AI 聊天助手</h3>
        <p>跨模型对话 · 流式回复 · 可选 Google 与专家模式</p>
        <button type="button" @click="$router.push('/ai-chat')">立即进入</button>
      </article>
      <article class="summary-card glass-panel">
        <h3>图像识别助手</h3>
        <p>上传图片即可识别 · 多模型算力融合 · 自适应压缩</p>
        <button type="button" @click="$router.push('/image-recognition')">开始识别</button>
      </article>
    </section>

    <main class="menu-main">
      <div class="menu-grid">
        <el-card class="menu-item glass-panel" @click="$router.push('/ai-chat')">
          <div class="card-content">
            <div class="icon-ring chat">
              <el-icon size="36"><ChatDotRound /></el-icon>
            </div>
            <div class="card-text">
              <h3>AI 聊天</h3>
              <p>多会话管理、模型切换、智能检索以及语音播报全都整合完毕。</p>
              <ul>
                <li>支持流式响应</li>
                <li>RAG 检索模式</li>
                <li>接入 Google 搜索</li>
              </ul>
            </div>
          </div>
          <footer class="card-footer">
            <span>最后访问 · 几分钟内</span>
            <span class="cta">进入模块 →</span>
          </footer>
        </el-card>

        <el-card class="menu-item glass-panel" @click="$router.push('/image-recognition')">
          <div class="card-content">
            <div class="icon-ring vision">
              <el-icon size="36"><Camera /></el-icon>
            </div>
            <div class="card-text">
              <h3>图像识别</h3>
              <p>快速上传图片，实时返回识别结果，支持移动端拖拽体验。</p>
              <ul>
                <li>即时上传反馈</li>
                <li>识别历史清晰</li>
                <li>自动保存结果</li>
              </ul>
            </div>
          </div>
          <footer class="card-footer">
            <span>最新模型 · 稳定在线</span>
            <span class="cta">立即使用 →</span>
          </footer>
        </el-card>
      </div>
    </main>
  </div>
</template>

<script>
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ChatDotRound, Camera } from '@element-plus/icons-vue'

export default {
  name: 'MenuView',
  components: {
    ChatDotRound,
    Camera
  },
  setup() {
    const router = useRouter()

    const handleLogout = async () => {
      try {
        await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        localStorage.removeItem('token')
        ElMessage.success('退出登录成功')
        router.push('/login')
      } catch {
        // 用户取消操作
      }
    }

    return {
      handleLogout
    }
  }
}
</script>

<style scoped>
.menu-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  padding: 60px 6vw 80px;
  gap: 32px;
}

.menu-hero {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  padding: 40px 48px;
  position: relative;
  overflow: hidden;
}

.menu-hero::after {
  content: '';
  position: absolute;
  width: 180px;
  height: 180px;
  top: -40px;
  right: -40px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.3), transparent 70%);
}

.hero-text h1 {
  font-size: 40px;
  margin: 18px 0 16px;
  line-height: 1.1;
  color: var(--text-color);
}

.hero-text p {
  max-width: 560px;
}

.hero-meta {
  margin-top: 28px;
  display: flex;
  gap: 32px;
}

.hero-meta strong {
  font-size: 32px;
  display: block;
  line-height: 1;
}

.hero-meta span {
  font-size: 14px;
  color: var(--text-muted);
}

.hero-actions {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  align-items: flex-end;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border-radius: 999px;
  background: rgba(45, 212, 191, 0.12);
  font-weight: 600;
  color: #0f766e;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #0f766e;
}

.menu-subgrid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 20px;
}

.summary-card {
  padding: 28px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.summary-card h3 {
  margin: 0;
  font-size: 22px;
}

.summary-card p {
  margin: 0;
}

.summary-card button {
  align-self: flex-start;
  border: none;
  padding: 10px 18px;
  border-radius: 14px;
  background: rgba(112, 100, 255, 0.1);
  color: var(--primary-dark);
  font-weight: 600;
  cursor: pointer;
}

.summary-card button:hover {
  background: rgba(112, 100, 255, 0.2);
}

.menu-main {
  flex: 1;
}

.menu-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 28px;
}

.menu-item {
  cursor: pointer;
  padding: 30px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  transition: transform 0.25s ease, box-shadow 0.25s ease;
}

.menu-item:hover {
  transform: translateY(-6px);
  box-shadow: 0 40px 60px rgba(15, 23, 42, 0.18);
}

.card-content {
  display: flex;
  gap: 24px;
}

.card-text h3 {
  margin: 0;
  font-size: 24px;
}

.card-text p {
  margin: 10px 0 16px;
}

.card-text ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 8px;
  font-size: 14px;
  color: var(--text-muted);
}

.icon-ring {
  width: 72px;
  height: 72px;
  border-radius: 24px;
  display: grid;
  place-items: center;
  flex-shrink: 0;
  font-size: 34px;
}

.icon-ring.chat {
  background: linear-gradient(135deg, rgba(112, 100, 255, 0.15), rgba(14, 165, 233, 0.2));
  color: #4338ca;
}

.icon-ring.vision {
  background: linear-gradient(135deg, rgba(45, 212, 191, 0.2), rgba(248, 113, 113, 0.15));
  color: #0f766e;
}

.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 28px;
  font-size: 14px;
  color: var(--text-muted);
}

.cta {
  font-weight: 600;
  color: var(--primary);
}

@media (max-width: 880px) {
  .menu-hero {
    flex-direction: column;
    padding: 32px;
  }

  .hero-actions {
    align-items: flex-start;
  }
}

@media (max-width: 640px) {
  .menu-container {
    padding: 48px 16px;
  }

  .card-content {
    flex-direction: column;
  }
}
</style>
