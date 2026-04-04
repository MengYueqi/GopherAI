<template>
  <div class="image-recognition-container">
    <aside class="vision-sidebar chinese-panel">
      <span class="chinese-tag">知形阁</span>
      <h1>知形 · 识万物</h1>
      <p>
        慧眼识真，万物皆可辨。拖拽或点击上传即可识别，历史记录皆可追溯。
      </p>
      <ul class="vision-highlights">
        <li>
          <strong>万式兼容</strong>
          <span>支持PNG、JPG、WebP、HEIC等主流格式。</span>
        </li>
        <li>
          <strong>传输无忧</strong>
          <span>仅当前会话可访问，安全无虞。</span>
        </li>
        <li>
          <strong>立等可得</strong>
          <span>识别成功即刻返回结果。</span>
        </li>
      </ul>
    </aside>

    <section class="vision-console chinese-panel">
      <div class="top-bar">
        <div>
          <h2>识鉴录</h2>
          <p>历次识别皆记录于此，便于比对查阅。</p>
        </div>
        <button class="back-btn" @click="$router.push('/menu')">返</button>
      </div>

      <div class="chat-messages" ref="chatContainerRef">
        <div
          v-for="(message, index) in messages"
          :key="index"
          :class="['message', message.role === 'user' ? 'user-message' : 'ai-message']"
        >
          <div class="message-header">
            <b>{{ message.role === 'user' ? '君' : '智' }}:</b>
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
              <strong>拖拽或点击上图</strong>
              <p>最大10MB，图片仅用于本次识别</p>
            </div>
            <span class="file-name" v-if="selectedFile">{{ selectedFile.name }}</span>
          </label>
          <div class="upload-actions">
            <span>支持批量识别，逐张上传以保精准。</span>
            <button type="submit" :disabled="!selectedFile">鉴图</button>
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
  background: var(--bg-color);
}

.vision-sidebar {
  padding: 48px;
  color: #fff;
  background: linear-gradient(160deg, var(--pine), var(--jade));
  display: flex;
  flex-direction: column;
  gap: 24px;
  position: relative;
  overflow: hidden;
}

.vision-sidebar::before {
  content: '';
  position: absolute;
  right: -20%;
  top: -20%;
  width: 300px;
  height: 300px;
  background: url("data:image/svg+xml,%3Csvg width='300' height='300' viewBox='0 0 300 300' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M150 30C80 30 30 80 30 150s50 120 120 120 120-50 120-120S220 30 150 30zm0 200c-44.2 0-80-35.8-80-80s35.8-80 80-80 80 35.8 80 80-35.8 80-80 80z' fill='%23fff' fill-opacity='0.05'/%3E%3Cpath d='M150 70c-44.2 0-80 35.8-80 80s35.8 80 80 80 80-35.8 80-80-35.8-80-80-80zm0 140c-33.1 0-60-26.9-60-60s26.9-60 60-60 60 26.9 60 60-26.9 60-60 60z' fill='%23fff' fill-opacity='0.05'/%3E%3C/svg%3E");
  opacity: 0.3;
}

.vision-sidebar h1 {
  margin: 0;
  font-size: 36px;
  font-family: 'Noto Serif SC', serif;
  font-weight: 600;
  letter-spacing: 0.05em;
}

.vision-sidebar p {
  color: rgba(255, 255, 255, 0.85);
  line-height: 1.8;
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
  border-left: 3px solid var(--gold);
  border-radius: var(--radius-md);
  padding: 20px;
  transition: all 0.3s ease;
}

.vision-highlights li:hover {
  background: rgba(255, 255, 255, 0.12);
  transform: translateX(4px);
}

.vision-highlights strong {
  display: block;
  margin-bottom: 6px;
  font-family: 'Noto Serif SC', serif;
  font-weight: 500;
  font-size: 16px;
}

.vision-highlights span {
  color: rgba(255, 255, 255, 0.75);
  font-size: 14px;
  line-height: 1.6;
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
