import { toast } from "sonner"

import { normalizeUnixTimestamp } from "@/features/chat/state"
import type { ChatAttachment, ChatMessage } from "@/store/chat"
import { updateChatStore } from "@/store/chat"

export interface PicoMessage {
  type: string
  id?: string
  session_id?: string
  timestamp?: number | string
  payload?: Record<string, unknown>
}

function normalizeMessageTimestamp(raw: number | string | undefined): number {
  if (raw === undefined) {
    return Date.now()
  }

  const numeric = Number(raw)
  if (!Number.isFinite(numeric)) {
    return Date.now()
  }

  return normalizeUnixTimestamp(numeric)
}

function readImageURL(item: unknown): string | null {
  if (typeof item === "string" && item.trim() !== "") {
    return item.trim()
  }

  if (item && typeof item === "object") {
    const data = item as Record<string, unknown>
    const candidate = data.data_url ?? data.url
    if (typeof candidate === "string" && candidate.trim() !== "") {
      return candidate.trim()
    }
  }

  return null
}

function toImageAttachments(payload: Record<string, unknown>): ChatAttachment[] | undefined {
  const attachments: ChatAttachment[] = []

  const media = payload.media
  if (Array.isArray(media)) {
    for (const item of media) {
      const url = readImageURL(item)
      if (!url) {
        continue
      }
      attachments.push({ type: "image", url })
    }
  } else {
    const url = readImageURL(media)
    if (url) {
      attachments.push({ type: "image", url })
    }
  }

  const singleMediaFields = [payload.image, payload.image_url, payload.url]
  for (const field of singleMediaFields) {
    const url = readImageURL(field)
    if (!url) {
      continue
    }

    const duplicated = attachments.some((attachment) => attachment.url === url)
    if (!duplicated) {
      attachments.push({ type: "image", url })
    }
  }

  return attachments.length > 0 ? attachments : undefined
}

function createAssistantMessage({
  id,
  content,
  timestamp,
  attachments,
}: {
  id: string
  content: string
  timestamp: number
  attachments?: ChatAttachment[]
}): ChatMessage {
  return {
    id,
    role: "assistant",
    content,
    timestamp,
    attachments,
  }
}

export function handlePicoMessage(
  message: PicoMessage,
  expectedSessionId: string,
) {
  if (message.session_id && message.session_id !== expectedSessionId) {
    return
  }

  const payload = message.payload || {}

  switch (message.type) {
    case "message.create": {
      const content = typeof payload.content === "string" ? payload.content : ""
      const messageId =
        (typeof payload.message_id === "string" && payload.message_id) ||
        `pico-${Date.now()}`
      const timestamp = normalizeMessageTimestamp(message.timestamp)
      const attachments = toImageAttachments(payload)

      updateChatStore((prev) => ({
        messages: [
          ...prev.messages,
          createAssistantMessage({
            id: messageId,
            content,
            timestamp,
            attachments,
          }),
        ],
        isTyping: false,
      }))
      break
    }

    case "message.update": {
      const content = typeof payload.content === "string" ? payload.content : ""
      const messageId =
        typeof payload.message_id === "string" ? payload.message_id : ""
      if (!messageId) {
        break
      }

      const timestamp = normalizeMessageTimestamp(message.timestamp)
      const attachments = toImageAttachments(payload)

      updateChatStore((prev) => {
        const index = prev.messages.findIndex((msg) => msg.id === messageId)
        if (index < 0) {
          return {
            messages: [
              ...prev.messages,
              createAssistantMessage({
                id: messageId,
                content,
                timestamp,
                attachments,
              }),
            ],
            isTyping: false,
          }
        }

        return {
          messages: prev.messages.map((msg) =>
            msg.id === messageId
              ? {
                  ...msg,
                  content,
                  attachments: attachments ?? msg.attachments,
                }
              : msg,
          ),
          isTyping: false,
        }
      })
      break
    }

    case "media.create": {
      const content = typeof payload.content === "string" ? payload.content : ""
      const messageId =
        (typeof payload.message_id === "string" && payload.message_id) ||
        `pico-media-${Date.now()}`
      const timestamp = normalizeMessageTimestamp(message.timestamp)
      const attachments = toImageAttachments(payload)

      if (!content.trim() && !attachments) {
        break
      }

      updateChatStore((prev) => ({
        messages: [
          ...prev.messages,
          createAssistantMessage({
            id: messageId,
            content,
            timestamp,
            attachments,
          }),
        ],
        isTyping: false,
      }))
      break
    }

    case "typing.start":
      updateChatStore({ isTyping: true })
      break

    case "typing.stop":
      updateChatStore({ isTyping: false })
      break

    case "error": {
      const requestId =
        typeof payload.request_id === "string" ? payload.request_id : ""
      const errorMessage =
        typeof payload.message === "string" ? payload.message : ""

      console.error("Pico error:", payload)
      if (errorMessage) {
        toast.error(errorMessage)
      }
      updateChatStore((prev) => ({
        messages: requestId
          ? prev.messages.filter((msg) => msg.id !== requestId)
          : prev.messages,
        isTyping: false,
      }))
      break
    }

    case "session.info":
    case "pong":
      break

    default:
      console.log("Unknown pico message type:", message.type)
  }
}
