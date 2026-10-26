import { describe, expect, it } from 'vitest'
import { formatJsonText, jsonFaultSnippet, locateJsonFault } from './jsonFormat'

const valid = [
  '0',
  '-0',
  '1.5',
  '1e2',
  '1E-2',
  'true',
  'false',
  'null',
  '""',
  '"中文"',
  '"\\n\\u0041"',
  '{}',
  '[]',
  '{"a":1,"b":[true,null]}',
  ' { "a" : [ 1 , 2 ] } ',
  '{"a":0.0}',
]

const invalid: Array<{ source: string; code: string; line: number; column: number }> = [
  { source: '', code: 'empty', line: 1, column: 1 },
  { source: '   ', code: 'empty', line: 1, column: 1 },
  { source: "{'a':1}", code: 'expected-value', line: 1, column: 2 },
  { source: '{"a":1,}', code: 'trailing-comma', line: 1, column: 7 },
  { source: '[1,]', code: 'trailing-comma', line: 1, column: 3 },
  { source: '{"a" 1}', code: 'expected-colon', line: 1, column: 6 },
  { source: '{"a":1 "b":2}', code: 'expected-comma-or-end', line: 1, column: 8 },
  { source: '[1 2]', code: 'expected-comma-or-end', line: 1, column: 4 },
  { source: '01', code: 'bad-number', line: 1, column: 2 },
  { source: '1.', code: 'bad-number', line: 1, column: 3 },
  { source: '"hi', code: 'unterminated-string', line: 1, column: 4 },
  { source: '"\\q"', code: 'bad-escape', line: 1, column: 3 },
  { source: '"\\u12"', code: 'bad-unicode', line: 1, column: 6 },
  { source: '{}{}', code: 'trailing', line: 1, column: 3 },
  { source: '{\n  "a": 1,\n}', code: 'trailing-comma', line: 2, column: 9 },
]

describe('locateJsonFault', () => {
  it('接受 JSON.parse 能接受的文本', () => {
    for (const source of valid) {
      expect(JSON.parse(source), source).toBeDefined()
      expect(locateJsonFault(source), source).toBeNull()
    }
  })

  it('拒绝 JSON.parse 会拒绝的文本，并标出行和列', () => {
    for (const sample of invalid) {
      expect(() => JSON.parse(sample.source), sample.source).toThrow()
      const fault = locateJsonFault(sample.source)
      expect(fault?.code, sample.source).toBe(sample.code)
      expect(fault?.line, sample.source).toBe(sample.line)
      expect(fault?.column, sample.source).toBe(sample.column)
    }
  })
})

describe('formatJsonText', () => {
  it('按缩进输出，并在末尾换行', () => {
    const result = formatJsonText('{"b":1,"a":[2]}', 2)
    expect(result).toEqual({
      ok: true,
      text: '{\n  "b": 1,\n  "a": [\n    2\n  ]\n}\n',
    })
  })

  it('压缩成一行', () => {
    const result = formatJsonText('{\n  "a": 1\n}')
    expect(result).toEqual({ ok: true, text: '{"a":1}' })
  })

  it('对照行用插入符标出列', () => {
    const source = '{\n  "a": 1,\n}'
    const fault = locateJsonFault(source)
    expect(fault).not.toBeNull()
    if (!fault) return
    const snippet = jsonFaultSnippet(source, fault)
    const [row, caret] = snippet.split('\n')
    expect(row?.startsWith('2 | ')).toBe(true)
    expect(caret?.indexOf('^')).toBe(row?.indexOf(','))
  })
})
