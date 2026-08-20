import { describe, expect, it } from 'vitest'
import { GATEWAY_PREFIX, moduleApiPrefix, requestPathHelper } from '../requestPath'

describe('requestPath', () => {
  it('moduleApiPrefix joins GATEWAY_PREFIX and module name', () => {
    expect(moduleApiPrefix('hub0007')).toBe(`${GATEWAY_PREFIX.replace(/\/$/, '')}/hub0007`)
    expect(moduleApiPrefix('hubcommon002')).toBe('/gateway/hubcommon002')
    expect(moduleApiPrefix('user')).toBe('/gateway/user')
  })

  it('readModule still parses hub code from prefixed path', () => {
    expect(requestPathHelper.readModule(moduleApiPrefix('hub0002'))).toBe('hub0002')
    expect(requestPathHelper.readModule(`${moduleApiPrefix('hubplugin')}/http/execute`)).toBe(
      'hubplugin',
    )
    expect(requestPathHelper.readModule(moduleApiPrefix('user'))).toBeUndefined()
  })
})
