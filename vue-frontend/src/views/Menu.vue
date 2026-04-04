<template>
  <div class="menu-container">
    <header class="menu-hero chinese-panel">
      <div class="hero-text">
        <span class="chinese-tag">智思集 · 中枢</span>
        <h1>运筹帷幄之中，决胜千里之外</h1>
        <p>
          东方智慧与现代科技的完美融合，集对话、识图、旅行规划于一体的智能平台。
          以匠心之作，为您带来前所未有的智能体验。
        </p>
        <div class="hero-meta">
          <div>
            <strong>三</strong>
            <span>大能齐备</span>
          </div>
          <div>
            <strong>∞</strong>
            <span>思如涌泉</span>
          </div>
          <div>
            <strong>恒</strong>
            <span>无时不在</span>
          </div>
        </div>
      </div>
      <div class="hero-actions">
        <div class="status-pill">
          <span class="status-dot"></span>
          <span>已登入</span>
        </div>
        <el-button type="primary" @click="handleLogout">退出登录</el-button>
      </div>
    </header>

    <section class="menu-subgrid">
      <article class="summary-card chinese-panel">
        <h3>知言 · 对话</h3>
        <p>多模型智慧融合 · 流式应答 · 支持联网检索与专家模式</p>
        <button type="button" @click="$router.push('/ai-chat')">入 内</button>
      </article>
      <article class="summary-card chinese-panel">
        <h3>知形 · 识图</h3>
        <p>上传即识 · 多算力融合 · 智能压缩优化</p>
        <button type="button" @click="$router.push('/image-recognition')">启 用</button>
      </article>
      <article class="summary-card chinese-panel">
        <h3>知行 · 行旅</h3>
        <p>输入出发地、目的地与偏好，生成完备的旅行行程与出行建议。</p>
        <button type="button" @click="$router.push('/travel-planning')">开 始</button>
      </article>
    </section>

    <main class="menu-main">
      <div class="menu-grid">
        <el-card class="menu-item chinese-panel" @click="$router.push('/ai-chat')">
          <div class="card-content">
            <div class="icon-ring chat">
              <el-icon size="36"><ChatDotRound /></el-icon>
            </div>
            <div class="card-text">
              <h3>知言 · 对话</h3>
              <p>多会话管理、模型切换、智能检索以及语音播报功能完备。</p>
              <ul>
                <li>流式应答，响应如流</li>
                <li>RAG 检索，博闻强识</li>
                <li>联网搜索，知无不言</li>
              </ul>
            </div>
          </div>
          <footer class="card-footer">
            <span>最近访问 · 方才</span>
            <span class="cta">进入 →</span>
          </footer>
        </el-card>

        <el-card class="menu-item chinese-panel" @click="$router.push('/image-recognition')">
          <div class="card-content">
            <div class="icon-ring vision">
              <el-icon size="36"><Camera /></el-icon>
            </div>
            <div class="card-text">
              <h3>知形 · 识图</h3>
              <p>上传图片即可识别，实时返回结果，支持移动端拖拽操作。</p>
              <ul>
                <li>上传即得，立等可取</li>
                <li>历史记录，一目了然</li>
                <li>自动保存，后顾无忧</li>
              </ul>
            </div>
          </div>
          <footer class="card-footer">
            <span>最新模型 · 稳如泰山</span>
            <span class="cta">使用 →</span>
          </footer>
        </el-card>

        <el-card class="menu-item chinese-panel" @click="$router.push('/travel-planning')">
          <div class="card-content">
            <div class="icon-ring travel">
              <el-icon size="36"><MapLocation /></el-icon>
            </div>
            <div class="card-text">
              <h3>知行 · 行旅</h3>
              <p>描述出行需求、预算与偏好，智能生成分天行程与交通建议。</p>
              <ul>
                <li>行程规划，顷刻而成</li>
                <li>交通食宿，安排妥当</li>
                <li>个性定制，称心如意</li>
              </ul>
            </div>
          </div>
          <footer class="card-footer">
            <span>旅途规划 · 一气呵成</span>
            <span class="cta">出发 →</span>
          </footer>
        </el-card>
      </div>
    </main>
  </div>
</template>

<script>
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ChatDotRound, Camera, MapLocation } from '@element-plus/icons-vue'

export default {
  name: 'MenuView',
  components: {
    ChatDotRound,
    Camera,
    MapLocation
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
  background: var(--bg-color);
}

.menu-hero {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  padding: 48px 56px;
  position: relative;
  overflow: hidden;
  background: linear-gradient(135deg, var(--ivory), var(--rice-paper));
}

.menu-hero::before {
  content: '';
  position: absolute;
  right: -10%;
  top: -10%;
  width: 280px;
  height: 280px;
  background: url("data:image/svg+xml,%3Csvg width='280' height='280' viewBox='0 0 280 280' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M140 20C70 20 20 70 20 140s50 120 120 120 120-50 120-120S210 20 140 20zm0 200c-44.2 0-80-35.8-80-80s35.8-80 80-80 80 35.8 80 80-35.8 80-80 80z' fill='%239D2933' fill-opacity='0.03'/%3E%3Cpath d='M140 60c-44.2 0-80 35.8-80 80s35.8 80 80 80 80-35.8 80-80-35.8-80-80-80zm0 140c-33.1 0-60-26.9-60-60s26.9-60 60-60 60 26.9 60 60-26.9 60-60 60z' fill='%239D2933' fill-opacity='0.03'/%3E%3C/svg%3E");
  opacity: 0.5;
}

.hero-text h1 {
  font-size: 42px;
  margin: 18px 0 16px;
  line-height: 1.2;
  color: var(--ink);
  font-family: 'Noto Serif SC', serif;
  font-weight: 600;
  letter-spacing: 0.05em;
}

.hero-text p {
  max-width: 560px;
  line-height: 1.8;
}

.hero-meta {
  margin-top: 32px;
  display: flex;
  gap: 48px;
}

.hero-meta strong {
  font-size: 36px;
  display: block;
  line-height: 1;
  color: var(--primary);
  font-family: 'Noto Serif SC', serif;
  font-weight: 600;
}

.hero-meta span {
  font-size: 14px;
  color: var(--text-muted);
  font-family: 'Noto Serif SC', serif;
  margin-top: 6px;
  display: block;
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
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  background: rgba(74, 112, 89, 0.1);
  font-weight: 500;
  color: var(--jade);
  border-left: 3px solid var(--jade);
  font-family: 'Noto Serif SC', serif;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--jade);
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
  transition: all 0.3s ease;
  border-left: 3px solid transparent;
}

.summary-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
  border-left-color: var(--primary);
}

.summary-card h3 {
  margin: 0;
  font-size: 22px;
  font-family: 'Noto Serif SC', serif;
  color: var(--secondary);
  font-weight: 600;
}

.summary-card p {
  margin: 0;
  line-height: 1.7;
}

.summary-card button {
  align-self: flex-start;
  border: 1px solid var(--primary);
  padding: 8px 20px;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--primary);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  font-family: 'Noto Serif SC', serif;
  letter-spacing: 0.1em;
  margin-top: 8px;
}

.summary-card button:hover {
  background: var(--primary);
  color: #fff;
  transform: translateY(-2px);
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
  padding: 32px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.menu-item::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 4px;
  height: 0;
  background: var(--primary);
  transition: height 0.3s ease;
}

.menu-item:hover {
  transform: translateY(-8px);
  box-shadow: var(--shadow-lg);
}

.menu-item:hover::before {
  height: 100%;
}

.card-content {
  display: flex;
  gap: 24px;
}

.card-text h3 {
  margin: 0;
  font-size: 24px;
  font-family: 'Noto Serif SC', serif;
  color: var(--secondary);
  font-weight: 600;
}

.card-text p {
  margin: 10px 0 16px;
  line-height: 1.7;
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

.card-text li::before {
  content: '• ';
  color: var(--primary);
  font-weight: bold;
}

.icon-ring {
  width: 72px;
  height: 72px;
  border-radius: var(--radius-md);
  display: grid;
  place-items: center;
  flex-shrink: 0;
  font-size: 34px;
  border: 1px solid rgba(28, 28, 28, 0.05);
}

.icon-ring.chat {
  background: rgba(157, 41, 51, 0.1);
  color: var(--primary);
}

.icon-ring.vision {
  background: rgba(74, 112, 89, 0.1);
  color: var(--jade);
}

.icon-ring.travel {
  background: rgba(26, 54, 93, 0.1);
  color: var(--secondary);
}

.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 32px;
  font-size: 14px;
  color: var(--text-muted);
  padding-top: 16px;
  border-top: 1px solid rgba(28, 28, 28, 0.06);
}

.cta {
  font-weight: 500;
  color: var(--primary);
  font-family: 'Noto Serif SC', serif;
  transition: transform 0.3s ease;
}

.menu-item:hover .cta {
  transform: translateX(4px);
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
