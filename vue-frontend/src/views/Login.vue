<template>
  <div class="auth-page login-page">
    <span class="blob blob-one"></span>
    <span class="blob blob-two"></span>

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
            />
          </el-form-item>
          <div class="form-meta">
            <span>账号受企业级安全保护</span>
          </div>
          <el-button
            type="primary"
            :loading="loading"
            @click="handleLogin"
            class="submit-btn"
          >
            登录
          </el-button>
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
    const loginForm = ref({
      username: '',
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

    return {
      loginFormRef,
      loading,
      loginForm,
      loginRules,
      handleLogin
    }
}
}
</script>

<style scoped>
.blob {
  position: absolute;
  width: 320px;
  height: 320px;
  border-radius: 50%;
  filter: blur(60px);
  opacity: 0.45;
}

.blob-one {
  top: 5%;
  right: 12%;
  background: #c4b5fd;
}

.blob-two {
  bottom: 8%;
  left: 8%;
  background: #5eead4;
}

.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 60px 5vw;
  position: relative;
  overflow: hidden;
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
  background: linear-gradient(140deg, rgba(20, 18, 50, 0.95), rgba(62, 66, 168, 0.95));
  position: relative;
  overflow: hidden;
}

.auth-hero::after {
  content: '';
  position: absolute;
  width: 180px;
  height: 180px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(59, 130, 246, 0.35), transparent 70%);
  right: -40px;
  top: -30px;
}

.auth-hero h1 {
  font-size: 36px;
  line-height: 1.2;
  margin: 20px 0 16px;
  font-weight: 700;
}

.auth-hero p {
  color: rgba(255, 255, 255, 0.75);
  margin-bottom: 28px;
}

.auth-highlights {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 18px;
}

.auth-highlights li {
  padding: 18px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.auth-highlights strong {
  display: block;
  font-size: 16px;
  margin-bottom: 6px;
}

.auth-highlights span {
  color: rgba(255, 255, 255, 0.7);
  font-size: 14px;
}

.auth-card {
  padding: 36px 32px 32px;
}

.card-header {
  margin-bottom: 28px;
}

.card-header h2 {
  margin: 0;
  font-size: 28px;
  font-weight: 700;
}

.card-header p {
  margin: 8px 0 0;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.el-form-item {
  margin-bottom: 0;
}

.el-form-item :deep(.el-form-item__label) {
  font-weight: 600;
  color: var(--text-color);
}

.el-input :deep(.el-input__wrapper) {
  border-radius: 16px;
  border: 1px solid rgba(15, 23, 42, 0.08);
  background: rgba(255, 255, 255, 0.9);
  padding: 0 14px;
  box-shadow: inset 0 0 0 1px transparent;
  transition: all 0.2s ease;
}

.el-input :deep(.el-input__wrapper.is-focus) {
  border-color: rgba(112, 100, 255, 0.85);
  box-shadow: 0 0 0 3px rgba(112, 100, 255, 0.2);
}

.form-meta {
  font-size: 13px;
  color: var(--text-muted);
}

.submit-btn {
  width: 100%;
  height: 50px;
  border-radius: 18px;
  font-size: 16px;
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
}
</style>
