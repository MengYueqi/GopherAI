<template>
  <div class="auth-page login-page">
    <div class="auth-grid">
      <section class="auth-hero chinese-panel">
        <span class="chinese-tag">GopherAI · 智思集</span>
        <h1>有朋自远方来，不亦乐乎</h1>
        <p>
          东方智慧与现代科技的完美融合，集聊天、识图、规划于一体的智能平台。
          以匠心之作，为您带来前所未有的智能体验。
        </p>
        <ul class="auth-highlights">
          <li>
            <strong>思如涌泉</strong>
            <span>多模型智慧融合，应答如流。</span>
          </li>
          <li>
            <strong>固若金汤</strong>
            <span>多重加密防护，数据安全无忧。</span>
          </li>
          <li>
            <strong>浑然一体</strong>
            <span>跨平台自适应设计，体验如一。</span>
          </li>
        </ul>
      </section>

      <el-card class="auth-card chinese-panel" shadow="never">
        <div class="card-topline">
          <span class="status-dot"></span>
          <span>安全登堂入室</span>
        </div>
        <div class="card-header">
          <h2>登 录</h2>
          <p>请输入账号信息，开启智能之旅。</p>
        </div>
        <el-form
          ref="loginFormRef"
          :model="loginForm"
          :rules="loginRules"
          label-position="top"
          class="login-form"
        >
          <el-form-item label="用户名" prop="username">
            <el-input
              v-model="loginForm.username"
              placeholder="请输入用户名"
            />
          </el-form-item>
          <el-form-item label="密码" prop="password">
            <el-input
              v-model="loginForm.password"
              placeholder="请输入密码"
              type="password"
              show-password
              @keyup.enter="handleLogin"
            />
          </el-form-item>
          <div class="form-options">
            <el-checkbox v-model="rememberMe">记住我</el-checkbox>
            <button type="button" class="link-btn subtle-link" @click="handleForgotPassword">
              忘记密码
            </button>
          </div>
          <div class="form-meta">
            <span>登录即代表你同意平台的安全策略与服务条款</span>
          </div>
          <el-button
            type="primary"
            :loading="loading"
            @click="handleLogin"
            class="submit-btn"
          >
            {{ loading ? '登录中...' : '立即登录' }}
          </el-button>
          <div class="divider-text">或</div>
          <div class="form-footer">
            <span>还没有账号？</span>
            <button type="button" class="link-btn" @click="$router.push('/register')">
              立即注册
            </button>
          </div>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '../utils/api'

export default {
  name: 'LoginView',
  setup() {
    const router = useRouter()
    const loginFormRef = ref()
    const loading = ref(false)
    const savedUsername = localStorage.getItem('saved_username') || ''
    const rememberMe = ref(Boolean(savedUsername))
    const loginForm = ref({
      username: savedUsername,
      password: ''
    })

    const loginRules = {
      username: [
        { required: true, message: '请输入用户名', trigger: 'blur' }
      ],
      password: [
        { required: true, message: '请输入密码', trigger: 'blur' },
        { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
      ]
    }

    const handleLogin = async () => {
      try {
        await loginFormRef.value.validate()
        loading.value = true
        const response = await api.post('/user/login', {
          username: loginForm.value.username,
          password: loginForm.value.password
        })
        if (response.data.status_code === 1000) {
          localStorage.setItem('token', response.data.token)
          if (rememberMe.value) {
            localStorage.setItem('saved_username', loginForm.value.username)
          } else {
            localStorage.removeItem('saved_username')
          }
          ElMessage.success('登录成功')
          router.push('/menu')
        } else {
          ElMessage.error(response.data.status_msg || '登录失败')
        }
      } catch (error) {
        console.error('Login error:', error)
        ElMessage.error('登录失败，请重试')
      } finally {
        loading.value = false
      }
    }

    const handleForgotPassword = () => {
      ElMessage.info('请联系管理员重置密码')
    }

    return {
      loginFormRef,
      loading,
      rememberMe,
      loginForm,
      loginRules,
      handleLogin,
      handleForgotPassword
    }
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 60px 5vw;
  position: relative;
  background: var(--bg-color);
}

.auth-grid {
  width: min(1180px, 100%);
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 32px;
  position: relative;
  z-index: 1;
}

.auth-hero {
  padding: 48px;
  color: #fff;
  background: linear-gradient(135deg, var(--indigo), var(--pine));
  position: relative;
  overflow: hidden;
}

.auth-hero::before {
  content: '';
  position: absolute;
  right: -20%;
  top: -20%;
  width: 300px;
  height: 300px;
  background: url("data:image/svg+xml,%3Csvg width='300' height='300' viewBox='0 0 300 300' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M150 30C80 30 30 80 30 150s50 120 120 120 120-50 120-120S220 30 150 30zm0 200c-44.2 0-80-35.8-80-80s35.8-80 80-80 80 35.8 80 80-35.8 80-80 80z' fill='%23fff' fill-opacity='0.05'/%3E%3Cpath d='M150 70c-44.2 0-80 35.8-80 80s35.8 80 80 80 80-35.8 80-80-35.8-80-80-80zm0 140c-33.1 0-60-26.9-60-60s26.9-60 60-60 60 26.9 60 60-26.9 60-60 60z' fill='%23fff' fill-opacity='0.05'/%3E%3C/svg%3E");
  opacity: 0.3;
}

.auth-hero h1 {
  font-size: 36px;
  line-height: 1.3;
  margin: 24px 0 18px;
  font-weight: 600;
  color: #fff;
  font-family: 'Noto Serif SC', serif;
  letter-spacing: 0.05em;
}

.auth-hero p {
  color: rgba(255, 255, 255, 0.8);
  margin-bottom: 32px;
  font-size: 15px;
  line-height: 1.8;
}

.auth-highlights {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 16px;
}

.auth-highlights li {
  padding: 20px;
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.08);
  border-left: 3px solid var(--gold);
  transition: all 0.3s ease;
}

.auth-highlights li:hover {
  background: rgba(255, 255, 255, 0.12);
  transform: translateX(4px);
}

.auth-highlights strong {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 17px;
  margin-bottom: 8px;
  color: #fff;
  font-weight: 500;
  font-family: 'Noto Serif SC', serif;
}

.auth-highlights strong::before {
  content: '•';
  color: var(--gold);
  font-size: 20px;
}

.auth-highlights span {
  color: rgba(255, 255, 255, 0.75);
  font-size: 14px;
  padding-left: 28px;
  display: block;
}

.auth-card {
  padding: 40px 40px 36px;
  border-radius: var(--radius-lg);
  position: relative;
}

.auth-card::after {
  content: '';
  position: absolute;
  top: 20px;
  right: 20px;
  width: 60px;
  height: 60px;
  background: url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M30 5C16.2 5 5 16.2 5 30s11.2 25 25 25 25-11.2 25-25S43.8 5 30 5zm0 45c-11 0-20-9-20-20s9-20 20-20 20 9 20 20-9 20-20 20z' fill='%239D2933' fill-opacity='0.1'/%3E%3Cpath d='M30 15c-8.3 0-15 6.7-15 15s6.7 15 15 15 15-6.7 15-15-6.7-15-15-15zm0 25c-5.5 0-10-4.5-10-10s4.5-10 10-10 10 4.5 10 10-4.5 10-10 10z' fill='%239D2933' fill-opacity='0.1'/%3E%3C/svg%3E");
  opacity: 0.3;
}

.card-topline {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  margin-bottom: 24px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 500;
  color: var(--jade);
  background: rgba(74, 112, 89, 0.1);
  border-left: 3px solid var(--jade);
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--jade);
}

.card-header {
  margin-bottom: 32px;
}

.card-header h2 {
  margin: 0;
  font-size: 36px;
  font-weight: 600;
  color: var(--ink);
  font-family: 'Noto Serif SC', serif;
  letter-spacing: 0.2em;
  text-align: center;
}

.card-header p {
  margin: 12px 0 0;
  color: var(--text-muted);
  font-size: 15px;
  text-align: center;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.el-form-item {
  margin-bottom: 0;
}

.el-form-item :deep(.el-form-item__label) {
  font-weight: 500;
  font-size: 14px;
  color: var(--secondary);
  padding-bottom: 8px;
  font-family: 'Noto Serif SC', serif;
}

.el-input :deep(.el-input__wrapper) {
  border-radius: var(--radius-sm);
  border: 1px solid rgba(28, 28, 28, 0.15);
  background: rgba(255, 255, 255, 0.8);
  padding: 4px 16px;
  transition: all 0.3s ease;
  height: 48px;
}

.el-input :deep(.el-input__inner) {
  font-size: 15px;
  font-family: inherit;
}

.el-input :deep(.el-input__wrapper.is-focus) {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(157, 41, 51, 0.08);
}

.form-meta {
  font-size: 13px;
  color: var(--text-muted);
  margin-top: -4px;
}

.form-options {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: -6px;
}

.form-options :deep(.el-checkbox__label) {
  color: #475569;
}

.subtle-link {
  color: #64748b;
}

.subtle-link:hover {
  color: #475569;
}

.submit-btn {
  width: 100%;
  height: 54px;
  border-radius: var(--radius-sm);
  font-size: 16px;
  font-weight: 500;
  margin-top: 8px;
  font-family: 'Noto Serif SC', serif;
  letter-spacing: 0.1em;
  position: relative;
  overflow: hidden;
}

.submit-btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
  transition: left 0.5s ease;
}

.submit-btn:hover::before {
  left: 100%;
}

.divider-text {
  text-align: center;
  font-size: 12px;
  color: #94a3b8;
  position: relative;
}

.divider-text::before,
.divider-text::after {
  content: '';
  position: absolute;
  top: 50%;
  width: calc(50% - 20px);
  height: 1px;
  background: rgba(148, 163, 184, 0.35);
}

.divider-text::before {
  left: 0;
}

.divider-text::after {
  right: 0;
}

.form-footer {
  display: flex;
  justify-content: center;
  gap: 8px;
  font-size: 14px;
  color: var(--text-muted);
}

.link-btn {
  border: none;
  background: none;
  color: var(--primary);
  font-weight: 600;
  cursor: pointer;
  padding: 0;
}

.link-btn:hover {
  color: var(--primary-dark);
}

@media (max-width: 720px) {
  .auth-hero {
    padding: 36px 28px;
  }

  .auth-page {
    padding: 40px 16px;
  }

  .auth-card {
    padding: 28px 24px;
  }
}
</style>
