<template>
  <div class="image-recognition-container">
    <aside class="vision-sidebar glass-panel">
      <span class="pill">Vision Lab</span>
      <h1>图像识别助手</h1>
      <p>
        重新绘制的界面为图像识别打造更沉浸的体验。拖拽或点击上传即可完成解析，
        历史记录与 AI 回复都保留在右侧的实时流中。
      </p>
      <ul class="vision-highlights">
        <li>
          <strong>全格式支持</strong>
          <span>PNG / JPG / WebP / HEIC 等主流格式。</span>
        </li>
        <li>
          <strong>安全上传</strong>
          <span>仅当前会话可访问的临时 URL。</span>
        </li>
        <li>
          <strong>实时结果</strong>
          <span>识别成功后立即推送到消息流。</span>
        </li>
      </ul>
    </aside>

    <section class="vision-console glass-panel">
      <div class="top-bar">
        <div>
          <h2>识别时间轴</h2>
          <p>每次识别都会记录在此，方便快速对比。</p>
        </div>
        <button class="back-btn" @click="$router.push('/menu')">回到菜单</button>
      </div>

      <div class="chat-messages" ref="chatContainerRef">
        <div
          v-for="(message, index) in messages"
          :key="index"
          :class="['message', message.role === 'user' ? 'user-message' : 'ai-message']"
        >
          <div class="message-header">
            <b>{{ message.role === 'user' ? '你' : 'AI' }}:</b>
          </div>
          <div class="message-content">
            <span>{{ message.content }}</span>
            <img v-if="message.imageUrl" :src="message.imageUrl" alt="上传的图片" />
          </div>
        </div>
      </div>

      <div class="chat-input">
        <form @submit.prevent="handleSubmit">
          <label class="upload-drop" for="vision-upload">
            <input
              id="vision-upload"
              ref="fileInputRef"
              type="file"
              accept="image/*"
              required
              @change="handleFileSelect"
            />
            <div class="drop-icon">＋</div>
            <div>
              <strong>拖拽或点击上传</strong>
              <p>最大 10MB，图片仅用于当前识别</p>
            </div>
            <span class="file-name" v-if="selectedFile">{{ selectedFile.name }}</span>
          </label>
          <div class="upload-actions">
            <span>支持批量识别，请逐张上传以确保准确率。</span>
            <button type="submit" :disabled="!selectedFile">发送图片</button>
          </div>
        </form>
      </div>
    </section>
  </div>
</template>

<script>
import { ref, nextTick } from 'vue'
import api from '../utils/api'

export default {
  name: 'ImageRecognition',
  setup() {
    const messages = ref([])
    const selectedFile = ref(null)
    const fileInputRef = ref()
    const chatContainerRef = ref()

    const handleFileSelect = (event) => {
      selectedFile.value = event.target.files[0]
    }

    const handleSubmit = async () => {
      if (!selectedFile.value) return

      const file = selectedFile.value
      const imageUrl = URL.createObjectURL(file)

      // Add user message to UI
      messages.value.push({
        role: 'user',
        content: `已上传图片: ${file.name}`,
        imageUrl: imageUrl,
      })

      await nextTick()
      scrollToBottom()

      // Create FormData
      const formData = new FormData()
      formData.append('image', file)

      try {
        const response = await api.post('/image/recognize', formData, {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        })


        if (response.data && response.data.class_name) {
             const aiText = `识别结果: ${response.data.class_name}`
            messages.value.push({
                role: 'assistant',
                content: aiText,
            })
        } else {
             messages.value.push({
                 role: 'assistant',
                 content: `[错误] ${response.data.status_msg || '识别失败'}`,
             })
        }
      } catch (error) {
        console.error('Upload error:', error)
        messages.value.push({
          role: 'assistant',
          content: `[错误] 无法连接到服务器或上传失败: ${error.message}`,
        })
      } finally {

        URL.revokeObjectURL(imageUrl)

            await nextTick()
        scrollToBottom()


        selectedFile.value = null
        if (fileInputRef.value) {
          fileInputRef.value.value = ''
        }
      }
    }

    const scrollToBottom = () => {
      if (chatContainerRef.value) {
        chatContainerRef.value.scrollTop = chatContainerRef.value.scrollHeight
      }
    }

    return {
      messages,
      selectedFile,
      fileInputRef,
      chatContainerRef,
      handleFileSelect,
      handleSubmit
    }
  }
}
</script>

<style scoped>
.image-recognition-container {
  min-height: 100vh;
  display: grid;
  grid-template-columns: minmax(260px, 360px) 1fr;
  gap: 32px;
  padding: 48px 5vw 64px;
}

.vision-sidebar {
  padding: 48px;
  color: #fff;
  background: linear-gradient(160deg, rgba(14, 23, 63, 0.95), rgba(49, 63, 166, 0.9));
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.vision-sidebar h1 {
  margin: 0;
  font-size: 36px;
}

.vision-sidebar p {
  color: rgba(255, 255, 255, 0.75);
}

.vision-highlights {
  list-style: none;
  padding: 0;
  margin: 8px 0 0;
  display: grid;
  gap: 18px;
}

.vision-highlights li {
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 18px;
  padding: 18px;
}

.vision-highlights strong {
  display: block;
  margin-bottom: 6px;
}

.vision-highlights span {
  color: rgba(255, 255, 255, 0.72);
  font-size: 14px;
}

.vision-console {
  border-radius: 32px;
  background: rgba(255, 255, 255, 0.95);
  border: 1px solid rgba(99, 102, 241, 0.08);
  box-shadow: 0 35px 80px rgba(15, 23, 42, 0.15);
  display: flex;
  flex-direction: column;
}

.top-bar {
  padding: 32px 40px 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
  border-bottom: 1px solid rgba(15, 23, 42, 0.06);
}

.top-bar h2 {
  margin: 0;
  font-size: 26px;
}

.top-bar p {
  margin: 6px 0 0;
  color: var(--text-muted);
}

.back-btn {
  border: none;
  border-radius: 18px;
  padding: 12px 24px;
  background: linear-gradient(120deg, var(--primary), var(--secondary));
  color: #fff;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 15px 30px rgba(112, 100, 255, 0.3);
}

.chat-messages {
  flex: 1;
  padding: 32px 40px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  overflow-y: auto;
  background: linear-gradient(180deg, rgba(248, 249, 255, 0.8), rgba(235, 240, 255, 0.6));
}

.chat-messages::-webkit-scrollbar {
  width: 6px;
}

.chat-messages::-webkit-scrollbar-thumb {
  background: rgba(15, 23, 42, 0.2);
  border-radius: 20px;
}

.message {
  max-width: 680px;
  padding: 18px 22px;
  border-radius: 20px;
  line-height: 1.6;
  box-shadow: 0 18px 30px rgba(15, 23, 42, 0.12);
  border: 1px solid rgba(15, 23, 42, 0.08);
  background: rgba(255, 255, 255, 0.95);
}

.user-message {
  align-self: flex-end;
  background: linear-gradient(135deg, #6366f1, #3bbef3);
  color: #fff;
}

.message-content {
  white-space: pre-wrap;
}

.message-content img {
  max-width: 320px;
  border-radius: 18px;
  margin-top: 14px;
  box-shadow: 0 18px 38px rgba(15, 23, 42, 0.25);
}

.chat-input {
  padding: 28px 40px 36px;
  border-top: 1px solid rgba(15, 23, 42, 0.06);
  background: rgba(255, 255, 255, 0.9);
}

.chat-input form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.upload-drop {
  border: 2px dashed rgba(15, 23, 42, 0.15);
  border-radius: 24px;
  padding: 24px;
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 18px;
  align-items: center;
  cursor: pointer;
  background: rgba(248, 249, 255, 0.8);
}

.upload-drop:hover {
  border-color: rgba(112, 100, 255, 0.6);
}

.upload-drop input {
  display: none;
}

.drop-icon {
  width: 56px;
  height: 56px;
  border-radius: 18px;
  display: grid;
  place-items: center;
  background: rgba(112, 100, 255, 0.1);
  color: var(--primary-dark);
  font-size: 28px;
  font-weight: 600;
}

.upload-drop strong {
  display: block;
  font-size: 18px;
}

.upload-drop p {
  margin: 6px 0 0;
  color: var(--text-muted);
  font-size: 13px;
}

.file-name {
  font-size: 14px;
  color: var(--primary-dark);
  font-weight: 600;
}

.upload-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  font-size: 13px;
  color: var(--text-muted);
  flex-wrap: wrap;
}

.upload-actions button {
  border: none;
  border-radius: 18px;
  padding: 12px 28px;
  background: linear-gradient(120deg, var(--primary), var(--secondary));
  color: #fff;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 15px 28px rgba(112, 100, 255, 0.3);
}

.upload-actions button:disabled {
  background: #cbd5f5;
  box-shadow: none;
  cursor: not-allowed;
}

@media (max-width: 960px) {
  .image-recognition-container {
    grid-template-columns: 1fr;
  }

  .vision-sidebar {
    padding: 36px 28px;
  }
}

@media (max-width: 640px) {
  .upload-drop {
    grid-template-columns: 1fr;
    text-align: center;
  }

  .upload-actions {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
