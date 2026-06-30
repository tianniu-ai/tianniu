import { fetchEventSource } from '@microsoft/fetch-event-source'

const BASE = '/api'

// Auth token management
let authToken: string | null = localStorage.getItem('auth_token')

export function setAuthToken(token: string) {
  authToken = token
  localStorage.setItem('auth_token', token)
}

export function clearAuthToken() {
  authToken = null
  localStorage.removeItem('auth_token')
}

export function getAuthToken() {
  return authToken
}

export function isLoggedIn() {
  return authToken !== null
}

// Get user info from token
let currentUser: UserVO | null = null

export function setCurrentUser(user: UserVO) {
  currentUser = user
  localStorage.setItem('current_user', JSON.stringify(user))
}

export function getCurrentUser() {
  if (!currentUser) {
    const stored = localStorage.getItem('current_user')
    if (stored) {
      try {
        currentUser = JSON.parse(stored) as UserVO
      } catch {
        // ignore
      }
    }
  }
  return currentUser
}

export function clearCurrentUser() {
  currentUser = null
  localStorage.removeItem('current_user')
}

export interface UserVO {
  id: string
  username: string
  email?: string
  created_at: number
}

export interface LoginRespVO {
  user: UserVO
  token: string
}

interface APIResponse<T> {
  code: number
  msg: string
  data?: T
}

export interface ConversationVO {
  id: string
  user_id: string
  title: string
  created_at: number
}

export interface ToolCallVO {
  id: string
  name: string
  arguments: string
}

export interface RoundMessageVO {
  role: 'user' | 'assistant' | 'tool'
  content?: string
  tool_calls?: ToolCallVO[]
  tool_name?: string
  tool_id?: string
}

export interface ChatMessageVO {
  id: string
  conversation_id: string
  parent_message_id: string
  query: string
  response: string
  model: string
  created_at: number
  rounds?: RoundMessageVO[]
}

export interface SSEMessageVO {
  id: string
  event: 'error' | 'reasoning' | 'content' | 'tool_call' | 'tool_result'
  content?: string
  reasoning_content?: string
  tool_call?: string
  tool_arguments?: string
  tool_result?: string
}

interface StreamThreadRunArgs {
  threadId: string
  query: string
  parentMessageId?: string
  signal?: AbortSignal
  onEvent: (event: SSEMessageVO) => void
  onClose: () => void
  onError?: (error: Error) => void
}

export type ThreadOperation = 'rename' | 'archive' | 'delete'

export interface ThreadOperationUnsupported {
	ok: false
	unsupported: true
	operation: ThreadOperation
  threadId: string
  message: string
}

export type ThreadOperationResult = ThreadOperationUnsupported

export const THREAD_OPERATION_SUPPORT: Record<ThreadOperation, boolean> = {
  rename: true,
  archive: false,
  delete: true,
}

export async function fetchThreads(): Promise<ConversationVO[]> {
  const json = await requestJSON<ConversationVO[]>(`${BASE}/conversation`)
  return json.data ?? []
}

export async function createThread(title = 'New Chat'): Promise<ConversationVO> {
  const json = await requestJSON<ConversationVO>(`${BASE}/conversation`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({title }),
  })
  if (!json.data) throw new Error('Conversation was not returned by the server')
  return json.data
}

export async function renameThread(threadId: string, title: string): Promise<ConversationVO> {
  const json = await requestJSON<ConversationVO>(`${BASE}/conversation/${threadId}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title }),
  })
  if (!json.data) throw new Error('Conversation rename did not return updated data')
  return json.data
}

export async function archiveThread(threadId: string): Promise<ThreadOperationResult> {
  return unsupportedThreadOperation('archive', threadId)
}

export async function deleteThread(threadId: string): Promise<void> {
  await requestJSON<{ conversation_id: string }>(`${BASE}/conversation/${threadId}`, {
    method: 'DELETE',
  })
}

export async function fetchThreadMessages(threadId: string): Promise<ChatMessageVO[]> {
  const json = await requestJSON<ChatMessageVO[]>(`${BASE}/conversation/${threadId}/message`)
  return json.data ?? []
}

export function streamThreadRun({
  threadId,
  query,
  parentMessageId,
  signal,
  onEvent,
  onClose,
  onError,
}: StreamThreadRunArgs): () => void {
  const ctrl = new AbortController()
  const cleanup = bindAbortSignal(signal, ctrl)
  let finalized = false

  const finalize = (callback?: () => void) => {
    if (finalized) return
    finalized = true
    cleanup()
    callback?.()
  }

  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (authToken) {
    headers['Authorization'] = `Bearer ${authToken}`
  }

  fetchEventSource(`${BASE}/conversation/${threadId}/message`, {
    method: 'POST',
    headers,
    body: JSON.stringify({query, parent_message_id: parentMessageId ?? '' }),
    signal: ctrl.signal,
    onmessage(ev) {
      const event = parseSSEMessage(ev.data)
      if (event) onEvent(event)
    },
    onclose() {
      finalize(onClose)
    },
    onerror(err) {
      throw err // stop retrying
    },
  }).catch((err) => {
    if (isAbortError(err)) {
      finalize()
      return
    }

    const error = normalizeStreamError(err)
    console.error('SSE error:', error)
    finalize(() => {
      onError?.(error)
      onClose()
    })
  })

  return () => {
    finalize()
    ctrl.abort()
  }
}

export const listConversations = fetchThreads
export const createConversation = createThread
export const listMessages = fetchThreadMessages

export function streamMessage(
  conversationId: string,
  query: string,
  onEvent: (event: SSEMessageVO) => void,
  onClose: () => void,
  parentMessageId?: string,
  onError?: (error: Error) => void,
): () => void {
  return streamThreadRun({
    threadId: conversationId,
    query,
    parentMessageId,
    onEvent,
    onClose,
    onError,
  })
}

async function requestJSON<T>(input: RequestInfo | URL, init?: RequestInit): Promise<APIResponse<T>> {
  const headers = new Headers(init?.headers)
  if (authToken) {
    headers.set('Authorization', `Bearer ${authToken}`)
  }

  const res = await fetch(input, {
    ...init,
    headers,
  })

  // Handle unauthorized
  if (res.status === 401) {
    clearAuthToken()
    clearCurrentUser()
    window.dispatchEvent(new Event('auth_required'))
    throw new Error('Unauthorized')
  }

  const json = await res.json() as APIResponse<T>
  if (json.code !== 0) throw new Error(json.msg)
  return json
}

// Auth APIs
export async function login(username: string, password: string): Promise<LoginRespVO> {
  const json = await fetch(`${BASE}/user/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  }).then(res => res.json()) as APIResponse<LoginRespVO>

  if (json.code !== 0) throw new Error(json.msg)
  if (!json.data) throw new Error('Login failed')

  return json.data
}

export async function register(username: string, email: string, password: string): Promise<UserVO> {
  const json = await fetch(`${BASE}/user/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, email, password }),
  }).then(res => res.json()) as APIResponse<UserVO>

  if (json.code !== 0) throw new Error(json.msg)
  if (!json.data) throw new Error('Registration failed')

  return json.data
}

function unsupportedThreadOperation(
  operation: ThreadOperation,
  threadId: string,
): ThreadOperationUnsupported {
  return {
    ok: false,
    unsupported: true,
    operation,
    threadId,
    message: `${operation}Thread is not implemented by the backend yet`,
  }
}

function parseSSEMessage(data: string): SSEMessageVO | null {
  try {
    return JSON.parse(data) as SSEMessageVO
  } catch {
    return null
  }
}

function bindAbortSignal(signal: AbortSignal | undefined, ctrl: AbortController): () => void {
  if (!signal) return () => {}
  if (signal.aborted) {
    ctrl.abort(signal.reason)
    return () => {}
  }

  const abort = () => ctrl.abort(signal.reason)
  signal.addEventListener('abort', abort, { once: true })
  return () => signal.removeEventListener('abort', abort)
}

function isAbortError(err: unknown): err is Error {
  return err instanceof Error && err.name === 'AbortError'
}

function normalizeStreamError(err: unknown): Error {
  if (err instanceof Error) return err
  if (typeof err === 'string') return new Error(err)

  try {
    return new Error(JSON.stringify(err))
  } catch {
    return new Error('Unknown SSE transport error')
  }
}