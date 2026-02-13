<template>
  <div class="auth-page login-page">
    <div class="auth-grid">
      <section class="auth-hero glass-panel">
        <span class="pill">GopherAI Suite</span>
        <h1>欢迎回来，指挥你的智能助手</h1>
        <p>
          统一的 AI 控制台，随时掌控聊天、图像识别与知识检索。我们重新优化了视觉系统，
          让每一次操作都充满现代感。
        </p>
        <ul class="auth-highlights">
          <li>
            <strong>实时响应</strong>
            <span>多模型即时切换，毫秒级反馈。</span>
          </li>
          <li>
            <strong>安全隔离</strong>
            <span>多重验证与加密守护数据。</span>
          </li>
          <li>
            <strong>跨端体验</strong>
            <span>自适应设计，桌面与移动一致。</span>
          </li>
        </ul>
      </section>

      <el-card class="auth-card glass-panel" shadow="never">
        <div class="card-topline">
          <span class="status-dot"></span>
          <span>企业级安全登录</span>
        </div>
        <div class="card-header">
          <h2>登录账户</h2>
          <p>输入账号信息，继续探索 AI 能力。</p>
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
  background: #f8fafc;
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
  background: #1e1b4b;
  position: relative;
}

.auth-hero .pill {
  display: inline-block;
  padding: 6px 14px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 600;
  color: #c4b5fd;
  background: rgba(139, 92, 246, 0.2);
  border: 1px solid rgba(139, 92, 246, 0.3);
}

.auth-hero h1 {
  font-size: 38px;
  line-height: 1.25;
  margin: 24px 0 18px;
  font-weight: 700;
  color: #fff;
}

.auth-hero p {
  color: rgba(255, 255, 255, 0.7);
  margin-bottom: 32px;
  font-size: 15px;
  line-height: 1.7;
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
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.12);
  transition: background 0.2s ease;
}

.auth-highlights li:hover {
  background: rgba(255, 255, 255, 0.12);
}

.auth-highlights strong {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  margin-bottom: 8px;
  color: #fff;
}

.auth-highlights strong::before {
  content: '';
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #8b5cf6;
}

.auth-highlights span {
  color: rgba(255, 255, 255, 0.6);
  font-size: 14px;
  padding-left: 18px;
}

.auth-card {
  padding: 36px 36px 32px;
  border-radius: 16px;
  border: 1px solid #e2e8f0;
  background: #fff;
}

.card-topline {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  margin-bottom: 20px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  color: #334155;
  background: #f1f5f9;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #16a34a;
}

.card-header {
  margin-bottom: 28px;
}

.card-header h2 {
  margin: 0;
  font-size: 32px;
  font-weight: 700;
  color: #1e293b;
}

.card-header p {
  margin: 10px 0 0;
  color: #64748b;
  font-size: 15px;
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
  font-weight: 600;
  font-size: 14px;
  color: #334155;
  padding-bottom: 8px;
}

.el-input :deep(.el-input__wrapper) {
  border-radius: 12px;
  border: 1px solid #d1d5db;
  background: #fff;
  padding: 4px 16px;
  transition: border-color 0.2s ease;
  height: 48px;
}

.el-input :deep(.el-input__inner) {
  font-size: 15px;
}

.el-input :deep(.el-input__wrapper.is-focus) {
  border-color: #6366f1;
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
  border-radius: 12px;
  font-size: 16px;
  font-weight: 600;
  margin-top: 8px;
  background: #6366f1;
  border: none;
  transition: background 0.2s ease;
}

.submit-btn:hover {
  background: #4f46e5;
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
