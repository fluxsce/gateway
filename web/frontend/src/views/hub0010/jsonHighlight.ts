export type JsonTokenKind = 'plain' | 'key' | 'string' | 'number' | 'literal' | 'punct'

export interface JsonToken {
  text: string
  kind: JsonTokenKind
}

/** 给已经格式化好的一行 JSON 上色。字符串跨行时只按本行可见部分处理。 */
export function highlightJsonLine(line: string): JsonToken[] {
  const tokens: JsonToken[] = []
  let index = 0
  const push = (kind: JsonTokenKind, text: string) => {
    if (text) tokens.push({ kind, text })
  }
  while (index < line.length) {
    const char = line[index] ?? ''
    if (char === '"') {
      let end = index + 1
      while (end < line.length) {
        if (line[end] === '\\') {
          end += 2
          continue
        }
        if (line[end] === '"') {
          end += 1
          break
        }
        end += 1
      }
      let look = end
      while (line[look] === ' ' || line[look] === '\t') look += 1
      push(line[look] === ':' ? 'key' : 'string', line.slice(index, end))
      index = end
      continue
    }
    if (char === ' ' || char === '\t') {
      let end = index + 1
      while (line[end] === ' ' || line[end] === '\t') end += 1
      push('plain', line.slice(index, end))
      index = end
      continue
    }
    if ('{}[],:'.includes(char)) {
      push('punct', char)
      index += 1
      continue
    }
    if (char === '-' || (char >= '0' && char <= '9')) {
      let end = index + 1
      while (end < line.length && /[0-9eE+.]/.test(line[end] ?? '')) end += 1
      push('number', line.slice(index, end))
      index = end
      continue
    }
    const word = line.startsWith('true', index) ? 'true' : line.startsWith('false', index) ? 'false' : line.startsWith('null', index) ? 'null' : ''
    if (word) {
      push('literal', word)
      index += word.length
      continue
    }
    push('plain', char)
    index += 1
  }
  return tokens
}
