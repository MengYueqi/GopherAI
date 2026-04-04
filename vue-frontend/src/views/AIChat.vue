<template>
  <div class="ai-chat-container">
    <!-- 左侧会话列表 -->
    <div class="session-list chinese-panel">
      <div class="session-list-header">
        <div>
          <h2>会话集</h2>
          <p>历史对话尽在掌握，往来问答历历可查。</p>
        </div>
        <button class="new-chat-btn" @click="createNewSession">＋ 新言</button>
      </div>
      <ul class="session-list-ul">
        <li
          v-for="session in sessions"
          :key="session.id"
          :class="['session-item', { active: currentSessionId === session.id }]"
          @click="switchSession(session.id)"
        >
          <div class="session-name">{{ session.name || `言 ${session.id}` }}</div>
          <div class="session-model" v-if="session.modelType">模型：{{ session.modelType }}</div>
          <div class="session-updated" v-if="session.updateAt">更新：{{ formatUpdateTime(session.updateAt) }}</div>
        </li>
      </ul>
    </div>

    <!-- 右侧聊天区域 -->
    <div class="chat-section chinese-panel">
      <div class="top-bar">
        <div class="top-left">
          <div class="top-heading">
            <h2>知言 · 对话阁</h2>
            <p>当前共有 {{ sessions.length }} 段对话，可开启新言或同步过往记录。</p>
          </div>
          <div class="top-actions">
            <button class="back-btn" @click="$router.push('/menu')">← 返</button>
            <button class="sync-btn" @click="syncHistory" :disabled="!currentSessionId || tempSession">同步过往</button>
          </div>
        </div>
        <div class="top-controls">
          <div class="select-group">
            <label for="modelType">择模</label>
            <select id="modelType" v-model="selectedModel" class="model-select">
              <option value="1">openai</option>
              <option value="2">ollama</option>
            </select>
          </div>
          <button
            type="button"
            class="chip-toggle"
            :class="{ active: isStreaming }"
            @click="isStreaming = !isStreaming"
          >
            <span class="chip-indicator google"></span>
            <span class="chip-text">
              <strong>流韵</strong>
              <small>应答如流</small>
            </span>
          </button>
          <!-- <button
            type="button"
            class="chip-toggle"
            :class="{ active: isUsingGoogle }"
            @click="toggleGoogle"
          >
            <span class="chip-indicator"></span>
            <span class="chip-text">
              <strong>使用 Google</strong>
              <small>{{ isUsingGoogle ? '已启用' : '未启用' }}</small>
            </span>
          </button> -->
          <!-- <button
            type="button"
            class="chip-toggle"
            :class="{ active: isUsingRAG }"
            @click="toggleRAG"
          >
            <span class="chip-indicator rag"></span>
            <span class="chip-text">
              <strong>行程增强</strong>
              <small>{{ isUsingRAG ? 'RAG 检索' : '默认模式' }}</small>
            </span>
          </button> -->
        </div>
      </div>

      <div class="chat-messages" ref="messagesRef">
        <div
          v-for="(message, index) in currentMessages"
          :key="index"
          :class="['message', message.role === 'user' ? 'user-message' : 'ai-message']"
        >
          <div class="message-header">
            <b>{{ message.role === 'user' ? '君' : '智' }}:</b>
            <button v-if="message.role === 'assistant'" class="tts-btn" @click="playTTS(message.content)">🔊</button>
            <span v-if="message.meta && message.meta.status === 'streaming'" class="streaming-indicator"> ··</span>
          </div>
          <div class="message-content">
            <MdPreview
              v-if="message.role === 'assistant'"
              :modelValue="message.content"
              previewTheme="github"
              :showCodeRowNumber="false"
            />
            <div v-else class="user-plain-text">{{ message.content }}</div>
          </div>
        </div>
      </div>

      <div class="chat-input">
        <div class="chat-input-hint" v-if="!canInteract">
          请点击“新言”或选择过往会话后输入
        </div>
        <textarea
          v-model="inputMessage"
          placeholder="请输入君之所问..."
          @keydown.enter.exact.prevent="sendMessage"
          :disabled="loading || !canInteract"
          ref="messageInput"
          rows="1"
        ></textarea>
        <button
          type="button"
          :disabled="!inputMessage.trim() || loading || !canInteract"
          @click="sendMessage"
          class="send-btn"
        >
          {{ loading ? '传书中...' : '传书' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, nextTick, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { MdPreview } from 'md-editor-v3'
import api from '../utils/api'

export default {
  name: 'AIChat',
  components: {
    MdPreview
  },
  setup() {

    const sessions = ref({})               
    const currentSessionId = ref(null)    
    const tempSession = ref(false)        
    const currentMessages = ref([])      
    const inputMessage = ref('')
    const loading = ref(false)
    const messagesRef = ref(null)
    const messageInput = ref(null)
    const selectedModel = ref('1')
    const isStreaming = ref(false)
    const isUsingGoogle = ref(false)
    const isUsingRAG = ref(false)

    const modelValueToLabel = (value) => {
      const normalized = String(value ?? '').toLowerCase()
      if (normalized === '1' || normalized === 'openai') return 'openai'
      if (normalized === '2' || normalized === 'ollama') return 'ollama'
      return normalized || ''
    }

    const modelLabelToValue = (label) => {
      const normalized = String(label ?? '').toLowerCase()
      if (!normalized) return selectedModel.value
      if (normalized === 'openai' || normalized === '1') return '1'
      if (normalized === 'ollama' || normalized === '2') return '2'
      return String(label)
    }

    const parseTimestamp = (value) => {
      if (!value) return 0
      const time = new Date(value).getTime()
      return Number.isNaN(time) ? 0 : time
    }

    const formatUpdateTime = (value) => {
      if (!value) return ''
      const date = new Date(value)
      if (Number.isNaN(date.getTime())) return value
      const pad = (num) => String(num).padStart(2, '0')
      return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
    }

    const touchSessionTimestamp = (sessionId, timestamp) => {
      if (!sessionId) return
      const sid = String(sessionId)
      if (!sessions.value[sid]) return
      sessions.value[sid].updateAt = timestamp || new Date().toISOString()
    }


    const playTTS = async (text) => {
      try {
        const response = await api.post('/chat/tts', { text })
        if (response.data && response.data.status_code === 1000 && response.data.url) {
          const audio = new Audio(response.data.url)
          audio.play()
        } else {
          ElMessage.error('无法获取语音')
        }
      } catch (error) {
        console.error('TTS error:', error)
        ElMessage.error('请求语音接口失败')
      }
    }

    const toggleGoogle = () => {
      if (!isUsingGoogle.value && isUsingRAG.value) {
        ElMessage.warning('Google 搜索和行程增强不能同时启用，请先关闭行程增强')
        return
      }
      isUsingGoogle.value = !isUsingGoogle.value
    }

    const toggleRAG = () => {
      if (!isUsingRAG.value && isUsingGoogle.value) {
        ElMessage.warning('Google 搜索和行程增强不能同时启用，请先关闭 Google 搜索')
        return
      }
      isUsingRAG.value = !isUsingRAG.value
    }

    const loadSessions = async () => {
      try {
        const response = await api.get('/AI/chat/sessions')
        if (response.data && response.data.status_code === 1000 && Array.isArray(response.data.sessions)) {
          const sessionMap = {}
          response.data.sessions.forEach(s => {
            const sid = String(s.sessionId)
            sessionMap[sid] = {
              id: sid,
              name: s.name || `会话 ${sid}`,
              modelType: s.modelType || '',
              updateAt: s.updateAt || s.updatedAt || '',
              messages: [] // lazy load
            }
          })
          sessions.value = sessionMap
        }
      } catch (error) {
        console.error('Load sessions error:', error)
      }
    }

    const createNewSession = () => {
      currentSessionId.value = 'temp'
      tempSession.value = true
      currentMessages.value = []
      // focus input
      nextTick(() => {
        if (messageInput.value) messageInput.value.focus()
      })
    }

    const switchSession = async (sessionId) => {
      if (!sessionId) return
      const normalizedId = String(sessionId)
      currentSessionId.value = normalizedId
      tempSession.value = false

      const sessionData = sessions.value[normalizedId]
      if (!sessionData) return
      if (sessionData.modelType) {
        selectedModel.value = modelLabelToValue(sessionData.modelType)
      }

      // lazy load history if not present
      if (!sessionData.messages || sessionData.messages.length === 0) {
        try {
          const response = await api.post('/AI/chat/history', { sessionId: currentSessionId.value })
          if (response.data && response.data.status_code === 1000 && Array.isArray(response.data.history)) {
            const messages = response.data.history.map(item => ({
              role: item.is_user ? 'user' : 'assistant',
              content: item.content
            }))
            sessions.value[normalizedId].messages = messages
            sessions.value[normalizedId].updateAt = response.data.updateAt || new Date().toISOString()
          }
        } catch (err) {
          console.error('Load history error:', err)
        }
      }


      currentMessages.value = [...(sessions.value[normalizedId].messages || [])]
      await nextTick()
      scrollToBottom()
    }

    const syncHistory = async () => {
      if (!currentSessionId.value || tempSession.value) {
        ElMessage.warning('请选择已有会话进行同步')
        return
      }
      try {
        const response = await api.post('/AI/chat/history', { sessionId: currentSessionId.value })
        if (response.data && response.data.status_code === 1000 && Array.isArray(response.data.history)) {
          const messages = response.data.history.map(item => ({
            role: item.is_user ? 'user' : 'assistant',
            content: item.content
          }))
          sessions.value[currentSessionId.value].messages = messages
          sessions.value[currentSessionId.value].updateAt = response.data.updateAt || new Date().toISOString()
          currentMessages.value = [...messages]
          await nextTick()
          scrollToBottom()
        } else {
          ElMessage.error('无法获取历史数据')
        }
      } catch (err) {
        console.error('Sync history error:', err)
        ElMessage.error('请求历史数据失败')
      }
    }


    const sendMessage = async () => {
      if (!tempSession.value && !currentSessionId.value) {
        ElMessage.warning('请先新建或选择会话')
        return
      }
      if (!inputMessage.value || !inputMessage.value.trim()) {
        ElMessage.warning('请输入消息内容')
        return
      }

      const userMessage = {
        role: 'user',
        content: inputMessage.value
      }
      const currentInput = inputMessage.value
      inputMessage.value = ''


      currentMessages.value.push(userMessage)
      await nextTick()
      scrollToBottom()

      try {
        loading.value = true
        if (isStreaming.value) {

          await handleStreaming(currentInput)
        } else {

          await handleNormal(currentInput)
        }
      } catch (err) {
        console.error('Send message error:', err)
        ElMessage.error('发送失败，请重试')

        if (!tempSession.value && currentSessionId.value && sessions.value[currentSessionId.value] && sessions.value[currentSessionId.value].messages) {

          const sessionArr = sessions.value[currentSessionId.value].messages
          if (sessionArr && sessionArr.length) sessionArr.pop()
        }
        currentMessages.value.pop()
      } finally {
        if (!isStreaming.value) {
          loading.value = false
        }
        await nextTick()
        scrollToBottom()
      }
    }


    async function handleStreaming(question) {

      const aiMessage = {
        role: 'assistant',
        content: '',
        meta: { status: 'streaming' } // mark streaming
      }


      const aiMessageIndex = currentMessages.value.length
      currentMessages.value.push(aiMessage)

      if (!tempSession.value && currentSessionId.value && sessions.value[currentSessionId.value]) {
        if (!sessions.value[currentSessionId.value].messages) sessions.value[currentSessionId.value].messages = []
        sessions.value[currentSessionId.value].messages.push({ role: 'assistant', content: '' })
      }


      const url = tempSession.value
        ? '/api/AI/chat/send-stream-new-session'  
        : '/api/AI/chat/send-stream'           

      const headers = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
      }

      const body = tempSession.value
        ? {
            question: question,
            modelType: selectedModel.value,
            usingGoogle: isUsingGoogle.value,
            usingRAG: isUsingRAG.value
          }
        : {
            question: question,
            modelType: selectedModel.value,
            sessionId: currentSessionId.value,
            usingGoogle: isUsingGoogle.value,
            usingRAG: isUsingRAG.value
          }

      try {
        // 创建 fetch 连接读取 SSE 流
        const response = await fetch(url, {
          method: 'POST',
          headers,
          body: JSON.stringify(body)
        })

        if (!response.ok) {
          loading.value = false
          throw new Error('Network response was not ok')
        }

        const reader = response.body.getReader()
        const decoder = new TextDecoder()
        let buffer = ''

        // 读取流数据
        // eslint-disable-next-line no-constant-condition
        while (true) {
          const { done, value } = await reader.read()
          if (done) break

          const chunk = decoder.decode(value, { stream: true })
          buffer += chunk

          // 按行分割
          const lines = buffer.split('\n')
          buffer = lines.pop() || '' // 保留未完成的行

          for (const line of lines) {
            const trimmedLine = line.trim()
            if (!trimmedLine) continue

            // 处理 SSE 格式：data: <content>
            if (trimmedLine.startsWith('data:')) {
              const data = trimmedLine.slice(5).trim()
              console.log('[SSE] Received:', data) // 调试日志

              if (data === '[DONE]') {
                // 流结束
                console.log('[SSE] Stream done')
                loading.value = false
                currentMessages.value[aiMessageIndex].meta = { status: 'done' }
                currentMessages.value = [...currentMessages.value]
              } else if (data.startsWith('{')) {
                // 尝试解析 JSON（如 sessionId）
                try {
                  const parsed = JSON.parse(data)
                  if (parsed.sessionId) {
                    const newSid = String(parsed.sessionId)
                    console.log('[SSE] Session ID:', newSid)
                    if (tempSession.value) {
                      sessions.value[newSid] = {
                        id: newSid,
                        name: '新会话',
                        modelType: parsed.modelType || modelValueToLabel(selectedModel.value),
                        updateAt: parsed.updateAt || new Date().toISOString(),
                        messages: [...currentMessages.value]
                      }
                      currentSessionId.value = newSid
                      tempSession.value = false
                    }
                  }
                } catch (e) {
                  // 不是 JSON，当作普通文本处理
                  currentMessages.value[aiMessageIndex].content += data
                  console.log('[SSE] Content updated:', currentMessages.value[aiMessageIndex].content.length)
                }
              } else {
                // 普通文本数据，直接追加
                // 使用数组索引直接更新，强制 Vue 响应式系统检测变化
                currentMessages.value[aiMessageIndex].content += data
                console.log('[SSE] Content updated:', currentMessages.value[aiMessageIndex].content.length)
              }

              // 每收到一条数据就立即更新 DOM
              // 强制更新整个数组以触发响应式
              currentMessages.value = [...currentMessages.value]
              
              // 使用 requestAnimationFrame 强制浏览器重排
              await new Promise(resolve => {
                requestAnimationFrame(() => {
                  scrollToBottom()
                  resolve()
                })
              })
            }
          }
        }

        // 流读取完成后的处理
        loading.value = false
        currentMessages.value[aiMessageIndex].meta = { status: 'done' }
        currentMessages.value = [...currentMessages.value]
        touchSessionTimestamp(currentSessionId.value)

        // 同步到 sessions 存储
        if (!tempSession.value && currentSessionId.value && sessions.value[currentSessionId.value]) {
          const sessMsgs = sessions.value[currentSessionId.value].messages
          if (Array.isArray(sessMsgs) && sessMsgs.length) {
            const lastIndex = sessMsgs.length - 1
            if (sessMsgs[lastIndex] && sessMsgs[lastIndex].role === 'assistant') {
              sessMsgs[lastIndex].content = currentMessages.value[aiMessageIndex].content
            }
          }
        }
      } catch (err) {
        console.error('Stream error:', err)
        loading.value = false
        currentMessages.value[aiMessageIndex].meta = { status: 'error' }
        currentMessages.value = [...currentMessages.value]
        ElMessage.error('流式传输出错')
      }
    }


    async function handleNormal(question) {
      if (tempSession.value) {

        const response = await api.post('/AI/chat/send-new-session', {
          question: question,
          modelType: selectedModel.value,
          usingGoogle: isUsingGoogle.value,
          usingRAG: isUsingRAG.value
        })
        if (response.data && response.data.status_code === 1000) {
          const sessionId = String(response.data.sessionId)
          const aiMessage = {
            role: 'assistant',
            content: response.data.Information || ''
          }

          sessions.value[sessionId] = {
            id: sessionId,
            name: '新会话',
            modelType: response.data.modelType || modelValueToLabel(selectedModel.value),
            updateAt: response.data.updateAt || new Date().toISOString(),
            messages: [ { role: 'user', content: question }, aiMessage ]
          }
          currentSessionId.value = sessionId
          tempSession.value = false
          currentMessages.value = [...sessions.value[sessionId].messages]
        } else {
          ElMessage.error(response.data?.status_msg || '发送失败')

          currentMessages.value.pop()
        }
      } else {

        const sessionMsgs = sessions.value[currentSessionId.value].messages

        sessionMsgs.push({ role: 'user', content: question })

        const response = await api.post('/AI/chat/send', {
          question: question,
          modelType: selectedModel.value,
          sessionId: currentSessionId.value,
          usingGoogle: isUsingGoogle.value,
          usingRAG: isUsingRAG.value
        })
        if (response.data && response.data.status_code === 1000) {
          const aiMessage = { role: 'assistant', content: response.data.Information || '' }
          sessionMsgs.push(aiMessage)
          currentMessages.value = [...sessionMsgs]
          touchSessionTimestamp(currentSessionId.value, response.data.updateAt || new Date().toISOString())
        } else {
          ElMessage.error(response.data?.status_msg || '发送失败')
          sessionMsgs.pop() // rollback
          currentMessages.value.pop()
        }
      }
    }


    const scrollToBottom = () => {
      if (messagesRef.value) {
        try {
          messagesRef.value.scrollTop = messagesRef.value.scrollHeight
        } catch (e) {
          // ignore
        }
      }
    }

    onMounted(() => {
      loadSessions()
    })

    const canInteract = computed(() => tempSession.value || !!currentSessionId.value)

    // expose to template
    return {
      sessions: computed(() => {
        const list = Object.values(sessions.value)
        return list.sort((a, b) => parseTimestamp(b.updateAt) - parseTimestamp(a.updateAt))
      }),
      currentSessionId,
      tempSession,
      currentMessages,
      inputMessage,
      loading,
      messagesRef,
      messageInput,
      selectedModel,
      isStreaming,
      isUsingGoogle,
      isUsingRAG,
      canInteract,
      formatUpdateTime,
      playTTS,
      createNewSession,
      switchSession,
      syncHistory,
      sendMessage,
      toggleGoogle,
      toggleRAG
    }
  }
}
</script>

<style scoped>
.ai-chat-container {
  min-height: 100vh;
  display: grid;
  grid-template-columns: minmax(260px, 320px) 1fr;
  gap: 24px;
  padding: 24px 4vw 32px;
  background: var(--bg-color);
  align-items: flex-start;
}

.session-list {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  height: calc(100vh - 56px);
  max-height: 760px;
  overflow-y: auto;
}

.session-list-header {
  padding: 28px 24px 18px;
  display: flex;
  justify-content: space-between;
  gap: 14px;
  border-bottom: 1px solid rgba(28, 28, 28, 0.08);
}

.session-list-header h2 {
  margin: 0 0 6px;
  font-size: 22px;
  font-family: 'Noto Serif SC', serif;
  color: var(--secondary);
  font-weight: 600;
}

.session-list-header p {
  margin: 0;
  font-size: 13px;
  color: var(--text-muted);
  line-height: 1.6;
}

.new-chat-btn {
  padding: 10px 16px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--primary);
  background: var(--primary);
  color: #fff;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  font-family: 'Noto Serif SC', serif;
}

.new-chat-btn:hover {
  transform: translateY(-2px);
  background: var(--primary-dark);
  box-shadow: var(--shadow-md);
}

.session-list-ul {
  list-style: none;
  padding: 0;
  margin: 0;
  flex: 1;
  overflow-y: auto;
}

.session-item {
  padding: 18px 24px;
  border-left: 3px solid transparent;
  cursor: pointer;
  transition: all 0.3s ease;
}

.session-item + .session-item {
  border-top: 1px solid rgba(28, 28, 28, 0.04);
}

.session-name {
  font-weight: 500;
  font-size: 15px;
  font-family: 'Noto Serif SC', serif;
}

.session-model,
.session-updated {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 4px;
}

.session-item:hover {
  background: rgba(157, 41, 51, 0.06);
  border-left-color: var(--primary);
}

.session-item.active {
  border-left-color: var(--primary);
  background: rgba(157, 41, 51, 0.1);
}

.chat-section {
  display: flex;
  flex-direction: column;
  min-width: 0;
  position: relative;
  overflow: hidden;
  height: calc(100vh - 56px);
  max-height: 760px;
}

.top-bar {
  position: relative;
  z-index: 1;
  padding: 20px 24px 16px;
  border-bottom: 1px solid rgba(28, 28, 28, 0.06);
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.top-left {
  display: flex;
  justify-content: space-between;
  gap: 18px;
  align-items: flex-end;
  flex-wrap: wrap;
}

.top-heading h2 {
  margin: 0;
  font-size: 24px;
  font-family: 'Noto Serif SC', serif;
  color: var(--secondary);
  font-weight: 600;
  letter-spacing: 0.05em;
}

.top-heading p {
  margin: 6px 0 0;
  color: var(--text-muted);
  font-size: 13px;
  line-height: 1.6;
}

.top-actions {
  display: flex;
  gap: 12px;
}

.top-controls {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  align-items: center;
}

.select-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 13px;
  color: var(--text-muted);
}

.select-group label {
  font-weight: 500;
  color: var(--secondary);
  font-family: 'Noto Serif SC', serif;
}

.model-select {
  border-radius: var(--radius-sm);
  border: 1px solid rgba(28, 28, 28, 0.15);
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.8);
  font-weight: 500;
  color: var(--text-color);
  transition: all 0.3s ease;
  font-family: inherit;
}

.model-select:focus {
  outline: none;
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(157, 41, 51, 0.08);
}

.chip-toggle {
  border: 1px solid rgba(28, 28, 28, 0.15);
  border-radius: var(--radius-sm);
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.6);
  color: var(--text-color);
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  min-width: 120px;
  transition: all 0.3s ease;
}

.chip-toggle .chip-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: rgba(28, 28, 28, 0.2);
}

.chip-toggle .chip-indicator.google {
  background: var(--jade);
}

.chip-toggle .chip-indicator.rag {
  background: var(--secondary);
}

.chip-toggle .chip-text {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
}

.chip-toggle strong {
  font-size: 12px;
  font-family: 'Noto Serif SC', serif;
  font-weight: 500;
}

.chip-toggle small {
  font-size: 10px;
  color: var(--text-muted);
}

.chip-toggle.active {
  border-color: var(--primary);
  background: rgba(157, 41, 51, 0.1);
  box-shadow: var(--shadow-sm);
}

.back-btn,
.sync-btn {
  padding: 8px 18px;
  border-radius: var(--radius-sm);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  font-family: 'Noto Serif SC', serif;
}

.back-btn {
  background: transparent;
  border: 1px solid var(--card-border);
  color: var(--text-color);
}

.back-btn:hover {
  border-color: var(--primary);
  color: var(--primary);
  background: rgba(157, 41, 51, 0.05);
  transform: translateY(-1px);
}

.sync-btn {
  background: var(--primary);
  border: 1px solid var(--primary);
  color: #fff;
  box-shadow: var(--shadow-sm);
}

.sync-btn:hover:not(:disabled) {
  background: var(--primary-dark);
  border-color: var(--primary-dark);
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.sync-btn:disabled {
  background: rgba(157, 41, 51, 0.3);
  border-color: rgba(157, 41, 51, 0.3);
  box-shadow: none;
  color: rgba(255, 255, 255, 0.7);
  cursor: not-allowed;
}

.chat-messages {
  position: relative;
  flex: 1;
  overflow-y: auto;
  padding: 24px 28px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  background: linear-gradient(180deg, rgba(245, 239, 226, 0.3), rgba(250, 246, 237, 0.3));
  background-image:
    url("data:image/svg+xml,%3Csvg width='100' height='100' viewBox='0 0 100 100' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M11 18c3.866 0 7-3.134 7-7s-3.134-7-7-7-7 3.134-7 7 3.134 7 7 7zm48 25c3.866 0 7-3.134 7-7s-3.134-7-7-7-7 3.134-7 7 3.134 7 7 7zm-43-7c1.657 0 3-1.343 3-3s-1.343-3-3-3-3 1.343-3 3 1.343 3 3 3zm63 31c1.657 0 3-1.343 3-3s-1.343-3-3-3-3 1.343-3 3 1.343 3 3 3zM34 90c1.657 0 3-1.343 3-3s-1.343-3-3-3-3 1.343-3 3 1.343 3 3 3zm56-76c1.657 0 3-1.343 3-3s-1.343-3-3-3-3 1.343-3 3 1.343 3 3 3zM12 86c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm28-65c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm23-11c2.76 0 5-2.24 5-5s-2.24-5-5-5-5 2.24-5 5 2.24 5 5 5zm-6 60c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm28 11c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4z' fill='%239D2933' fill-opacity='0.02' fill-rule='evenodd'/%3E%3C/svg%3E");
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
  border-radius: var(--radius-md);
  line-height: 1.8;
  font-size: 14px;
  box-shadow: var(--shadow-md);
  position: relative;
  transition: all 0.3s ease;
}

.message:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
}

.user-message {
  align-self: flex-end;
  background: linear-gradient(135deg, var(--cinnabar), var(--ochre));
  color: #fff;
  border-bottom-right-radius: 2px;
}

.ai-message {
  align-self: flex-start;
  background: #ffffff;
  color: var(--ink);
  border: 1px solid rgba(28, 28, 28, 0.06);
  border-bottom-left-radius: 2px;
}

.ai-message .message-content {
  color: var(--ink);
}

.ai-message :deep(.md-editor-preview),
.ai-message :deep(.md-editor-preview-wrapper),
.ai-message :deep(.markdown-body),
.ai-message :deep(.markdown-body p),
.ai-message :deep(.markdown-body span) {
  color: var(--ink);
}

.message-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-size: 12px;
  font-family: 'Noto Serif SC', serif;
  font-weight: 600;
  letter-spacing: 0.05em;
  color: rgba(28, 28, 28, 0.6);
}

.user-message .message-header {
  color: rgba(255, 255, 255, 0.8);
}

.message-content {
  white-space: normal;
  word-break: break-word;
}

.message-content :deep(.md-editor-preview) {
  background: transparent;
  padding: 0;
}

.message-content :deep(pre) {
  background: var(--ink);
  color: var(--ivory);
  border-radius: var(--radius-sm);
  padding: 12px;
  overflow-x: auto;
  border-left: 3px solid var(--gold);
}

.tts-btn {
  border: none;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.15);
  color: inherit;
  padding: 4px 10px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.tts-btn:hover {
  background: rgba(255, 255, 255, 0.25);
}

.streaming-indicator {
  color: var(--gold);
  font-weight: 600;
}

.chat-input {
  padding: 20px 24px 24px;
  display: flex;
  gap: 12px;
  align-items: flex-end;
  position: relative;
  border-top: 1px solid rgba(28, 28, 28, 0.06);
  background: var(--card-bg);
}

.chat-input textarea {
  flex: 1;
  border-radius: var(--radius-sm);
  border: 1px solid rgba(28, 28, 28, 0.15);
  padding: 14px;
  resize: none;
  min-height: 56px;
  max-height: 180px;
  background: rgba(255, 255, 255, 0.8);
  transition: all 0.3s ease;
  font-size: 14px;
  font-family: inherit;
}

.chat-input textarea:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(157, 41, 51, 0.08);
  outline: none;
}

.send-btn {
  border-radius: var(--radius-sm);
  border: 1px solid var(--primary);
  padding: 12px 28px;
  background: var(--primary);
  color: #fff;
  font-weight: 500;
  cursor: pointer;
  box-shadow: var(--shadow-sm);
  transition: all 0.3s ease;
  font-family: 'Noto Serif SC', serif;
  letter-spacing: 0.1em;
  position: relative;
  overflow: hidden;
}

.send-btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
  transition: left 0.5s ease;
}

.send-btn:hover::before {
  left: 100%;
}

.send-btn:disabled {
  background: rgba(157, 41, 51, 0.3);
  border-color: rgba(157, 41, 51, 0.3);
  box-shadow: none;
  cursor: not-allowed;
}

.send-btn:not(:disabled):hover {
  transform: translateY(-2px);
  background: var(--primary-dark);
  box-shadow: var(--shadow-md);
}

.chat-input-hint {
  position: absolute;
  top: 6px;
  left: 30px;
  color: #f97316;
  font-size: 12px;
}

@media (max-width: 1100px) {
  .ai-chat-container {
    grid-template-columns: 1fr;
    align-items: stretch;
  }

  .session-list {
    height: auto;
    max-height: none;
  }

  .chat-section {
    height: auto;
    max-height: none;
  }
}

@media (max-width: 720px) {
  .ai-chat-container {
    padding: 24px 16px 32px;
  }

  .chat-messages {
    padding: 20px 16px;
  }

  .top-left {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
