/**
 * 图标工具类单元测试
 * 测试图标获取和缓存功能
 */

import { describe, expect, it } from 'vitest'
import {
    CommonIcons,
    getIcon,
    getIcons,
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

  describe('getIcons', () => {
    it('should get multiple icons', async () => {
      const icons = await getIcons([
        'AddOutline',
        'RefreshOutline',
        'TrashOutline'
      ])
      
      expect(icons).toHaveLength(3)
      expect(icons[0]).toBeTruthy()
      expect(icons[1]).toBeTruthy()
      expect(icons[2]).toBeTruthy()
    })

    it('should handle mix of valid and invalid icons', async () => {
      const icons = await getIcons([
        'AddOutline',
        'NonExistentIcon',
        'RefreshOutline'
      ])
      
      expect(icons).toHaveLength(3)
      expect(icons[0]).toBeTruthy()
      expect(icons[1]).toBeNull()
      expect(icons[2]).toBeTruthy()
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

