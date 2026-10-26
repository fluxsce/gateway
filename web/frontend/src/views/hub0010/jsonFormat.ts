/** JSON 语法错误种类。行列从 1 开始，index 是源文本中的字符偏移。 */
export type JsonFaultCode =
  | 'empty'
  | 'unexpected'
  | 'trailing'
  | 'trailing-comma'
  | 'unterminated-string'
  | 'bad-escape'
  | 'bad-unicode'
  | 'control-char'
  | 'bad-number'
  | 'expected-colon'
  | 'expected-comma-or-end'
  | 'expected-value'
  | 'eof'

export interface JsonFault {
  code: JsonFaultCode
  index: number
  length: number
  line: number
  column: number
  token: string
}

export type JsonFormatResult = { ok: true; text: string } | { ok: false; fault: JsonFault }

const WS = new Set([' ', '\t', '\n', '\r'])

class Reader {
  index = 0
  line = 1
  column = 1

  constructor(readonly source: string) {}

  peek(): string {
    return this.source[this.index] ?? ''
  }

  bump(): string {
    const char = this.source[this.index] ?? ''
    this.index += 1
    if (char === '\n') {
      this.line += 1
      this.column = 1
    } else if (char) {
      this.column += 1
    }
    return char
  }

  skipWs() {
    while (WS.has(this.peek())) this.bump()
  }
}

function faultAt(reader: Reader, code: JsonFaultCode, length = 1): JsonFault {
  const token = reader.source.slice(reader.index, reader.index + Math.max(length, 0))
  return {
    code,
    index: reader.index,
    length: token.length,
    line: reader.line,
    column: reader.column,
    token,
  }
}

function parseLiteral(reader: Reader, word: string) {
  for (let i = 0; i < word.length; i += 1) {
    if (reader.peek() !== word[i]) return faultAt(reader, 'unexpected')
    reader.bump()
  }
  return null
}

function parseString(reader: Reader): JsonFault | null {
  reader.bump()
  while (true) {
    const char = reader.peek()
    if (!char) return faultAt(reader, 'unterminated-string', 0)
    if (char === '"') {
      reader.bump()
      return null
    }
    if (char === '\n' || char === '\r') return faultAt(reader, 'unterminated-string')
    if (char < ' ') return faultAt(reader, 'control-char')
    if (char !== '\\') {
      reader.bump()
      continue
    }
    reader.bump()
    const escaped = reader.peek()
    if (!escaped) return faultAt(reader, 'bad-escape', 0)
    if ('"\\/bfnrt'.includes(escaped)) {
      reader.bump()
      continue
    }
    if (escaped !== 'u') return faultAt(reader, 'bad-escape')
    reader.bump()
    for (let i = 0; i < 4; i += 1) {
      if (!/[0-9a-fA-F]/.test(reader.peek())) return faultAt(reader, 'bad-unicode', reader.peek() ? 1 : 0)
      reader.bump()
    }
  }
}

function parseNumber(reader: Reader): JsonFault | null {
  if (reader.peek() === '-') {
    reader.bump()
    if (!isDigit(reader.peek())) return faultAt(reader, 'bad-number', reader.peek() ? 1 : 0)
  }
  if (!isDigit(reader.peek())) return faultAt(reader, 'bad-number', reader.peek() ? 1 : 0)
  if (reader.peek() === '0') {
    reader.bump()
    if (isDigit(reader.peek())) return faultAt(reader, 'bad-number')
  } else {
    while (isDigit(reader.peek())) reader.bump()
  }
  if (reader.peek() === '.') {
    reader.bump()
    if (!isDigit(reader.peek())) return faultAt(reader, 'bad-number', reader.peek() ? 1 : 0)
    while (isDigit(reader.peek())) reader.bump()
  }
  if (reader.peek() === 'e' || reader.peek() === 'E') {
    reader.bump()
    if (reader.peek() === '+' || reader.peek() === '-') reader.bump()
    if (!isDigit(reader.peek())) return faultAt(reader, 'bad-number', reader.peek() ? 1 : 0)
    while (isDigit(reader.peek())) reader.bump()
  }
  return null
}

function parseValue(reader: Reader): JsonFault | null {
  reader.skipWs()
  const char = reader.peek()
  if (!char) return faultAt(reader, 'eof', 0)
  if (char === '"') return parseString(reader)
  if (char === '{') return parseObject(reader)
  if (char === '[') return parseArray(reader)
  if (char === 't') return parseLiteral(reader, 'true')
  if (char === 'f') return parseLiteral(reader, 'false')
  if (char === 'n') return parseLiteral(reader, 'null')
  if (char === '-' || isDigit(char)) return parseNumber(reader)
  return faultAt(reader, 'unexpected')
}

function parseObject(reader: Reader): JsonFault | null {
  reader.bump()
  reader.skipWs()
  if (reader.peek() === '}') {
    reader.bump()
    return null
  }
  while (true) {
    reader.skipWs()
    if (reader.peek() !== '"') return faultAt(reader, reader.peek() ? 'expected-value' : 'eof', reader.peek() ? 1 : 0)
    const key = parseString(reader)
    if (key) return key
    reader.skipWs()
    if (reader.peek() !== ':') return faultAt(reader, 'expected-colon', reader.peek() ? 1 : 0)
    reader.bump()
    const value = parseValue(reader)
    if (value) return value
    reader.skipWs()
    if (reader.peek() === ',') {
      const comma = faultAt(reader, 'trailing-comma')
      reader.bump()
      reader.skipWs()
      if (reader.peek() === '}' || !reader.peek()) return comma
      continue
    }
    if (reader.peek() === '}') {
      reader.bump()
      return null
    }
    return faultAt(reader, 'expected-comma-or-end', reader.peek() ? 1 : 0)
  }
}

function parseArray(reader: Reader): JsonFault | null {
  reader.bump()
  reader.skipWs()
  if (reader.peek() === ']') {
    reader.bump()
    return null
  }
  while (true) {
    const value = parseValue(reader)
    if (value) return value
    reader.skipWs()
    if (reader.peek() === ',') {
      const comma = faultAt(reader, 'trailing-comma')
      reader.bump()
      reader.skipWs()
      if (reader.peek() === ']' || !reader.peek()) return comma
      continue
    }
    if (reader.peek() === ']') {
      reader.bump()
      return null
    }
    return faultAt(reader, 'expected-comma-or-end', reader.peek() ? 1 : 0)
  }
}

function isDigit(char: string) {
  return char >= '0' && char <= '9'
}

/**
 * 找出第一处不符合 JSON 的字符。合法文本返回 null。
 * 合法与否和 JSON.parse 对齐：空白只认空格、制表符、换行和回车。
 */
export function locateJsonFault(source: string): JsonFault | null {
  if (!source.trim()) {
    return { code: 'empty', index: 0, length: 0, line: 1, column: 1, token: '' }
  }
  const reader = new Reader(source)
  const fault = parseValue(reader)
  if (fault) return fault
  reader.skipWs()
  if (reader.peek()) return faultAt(reader, 'trailing')
  return null
}

/** 缩进格式化在末尾补一个换行；压缩保持单行。失败时给出可定位的语法错误。 */
export function formatJsonText(source: string, space?: number): JsonFormatResult {
  const fault = locateJsonFault(source)
  if (fault) return { ok: false, fault }
  try {
    const value = JSON.parse(source) as unknown
    const text = space === undefined ? JSON.stringify(value) : `${JSON.stringify(value, null, space)}\n`
    return { ok: true, text }
  } catch {
    return {
      ok: false,
      fault: { code: 'unexpected', index: 0, length: source ? 1 : 0, line: 1, column: 1, token: source[0] ?? '' },
    }
  }
}

/** 把出错行和插入符排成等宽对照。制表符按两个空格对齐。 */
export function jsonFaultSnippet(source: string, fault: JsonFault): string {
  const raw = source.split('\n')[fault.line - 1] ?? ''
  let visual = ''
  let caretAt = 0
  for (let i = 0; i < raw.length; i += 1) {
    if (i === fault.column - 1) caretAt = visual.length
    visual += raw[i] === '\t' ? '  ' : raw[i]
  }
  if (fault.column - 1 >= raw.length) caretAt = visual.length
  const mark = fault.length > 1 ? '~'.repeat(Math.min(fault.length, 8)) : '^'
  const prefix = `${fault.line} | `
  return `${prefix}${visual}\n${' '.repeat(prefix.length + caretAt)}${mark}`
}
