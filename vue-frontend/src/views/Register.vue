<template>
  <div class="auth-page register-page">
    <span class="blob blob-one"></span>
    <span class="blob blob-two"></span>

    <div class="auth-grid">
      <section class="auth-hero glass-panel">
        <span class="pill">Create Workspace</span>
        <h1>立即启用你的 AI 工作室</h1>
        <p>
          激活邮箱即可完成注册。我们通过验证码确认身份，确保数据更安全。
          完成注册后，即可在统一的控制面板中管理聊天、图像识别与私有知识库。
        </p>
        <div class="register-steps">
          <div class="step">
            <span>01</span>
            <strong>验证邮箱</strong>
          </div>
          <div class="step">
            <span>02</span>
            <strong>设置安全密码</strong>
          </div>
          <div class="step">
            <span>03</span>
            <strong>绑定你的工作区</strong>
          </div>
        </div>
      </section>

      <el-card class="auth-card glass-panel" shadow="never">
        <div class="card-header">
          <h2>注册全新账号</h2>
          <p>精准的字段校验与验证码防护，确保流程可信。</p>
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
}

.blob {
  position: absolute;
  width: 360px;
  height: 360px;
  border-radius: 50%;
  filter: blur(70px);
  opacity: 0.4;
}

.blob-one {
  top: -40px;
  left: 10%;
  background: #c084fc;
}

.blob-two {
  bottom: -40px;
  right: 12%;
  background: #34d399;
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
  background: linear-gradient(135deg, rgba(12, 15, 40, 0.95), rgba(48, 56, 126, 0.95));
}

.auth-hero h1 {
  font-size: 34px;
  margin: 20px 0 14px;
  line-height: 1.2;
}

.auth-hero p {
  color: rgba(255, 255, 255, 0.78);
  margin-bottom: 30px;
}

.register-steps {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.step {
  flex: 1;
  min-width: 140px;
  padding: 18px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.step span {
  display: inline-flex;
  width: 32px;
  height: 32px;
  border-radius: 12px;
  align-items: center;
  justify-content: center;
  margin-bottom: 10px;
  background: rgba(255, 255, 255, 0.2);
  font-weight: 600;
}

.step strong {
  display: block;
  font-size: 16px;
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

.register-form {
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
  transition: all 0.2s ease;
}

.el-input :deep(.el-input__wrapper.is-focus) {
  border-color: rgba(112, 100, 255, 0.85);
  box-shadow: 0 0 0 3px rgba(112, 100, 255, 0.2);
}

.code-row {
  display: flex;
  gap: 12px;
  align-items: center;
}

.code-row .el-button {
  min-width: 140px;
  height: 48px;
  border-radius: 16px;
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
