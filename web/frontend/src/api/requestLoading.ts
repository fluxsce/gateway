/**
 * 请求顶栏 loading 的最小接口，由 App 注入 RsLoadingBar。
 */
export interface RequestLoadingBar {
  /** 开始展示进度 */
  start: () => void
  /** 结束并收起进度 */
  finish: () => void
}

/**
 * 请求加载辅助类。
 * 并发请求共用一条顶栏进度：第一条才 start，全部结束才 finish。
 * 未开 showLoading 的请求不计入计数，避免提前关掉进度条。
 */
export class RequestLoadingHelper {
  private count = 0
  private bar: RequestLoadingBar | null = null

  /**
   * 注入顶栏进度条。App 挂载后调用一次。
   * @param bar - RsLoadingBar 实例
   */
  bind(bar: RequestLoadingBar): void {
    this.bar = bar
  }

  /**
   * 一条开启 loading 的请求开始。第一条才真正 start。
   */
  begin(): void {
    if (this.count === 0 && this.bar) {
      this.bar.start()
    }
    this.count++
  }

  /**
   * 一条请求结束。未开 loading 的请求不能动计数。
   * @param showLoading - 该请求是否计入顶栏进度
   */
  end(showLoading?: boolean): void {
    if (!showLoading) {
      return
    }
    if (this.count <= 0) {
      return
    }
    this.count--
    if (this.count === 0 && this.bar) {
      this.bar.finish()
    }
  }
}

/** 全局单例，供 HTTP 拦截器与 App 注入共用。 */
export const requestLoadingHelper = new RequestLoadingHelper()

/**
 * 注入顶栏 loading。在 App 挂载后调用一次。
 * @param loadingBarInstance - 顶栏进度条实例
 */
export function initRequestTools(loadingBarInstance: RequestLoadingBar): void {
  requestLoadingHelper.bind(loadingBarInstance)
}
