<template>
  <div class="ai-chat-container">
    <!-- 左侧会话列表 -->
    <div class="session-list">
      <div class="session-list-header">
        <div>
          <h2>会话总览</h2>
          <p>所有历史对话集中在此，随时切换与追溯。</p>
        </div>
        <button class="new-chat-btn" @click="createNewSession">＋ 新对话</button>
      </div>
      <ul class="session-list-ul">
        <li
          v-for="session in sessions"
          :key="session.id"
          :class="['session-item', { active: currentSessionId === session.id }]"
          @click="switchSession(session.id)"
        >
          <div class="session-name">{{ session.name || `会话 ${session.id}` }}</div>
          <div class="session-model" v-if="session.modelType">模型：{{ session.modelType }}</div>
          <div class="session-updated" v-if="session.updateAt">更新：{{ formatUpdateTime(session.updateAt) }}</div>
        </li>
      </ul>
    </div>

    <!-- 右侧聊天区域 -->
    <div class="chat-section">
      <div class="top-bar">
        <div class="top-left">
          <div class="top-heading">
            <h2>智能聊天控制台</h2>
            <p>当前共有 {{ sessions.length }} 个会话可用，开启新对话或同步历史。</p>
          </div>
          <div class="top-actions">
            <button class="back-btn" @click="$router.push('/menu')">← 返回</button>
            <button class="sync-btn" @click="syncHistory" :disabled="!currentSessionId || tempSession">同步历史数据</button>
          </div>
        </div>
        <div class="top-controls">
          <div class="select-group">
            <label for="modelType">选择模型</label>
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
              <strong>流式响应</strong>
              <small>实时输出</small>
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
            <b>{{ message.role === 'user' ? '你' : 'AI' }}:</b>
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
          请点击“新聊天”或选择历史会话后再输入
        </div>
        <textarea
          v-model="inputMessage"
          placeholder="请输入你的问题..."
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
          {{ loading ? '发送中...' : '发送' }}
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
  gap: 20px;
  padding: 24px 4vw 32px;
  background: radial-gradient(circle at 25% 20%, rgba(112, 100, 255, 0.18), transparent 55%),
    radial-gradient(circle at 80% 10%, rgba(45, 212, 191, 0.18), transparent 40%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.65), rgba(248, 249, 255, 0.8));
  align-items: flex-start;
}

.session-list {
  border-radius: 26px;
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(99, 102, 241, 0.12);
  box-shadow: 0 25px 60px rgba(15, 23, 42, 0.1);
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
  border-bottom: 1px solid rgba(15, 23, 42, 0.05);
}

.session-list-header h2 {
  margin: 0 0 6px;
  font-size: 22px;
}

.session-list-header p {
  margin: 0;
  font-size: 13px;
  color: var(--text-muted);
}

.new-chat-btn {
  padding: 12px 18px;
  border-radius: 18px;
  border: none;
  background: linear-gradient(130deg, var(--primary), var(--secondary));
  color: #fff;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 15px 30px rgba(112, 100, 255, 0.35);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.new-chat-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 22px 36px rgba(112, 100, 255, 0.45);
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
  border-left: 4px solid transparent;
  cursor: pointer;
  transition: 0.2s ease;
}

.session-item + .session-item {
  border-top: 1px solid rgba(15, 23, 42, 0.04);
}

.session-name {
  font-weight: 600;
  font-size: 15px;
}

.session-model,
.session-updated {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 4px;
}

.session-item:hover {
  background: rgba(112, 100, 255, 0.08);
}

.session-item.active {
  border-left-color: var(--primary);
  background: rgba(112, 100, 255, 0.12);
}

.chat-section {
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.97);
  box-shadow: 0 24px 60px rgba(15, 23, 42, 0.15);
  border: 1px solid rgba(99, 102, 241, 0.08);
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
  padding: 18px 24px 12px;
  border-bottom: 1px solid rgba(15, 23, 42, 0.06);
  display: flex;
  flex-direction: column;
  gap: 12px;
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
  font-size: 22px;
}

.top-heading p {
  margin: 4px 0 0;
  color: var(--text-muted);
  font-size: 13px;
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
  font-weight: 600;
  color: var(--text-color);
}

.model-select {
  border-radius: 16px;
  border: 1px solid rgba(15, 23, 42, 0.08);
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.9);
  font-weight: 600;
  color: var(--text-color);
}

.chip-toggle {
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 16px;
  padding: 8px 12px;
  background: rgba(248, 249, 255, 0.8);
  color: var(--text-color);
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  min-width: 140px;
  transition: 0.2s ease;
}

.chip-toggle .chip-indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: rgba(15, 23, 42, 0.2);
}

.chip-toggle .chip-indicator.google {
  background: linear-gradient(135deg, #34a853, #fbbc04, #4285f4);
}

.chip-toggle .chip-indicator.rag {
  background: linear-gradient(135deg, #c084fc, #9333ea);
}

.chip-toggle .chip-text {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
}

.chip-toggle strong {
  font-size: 12px;
}

.chip-toggle small {
  font-size: 10px;
  color: var(--text-muted);
}

.chip-toggle.active {
  border-color: transparent;
  background: rgba(112, 100, 255, 0.14);
  box-shadow: 0 15px 30px rgba(112, 100, 255, 0.25);
}

.back-btn,
.sync-btn {
  padding: 10px 20px;
  border-radius: 18px;
  border: none;
  font-weight: 600;
  cursor: pointer;
  transition: 0.2s ease;
}

.back-btn {
  background: rgba(112, 100, 255, 0.12);
  color: var(--primary-dark);
}

.sync-btn {
  background: linear-gradient(120deg, var(--primary), var(--secondary));
  color: #fff;
  box-shadow: 0 15px 28px rgba(112, 100, 255, 0.35);
}

.sync-btn:disabled {
  background: #cbd5f5;
  box-shadow: none;
  color: #64748b;
  cursor: not-allowed;
}

.chat-messages {
  position: relative;
  flex: 1;
  overflow-y: auto;
  padding: 22px 26px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: linear-gradient(180deg, rgba(248, 250, 255, 0.92), rgba(236, 241, 255, 0.74));
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
  padding: 16px 20px;
  border-radius: 16px;
  line-height: 1.6;
  font-size: 14px;
  box-shadow: 0 18px 30px rgba(15, 23, 42, 0.12);
  position: relative;
}

.user-message {
  align-self: flex-end;
  background: linear-gradient(135deg, #6d5bff, #3ec5ff);
  color: #fff;
}

.ai-message {
  align-self: flex-start;
  background: #ffffff;
  color: #000000;
  border: 1px solid rgba(15, 23, 42, 0.05);
}

.ai-message .message-content {
  color: #000000;
}

.ai-message :deep(.md-editor-preview),
.ai-message :deep(.md-editor-preview-wrapper),
.ai-message :deep(.markdown-body),
.ai-message :deep(.markdown-body p),
.ai-message :deep(.markdown-body span) {
  color: #000000;
}

.message-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: rgba(15, 23, 42, 0.6);
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
  background: #0f172a;
  color: #f8fafc;
  border-radius: 12px;
  padding: 12px;
  overflow-x: auto;
}

.tts-btn {
  border: none;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.15);
  color: inherit;
  padding: 4px 12px;
  cursor: pointer;
}

.streaming-indicator {
  color: #22d3ee;
  font-weight: 600;
}

.chat-input {
  padding: 16px 22px 22px;
  display: flex;
  gap: 12px;
  align-items: flex-end;
  position: relative;
  border-top: 1px solid rgba(15, 23, 42, 0.05);
  background: rgba(255, 255, 255, 0.92);
}

.chat-input textarea {
  flex: 1;
  border-radius: 18px;
  border: 1px solid rgba(15, 23, 42, 0.08);
  padding: 14px;
  resize: none;
  min-height: 56px;
  max-height: 180px;
  background: rgba(248, 249, 255, 0.92);
  box-shadow: inset 0 3px 8px rgba(15, 23, 42, 0.05);
  font-size: 14px;
}

.chat-input textarea:focus {
  border-color: rgba(112, 100, 255, 0.9);
  box-shadow: 0 0 0 3px rgba(112, 100, 255, 0.2);
  outline: none;
}

.send-btn {
  border-radius: 16px;
  border: none;
  padding: 12px 28px;
  background: linear-gradient(120deg, var(--primary), var(--secondary));
  color: #fff;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 14px 24px rgba(112, 100, 255, 0.3);
  transition: transform 0.2s ease;
}

.send-btn:disabled {
  background: #cbd5f5;
  box-shadow: none;
  cursor: not-allowed;
}

.send-btn:not(:disabled):hover {
  transform: translateY(-2px);
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
