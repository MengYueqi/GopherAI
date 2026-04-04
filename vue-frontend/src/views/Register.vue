<template>
  <div class="auth-page register-page">
    <div class="auth-grid">
      <section class="auth-hero chinese-panel">
        <span class="chinese-tag">开启智慧之旅</span>
        <h1>工欲善其事，必先利其器</h1>
        <p>
          邮箱验证即可完成注册，三步开启您的智能助手服务。
          注册完成后，即可畅享对话、识图、旅行规划等全方位智能体验。
        </p>
        <div class="register-steps">
          <div class="step">
            <span>壹</span>
            <strong>邮箱验证</strong>
          </div>
          <div class="step">
            <span>贰</span>
            <strong>设置密码</strong>
          </div>
          <div class="step">
            <span>叁</span>
            <strong>开启智思</strong>
          </div>
        </div>
      </section>

      <el-card class="auth-card chinese-panel" shadow="never">
        <div class="card-header">
          <h2>注 册</h2>
          <p>请完善以下信息，开启您的智能体验。</p>
        </div>
        <el-form
          ref="registerFormRef"
          :model="registerForm"
          :rules="registerRules"
          label-position="top"
          class="register-form"
        >
          <el-form-item label="邮箱" prop="email">
            <el-input
              v-model="registerForm.email"
              placeholder="请输入邮箱"
              type="email"
            />
          </el-form-item>
          <el-form-item label="验证码" prop="captcha" class="captcha-item">
            <div class="code-row">
              <el-input
                v-model="registerForm.captcha"
                placeholder="请输入验证码"
              />
              <el-button
                type="primary"
                :loading="codeLoading"
                :disabled="countdown > 0"
                @click="sendCode"
              >
                {{ countdown > 0 ? `${countdown}s` : '发送验证码' }}
              </el-button>
            </div>
          </el-form-item>
          <el-form-item label="密码" prop="password">
            <el-input
              v-model="registerForm.password"
              placeholder="请输入密码"
              type="password"
              show-password
            />
          </el-form-item>
          <el-form-item label="确认密码" prop="confirmPassword">
            <el-input
              v-model="registerForm.confirmPassword"
              placeholder="请再次输入密码"
              type="password"
              show-password
            />
          </el-form-item>
          <div class="form-meta">
            <span>密码需至少 6 位，支持字母数字组合。</span>
          </div>
          <el-button
            type="primary"
            :loading="loading"
            @click="handleRegister"
            class="submit-btn"
          >
            注册
          </el-button>
          <div class="form-footer">
            <span>已有账号？</span>
            <button type="button" class="link-btn" @click="$router.push('/login')">
              去登录
            </button>
          </div>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '../utils/api'

export default {
  name: 'RegisterView',
  setup() {
    const router = useRouter()
    const registerFormRef = ref()
    const loading = ref(false)
    const codeLoading = ref(false)
    const countdown = ref(0)

    const registerForm = reactive({
      email: '',
      captcha: '',
      password: '',
      confirmPassword: ''
    })

    const validateConfirmPassword = (rule, value, callback) => {
      if (value !== registerForm.password) {
        callback(new Error('两次输入密码不一致'))
      } else {
        callback()
      }
    }

    const registerRules = {
      email: [
        { required: true, message: '请输入邮箱', trigger: 'blur' },
        { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
      ],
      captcha: [
        { required: true, message: '请输入验证码', trigger: 'blur' }
      ],
      password: [
        { required: true, message: '请输入密码', trigger: 'blur' },
        { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
      ],
      confirmPassword: [
        { required: true, message: '请确认密码', trigger: 'blur' },
        { validator: validateConfirmPassword, trigger: 'blur' }
      ]
    }

    const sendCode = async () => {
      if (!registerForm.email) {
        ElMessage.warning('请先输入邮箱')
        return
      }
      try {
        codeLoading.value = true
        const response = await api.post('/user/captcha', { email: registerForm.email })
        if (response.data.status_code === 1000) {
          ElMessage.success('验证码发送成功')
          countdown.value = 60
          const timer = setInterval(() => {
            countdown.value--
            if (countdown.value <= 0) {
              clearInterval(timer)
            }
          }, 1000)
        } else {
          ElMessage.error(response.data.status_msg || '验证码发送失败')
        }
      } catch (error) {
        console.error('Send code error:', error)
        ElMessage.error('验证码发送失败，请重试')
      } finally {
        codeLoading.value = false
      }
    }

    const handleRegister = async () => {
      try {
        await registerFormRef.value.validate()
        loading.value = true
        const response = await api.post('/user/register', {
              email: registerForm.email,
              captcha: registerForm.captcha,
              password: registerForm.password
        })
        if (response.data.status_code === 1000) {
          ElMessage.success('注册成功，请登录')
          router.push('/login')
        } else {
          ElMessage.error(response.data.status_msg || '注册失败')
        }
      } catch (error) {
        console.error('Register error:', error)
        ElMessage.error('注册失败，请重试')
      } finally {
        loading.value = false
      }
    }

    return {
      registerFormRef,
      loading,
      codeLoading,
      countdown,
      registerForm,
      registerRules,
      sendCode,
      handleRegister
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
  overflow: hidden;
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
  background: linear-gradient(135deg, var(--ochre), var(--cinnabar));
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
  font-size: 34px;
  margin: 20px 0 14px;
  line-height: 1.3;
  font-weight: 600;
  font-family: 'Noto Serif SC', serif;
  letter-spacing: 0.05em;
}

.auth-hero p {
  color: rgba(255, 255, 255, 0.85);
  margin-bottom: 30px;
  line-height: 1.8;
}

.register-steps {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.step {
  flex: 1;
  min-width: 140px;
  padding: 20px;
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.08);
  border-left: 3px solid var(--gold);
  transition: all 0.3s ease;
}

.step:hover {
  background: rgba(255, 255, 255, 0.12);
  transform: translateX(4px);
}

.step span {
  display: inline-flex;
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  align-items: center;
  justify-content: center;
  margin-bottom: 10px;
  background: rgba(255, 255, 255, 0.2);
  font-weight: 600;
  font-family: 'Noto Serif SC', serif;
  font-size: 16px;
}

.step strong {
  display: block;
  font-size: 16px;
  font-weight: 500;
  font-family: 'Noto Serif SC', serif;
}

.auth-card {
  padding: 40px 40px 36px;
  position: relative;
}

.auth-card::after {
  content: '';
  position: absolute;
  top: 20px;
  right: 20px;
  width: 60px;
  height: 60px;
  background: url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M30 5C16.2 5 5 16.2 5 30s11.2 25 25 25 25-11.2 25-25S43.8 5 30 5zm0 45c-11 0-20-9-20-20s9-20 20-20 20 9 20 20-9 20-20 20z' fill='%238C4320' fill-opacity='0.1'/%3E%3Cpath d='M30 15c-8.3 0-15 6.7-15 15s6.7 15 15 15 15-6.7 15-15-6.7-15-15-15zm0 25c-5.5 0-10-4.5-10-10s4.5-10 10-10 10 4.5 10 10-4.5 10-10 10z' fill='%238C4320' fill-opacity='0.1'/%3E%3C/svg%3E");
  opacity: 0.3;
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

.register-form {
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

.code-row {
  display: flex;
  gap: 12px;
  align-items: center;
}

.code-row .el-button {
  min-width: 140px;
  height: 48px;
  border-radius: var(--radius-sm);
  font-family: 'Noto Serif SC', serif;
}

.form-meta {
  font-size: 13px;
  color: var(--text-muted);
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
  .auth-page {
    padding: 40px 16px;
  }

  .auth-hero {
    padding: 36px 28px;
  }

  .code-row {
    flex-direction: column;
  }

  .code-row .el-button {
    width: 100%;
  }
}
</style>
