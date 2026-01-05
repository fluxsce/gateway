/**
 * 本地存储工具类
 * 封装浏览器的localStorage和sessionStorage API
 * 提供类型安全、序列化/反序列化支持的存储方法
 */

/**
 * 存储数据到localStorage
 * localStorage在浏览器关闭后数据仍然保留
 *
 * @param key 键名 - 存储数据的唯一标识
 * @param value 值 - 要存储的数据，会自动将对象转为JSON字符串
 */
export const setLocalStorage = (key: string, value: any): void => {
  if (typeof value === 'object') {
    localStorage.setItem(key, JSON.stringify(value))
  } else {
    localStorage.setItem(key, value)
  }
}

/**
 * 从localStorage获取数据
 * 会自动尝试将存储的JSON字符串解析为对象
 *
 * @param key 键名 - 要获取数据的唯一标识
 * @param defaultValue 默认值 - 当数据不存在时返回的默认值
 * @returns 存储的值(自动解析JSON)或默认值
 */
export const getLocalStorage = <T = any>(key: string, defaultValue?: T): T => {
  const value = localStorage.getItem(key)
  if (value === null) return defaultValue as T

  try {
    return JSON.parse(value) as T
  } catch (error) {
    return value as unknown as T
  }
}

/**
 * 从localStorage移除数据
 *
 * @param key 键名 - 要移除数据的唯一标识
 */
export const removeLocalStorage = (key: string): void => {
  localStorage.removeItem(key)
}

/**
 * 清空localStorage中的所有数据
 * 慎用！会清除网站在当前域下所有的localStorage数据
 */
export const clearLocalStorage = (): void => {
  localStorage.clear()
}

/**
 * 存储数据到sessionStorage
 * sessionStorage在浏览器窗口/标签关闭后数据会被清除
 *
 * @param key 键名 - 存储数据的唯一标识
 * @param value 值 - 要存储的数据，会自动将对象转为JSON字符串
 */
export const setSessionStorage = (key: string, value: any): void => {
  if (typeof value === 'object') {
    sessionStorage.setItem(key, JSON.stringify(value))
  } else {
    sessionStorage.setItem(key, value)
  }
}

/**
 * 从sessionStorage获取数据
 * 会自动尝试将存储的JSON字符串解析为对象
 *
 * @param key 键名 - 要获取数据的唯一标识
 * @param defaultValue 默认值 - 当数据不存在时返回的默认值
 * @returns 存储的值(自动解析JSON)或默认值
 */
export const getSessionStorage = <T = any>(key: string, defaultValue?: T): T => {
  const value = sessionStorage.getItem(key)
  if (value === null) return defaultValue as T

  try {
    return JSON.parse(value) as T
  } catch (error) {
    return value as unknown as T
  }
}

/**
 * 从sessionStorage移除数据
 *
 * @param key 键名 - 要移除数据的唯一标识
 */
export const removeSessionStorage = (key: string): void => {
  sessionStorage.removeItem(key)
}

/**
 * 清空sessionStorage中的所有数据
 * 慎用！会清除网站在当前域下所有的sessionStorage数据
 */
export const clearSessionStorage = (): void => {
  sessionStorage.clear()
}
