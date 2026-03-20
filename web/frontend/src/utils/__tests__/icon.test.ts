/**
 * 图标工具类单元测试
 * 测试图标获取和缓存功能
 */

import { describe, expect, it } from 'vitest'
import {
  CommonIcons,
  getIcon,
  getIconSync,
  IconLibrary
} from '../icon'

describe('Icon Utils', () => {
  describe('getIcon', () => {
    it('should get icon from Ionicons5 library', async () => {
      const icon = await getIcon('AddOutline')
      expect(icon).toBeTruthy()
    })

    it('should get icon from AntD library', async () => {
      const icon = await getIcon('UserOutlined', IconLibrary.ANTD)
      expect(icon).toBeTruthy()
    })

    it('should return null for non-existent icon', async () => {
      const icon = await getIcon('NonExistentIcon')
      expect(icon).toBeNull()
    })

    it('should return cached icon on second call', async () => {
      const icon1 = await getIcon('AddOutline')
      const icon2 = await getIcon('AddOutline')
      expect(icon1).toBe(icon2)
    })
  })

  describe('getIconSync', () => {
    it('should return null before icon is loaded', () => {
      expect(getIconSync('UnloadedIcon')).toBeNull()
    })

    it('should return cached icon after getIcon has loaded it', async () => {
      await getIcon('RefreshOutline')
      expect(getIconSync('RefreshOutline')).toBeTruthy()
    })
  })

  describe('CommonIcons', () => {
    it('should have common icon constants', () => {
      expect(CommonIcons.ADD).toBe('AddOutline')
      expect(CommonIcons.EDIT).toBe('CreateOutline')
      expect(CommonIcons.DELETE).toBe('TrashOutline')
      expect(CommonIcons.SAVE).toBe('SaveOutline')
      expect(CommonIcons.REFRESH).toBe('RefreshOutline')
    })

    it('should work with getIcon', async () => {
      const icon = await getIcon(CommonIcons.ADD)
      expect(icon).toBeTruthy()
    })
  })
})

