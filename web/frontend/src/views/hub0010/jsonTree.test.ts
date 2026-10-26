import { describe, expect, it } from 'vitest'
import { highlightJsonLine } from './jsonHighlight'
import { buildJsonTree, filterJsonTree, initialOpenIds } from './jsonTree'

describe('buildJsonTree', () => {
  it('按键顺序展开对象和数组', () => {
    const tree = buildJsonTree(JSON.parse('{"b":1,"a":[true,null,"hi"]}'))
    expect(tree.kind).toBe('object')
    expect(tree.children.map((child) => child.label)).toEqual(['b', 'a'])
    expect(tree.children[0]?.preview).toBe('1')
    expect(tree.children[1]?.preview).toBe('[3]')
    expect(tree.children[1]?.children.map((child) => child.preview)).toEqual(['true', 'null', '"hi"'])
    expect(tree.children[1]?.children[0]?.label).toBe('0')
  })

  it('按键或值留下命中路径', () => {
    const tree = buildJsonTree({ user: { name: '安' }, tags: ['a'] })
    const byKey = filterJsonTree(tree, 'name')
    expect(byKey?.children.map((child) => child.label)).toEqual(['user'])
    expect(byKey?.children[0]?.children.map((child) => child.label)).toEqual(['name'])
    const byValue = filterJsonTree(tree, '安')
    expect(byValue?.children[0]?.children[0]?.preview).toBe('"安"')
  })

  it('小树默认全部展开', () => {
    const tree = buildJsonTree({ a: { b: 1 } })
    expect(initialOpenIds(tree)).toEqual(['$', '$.a'])
  })
})

describe('highlightJsonLine', () => {
  it('区分键、字符串和数字', () => {
    expect(highlightJsonLine('  "a": "x",')).toEqual([
      { text: '  ', kind: 'plain' },
      { text: '"a"', kind: 'key' },
      { text: ':', kind: 'punct' },
      { text: ' ', kind: 'plain' },
      { text: '"x"', kind: 'string' },
      { text: ',', kind: 'punct' },
    ])
  })
})
