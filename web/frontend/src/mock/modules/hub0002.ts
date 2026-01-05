/**
 * Hub0002模块API的Mock数据
 * 用户管理相关接口
 */
import type { MockMethod } from 'vite-plugin-mock'
import type { JsonDataObj, PageInfoObj } from '@/types/api'
import Mock from 'mockjs'
import type { User } from '@/views/hub0002/types'

/**
 * 请求处理函数参数接口
 */
interface RequestParams {
  url: string
  body: Record<string, any>
  query: Record<string, string>
  headers: Record<string, string>
  method: string
  params: Record<string, string>
}

/**
 * 日期格式化函数
 * @param date 日期对象
 * @returns 格式化后的日期字符串 yyyy-MM-dd HH:mm:ss
 */
function formatDate(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')

  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

// 创建JsonDataObj格式的响应数据
function createJsonDataResponse<T>(
  data: T,
  pageInfo?: PageInfoObj,
  success = true,
  message = '',
): JsonDataObj {
  if (success) {
    return {
      oK: true,
      state: true,
      bizData: data ? JSON.stringify(data) : '',
      extObj: null,
      pageQueryData: pageInfo ? JSON.stringify(pageInfo) : '',
      messageId: '',
      errMsg: '',
      popMsg: message || '操作成功',
      extMsg: '',
      pkey1: '',
      pkey2: '',
      pkey3: '',
      pkey4: '',
      pkey5: '',
      pkey6: '',
    }
  } else {
    return {
      oK: false,
      state: false,
      bizData: '',
      extObj: null,
      pageQueryData: '',
      messageId: '',
      errMsg: message || '操作失败',
      popMsg: message || '操作失败',
      extMsg: '',
      pkey1: '',
      pkey2: '',
      pkey3: '',
      pkey4: '',
      pkey5: '',
      pkey6: '',
    }
  }
}

// 当前日期，用于生成一致的日期字符串
const now = new Date()
const expireDate = new Date('2099-12-31T23:59:59Z')

// 模拟用户数据
const userList: User[] = [
  {
    userId: '1',
    tenantId: 'default',
    userName: 'admin',
    realName: '管理员',
    deptId: '1',
    email: 'admin@example.com',
    mobile: '13800138000',
    avatar: 'https://avatars.githubusercontent.com/u/10000000',
    gender: 1,
    statusFlag: 'Y',
    deptAdminFlag: 'Y',
    tenantAdminFlag: 'Y',
    userExpireDate: formatDate(new Date('2099-12-31T23:59:59Z')),
    lastLoginTime: formatDate(new Date('2023-01-01T00:00:00Z')),
    lastLoginIp: '127.0.0.1',
    addTime: formatDate(new Date('2023-01-01T00:00:00Z')),
    addWho: 'system',
    editTime: formatDate(new Date('2023-01-01T00:00:00Z')),
    editWho: 'system',
    oprSeqFlag: 'A',
    currentVersion: 1,
    activeFlag: 'Y',
    noteText: '系统管理员账号',
  },
  {
    userId: '2',
    tenantId: 'default',
    userName: 'user1',
    realName: '张三',
    deptId: '2',
    email: 'zhangsan@example.com',
    mobile: '13900139001',
    gender: 1,
    statusFlag: 'Y',
    deptAdminFlag: 'Y',
    tenantAdminFlag: 'N',
    userExpireDate: formatDate(new Date('2099-12-31T23:59:59Z')),
    lastLoginTime: formatDate(new Date('2023-01-02T00:00:00Z')),
    lastLoginIp: '192.168.1.1',
    addTime: formatDate(new Date('2023-01-02T00:00:00Z')),
    addWho: 'admin',
    editTime: formatDate(new Date('2023-01-02T00:00:00Z')),
    editWho: 'admin',
    oprSeqFlag: 'A',
    currentVersion: 1,
    activeFlag: 'Y',
    noteText: '部门管理员',
  },
  {
    userId: '3',
    tenantId: 'default',
    userName: 'user2',
    realName: '李四',
    deptId: '2',
    email: 'lisi@example.com',
    mobile: '13900139002',
    gender: 1,
    statusFlag: 'Y',
    deptAdminFlag: 'N',
    tenantAdminFlag: 'N',
    userExpireDate: formatDate(new Date('2099-12-31T23:59:59Z')),
    lastLoginTime: formatDate(new Date('2023-01-03T00:00:00Z')),
    lastLoginIp: '192.168.1.2',
    addTime: formatDate(new Date('2023-01-03T00:00:00Z')),
    addWho: 'admin',
    editTime: formatDate(new Date('2023-01-03T00:00:00Z')),
    editWho: 'admin',
    oprSeqFlag: 'A',
    currentVersion: 1,
    activeFlag: 'Y',
  },
  {
    userId: '4',
    tenantId: 'default',
    userName: 'user3',
    realName: '王五',
    deptId: '3',
    email: 'wangwu@example.com',
    mobile: '13900139003',
    gender: 1,
    statusFlag: 'Y',
    deptAdminFlag: 'N',
    tenantAdminFlag: 'N',
    userExpireDate: formatDate(new Date('2099-12-31T23:59:59Z')),
    lastLoginTime: formatDate(new Date('2023-01-04T00:00:00Z')),
    lastLoginIp: '192.168.1.3',
    addTime: formatDate(new Date('2023-01-04T00:00:00Z')),
    addWho: 'admin',
    editTime: formatDate(new Date('2023-01-04T00:00:00Z')),
    editWho: 'admin',
    oprSeqFlag: 'A',
    currentVersion: 1,
    activeFlag: 'Y',
  },
  {
    userId: '5',
    tenantId: 'default',
    userName: 'user4',
    realName: '赵六',
    deptId: '3',
    email: 'zhaoliu@example.com',
    mobile: '13900139004',
    gender: 2,
    statusFlag: 'N',
    deptAdminFlag: 'N',
    tenantAdminFlag: 'N',
    userExpireDate: formatDate(new Date('2099-12-31T23:59:59Z')),
    lastLoginTime: formatDate(new Date('2023-01-05T00:00:00Z')),
    lastLoginIp: '192.168.1.4',
    addTime: formatDate(new Date('2023-01-05T00:00:00Z')),
    addWho: 'admin',
    editTime: formatDate(new Date('2023-01-05T00:00:00Z')),
    editWho: 'admin',
    oprSeqFlag: 'A',
    currentVersion: 1,
    activeFlag: 'Y',
  },
  {
    userId: '6',
    tenantId: 'tenant2',
    userName: 'tenant2_admin',
    realName: '租户2管理员',
    deptId: '1',
    email: 'tenant2@example.com',
    mobile: '13900139005',
    gender: 1,
    statusFlag: 'Y',
    deptAdminFlag: 'Y',
    tenantAdminFlag: 'Y',
    userExpireDate: formatDate(new Date('2099-12-31T23:59:59Z')),
    lastLoginTime: formatDate(new Date('2023-01-06T00:00:00Z')),
    lastLoginIp: '192.168.1.5',
    addTime: formatDate(new Date('2023-01-06T00:00:00Z')),
    addWho: 'system',
    editTime: formatDate(new Date('2023-01-06T00:00:00Z')),
    editWho: 'system',
    oprSeqFlag: 'A',
    currentVersion: 1,
    activeFlag: 'Y',
    noteText: '租户2的管理员账号',
  },
]

// 部门数据
const deptList = [
  { id: '1', label: '系统管理部', value: '1', parentId: '0' },
  { id: '2', label: '研发部', value: '2', parentId: '0' },
  { id: '3', label: '市场部', value: '3', parentId: '0' },
  { id: '4', label: '测试组', value: '4', parentId: '2' },
  { id: '5', label: '前端组', value: '5', parentId: '2' },
  { id: '6', label: '后端组', value: '6', parentId: '2' },
]

export default [
  // 获取用户列表
  {
    url: '/gateway/hub0002/queryUsers',
    method: 'post',
    response: ({ query }: Pick<RequestParams, 'query'>) => {
      const { userName, realName, deptId, mobile, statusFlag, pageIndex = 1, pageSize = 10 } = query

      // 筛选数据
      let filteredList = [...userList]

      if (userName) {
        filteredList = filteredList.filter((user) =>
          user.userName.toLowerCase().includes(userName.toLowerCase()),
        )
      }

      if (realName) {
        filteredList = filteredList.filter((user) => user.realName.includes(realName))
      }

      if (deptId) {
        filteredList = filteredList.filter((user) => user.deptId === deptId)
      }

      if (mobile) {
        filteredList = filteredList.filter((user) => user.mobile && user.mobile.includes(mobile))
      }

      if (statusFlag) {
        filteredList = filteredList.filter((user) => user.statusFlag === statusFlag)
      }

      // 添加部门名称等衍生字段
      const enhancedList = filteredList.map((user) => {
        const dept = deptList.find((d) => d.value === user.deptId)
        return {
          ...user,
          deptName: dept ? dept.label : '未知部门',
          // genderText: user.gender === 1 ? '男' : user.gender === 2 ? '女' : '保密',
          // statusFlagText: user.statusFlag === 'Y' ? '启用' : '禁用',
          // deptAdminFlagText: user.deptAdminFlag === 'Y' ? '是' : '否',
          // tenantAdminFlagText: user.tenantAdminFlag === 'Y' ? '是' : '否',
        }
      })

      // 分页处理
      const start = (Number(pageIndex) - 1) * Number(pageSize)
      const end = start + Number(pageSize)
      const pagedList = enhancedList.slice(start, end)

      // 构造分页信息，符合PageInfoObj结构
      const pageInfo: PageInfoObj = {
        baseData: '',
        curPageCount: pagedList.length,
        dbsId: 'default',
        mainKey: 'userId,tenantId',
        orderByList: '',
        otherData: '',
        pageIndex: Number(pageIndex),
        pageSize: Number(pageSize),
        paramObjectsJson: JSON.stringify(query),
        timeTypeFieldNames: 'addTime,editTime,lastLoginTime',
        totalCount: enhancedList.length,
        totalPageIndex: Math.ceil(enhancedList.length / Number(pageSize)),
      }

      // 返回用户列表数据，分页信息放在pageQueryData字段
      return createJsonDataResponse(pagedList, pageInfo, true, '获取用户列表成功')
    },
  },
  // 获取部门树
  {
    url: '/gateway/hub0002/queryDeptsTree',
    method: 'POST',
    response: () => {
      // 构建部门树
      interface TreeOptionBase {
        key: string
        label: string
        children?: TreeOptionBase[]
        disabled?: boolean
        [key: string]: any
      }

      const buildDeptTree = (parentId = '0'): TreeOptionBase[] => {
        return deptList
          .filter((dept) => dept.parentId === parentId)
          .map((dept) => ({
            key: dept.value,
            label: dept.label,
            value: dept.value,
            id: dept.id,
            parentId: dept.parentId,
            children: buildDeptTree(dept.id),
          }))
      }

      const deptTree = buildDeptTree()

      return createJsonDataResponse(deptTree, undefined, true, '获取部门树成功')
    },
  },
  // 编辑用户
  {
    url: '/gateway/hub0002/editUser',
    method: 'POST',
    response: ({ body }: Pick<RequestParams, 'body'>) => {
      const { userId, tenantId, ...updateData } = body as User

      // 在用户列表中查找对应的用户
      const userIndex = userList.findIndex(
        (user) => user.userId === userId && user.tenantId === tenantId,
      )

      if (userIndex === -1) {
        return createJsonDataResponse(null, undefined, false, '未找到指定用户')
      }

      // 更新用户数据
      const updatedUser = {
        ...userList[userIndex],
        ...updateData,
        editTime: formatDate(new Date()), // 更新修改时间为当前时间的字符串格式
        editWho: 'current_user', // 假设当前用户为修改者
        currentVersion: userList[userIndex].currentVersion + 1, // 版本号+1
      }

      // 替换原有用户数据
      userList[userIndex] = updatedUser

      // 添加部门名称
      const dept = deptList.find((d) => d.value === updatedUser.deptId)
      const enhancedUser = {
        ...updatedUser,
        deptName: dept ? dept.label : '未知部门',
      }

      // 返回更新后的完整用户数据
      return createJsonDataResponse(enhancedUser, undefined, true, '用户更新成功')
    },
  },
  // 删除用户
  {
    url: '/gateway/hub0002/deleteUser',
    method: 'POST',
    response: ({ body }: Pick<RequestParams, 'body'>) => {
      const { userId, tenantId } = body

      // 参数校验
      if (!userId || !tenantId) {
        return createJsonDataResponse(null, undefined, false, '用户ID和租户ID不能为空')
      }

      // 在用户列表中查找对应的用户
      const userIndex = userList.findIndex(
        (user) => user.userId === userId && user.tenantId === tenantId,
      )

      if (userIndex === -1) {
        return createJsonDataResponse(null, undefined, false, '未找到指定用户')
      }

      // 获取将要删除的用户信息
      const deletedUser = userList[userIndex]

      // 从数组中删除用户
      userList.splice(userIndex, 1)

      // 返回删除成功消息
      return createJsonDataResponse(
        { userId, tenantId },
        undefined,
        true,
        `成功删除用户 ${deletedUser.realName || deletedUser.userName}`,
      )
    },
  },
  // 添加用户
  {
    url: '/gateway/hub0002/addUser',
    method: 'POST',
    response: ({ body }: Pick<RequestParams, 'body'>) => {
      const userData = body as User

      // 生成新的用户ID
      const newUserId = String(Math.max(...userList.map((u) => Number(u.userId))) + 1)

      // 创建新用户数据
      const newUser: User = {
        ...userData,
        userId: newUserId,
        addTime: formatDate(new Date()),
        editTime: formatDate(new Date()),
        addWho: 'current_user',
        editWho: 'current_user',
        currentVersion: 1,
        oprSeqFlag: 'A',
        lastLoginTime: '',
        lastLoginIp: '',
      }

      // 添加到用户列表
      userList.push(newUser)

      // 添加部门名称
      const dept = deptList.find((d) => d.value === newUser.deptId)
      const enhancedUser = {
        ...newUser,
        deptName: dept ? dept.label : '未知部门',
      }

      return createJsonDataResponse(enhancedUser, undefined, true, '用户添加成功')
    },
  },
] as MockMethod[]
