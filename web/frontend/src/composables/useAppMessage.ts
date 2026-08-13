/**
 * useAppMessage 提供接近 Naive `useMessage` 的调用面，底层走 niuma-ui `useRsToast`。
 */
import { useRsToast, type RsToastInput, type RsToastPosition } from '@/ui'

type MessageContent =
  | string
  | number
  | {
      content?: string | number
      title?: string
      description?: string
      duration?: number
      position?: RsToastPosition | 'top' | 'bottom'
      type?: string
    }

function mapPosition(position?: string): RsToastPosition | undefined {
  if (!position) return undefined
  if (position === 'top') return 'top-center'
  if (position === 'bottom') return 'bottom-center'
  return position as RsToastPosition
}

function toToastInput(content: MessageContent): RsToastInput {
  if (typeof content === 'string' || typeof content === 'number') {
    return String(content)
  }
  const title = content.title ?? content.content
  if (title === undefined || title === null) return ''
  return {
    title: String(title),
    description: content.description,
    duration: content.duration,
    position: mapPosition(content.position),
  }
}

export function useAppMessage() {
  const toast = useRsToast()

  return {
    success: (content: MessageContent) => toast.success(toToastInput(content)),
    error: (content: MessageContent) => toast.error(toToastInput(content)),
    info: (content: MessageContent) => toast.info(toToastInput(content)),
    warning: (content: MessageContent) => toast.warning(toToastInput(content)),
    loading: (content: MessageContent) => toast.info(toToastInput(content)),
    destroyAll: () => toast.dismiss(),
  }
}
