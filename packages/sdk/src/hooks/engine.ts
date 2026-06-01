import type { ActionHandler, FilterHandler, HookContextMap, FilterDataMap } from './types';

export class HookEngine {
  private actions: Map<string, Array<{ handler: ActionHandler<any>; priority: number }>> = new Map();
  private filters: Map<string, Array<{ handler: FilterHandler<any>; priority: number }>> = new Map();

  addAction<K extends keyof HookContextMap>(hookName: K, handler: ActionHandler<K>, priority: number = 10): void {
    if (!this.actions.has(hookName)) {
      this.actions.set(hookName, []);
    }
    this.actions.get(hookName)!.push({ handler, priority });
    this.actions.get(hookName)!.sort((a, b) => a.priority - b.priority);
  }

  addFilter<K extends keyof FilterDataMap>(hookName: K, handler: FilterHandler<K>, priority: number = 10): void {
    if (!this.filters.has(hookName)) {
      this.filters.set(hookName, []);
    }
    this.filters.get(hookName)!.push({ handler, priority });
    this.filters.get(hookName)!.sort((a, b) => a.priority - b.priority);
  }

  async doAction<K extends keyof HookContextMap>(hookName: K, context: HookContextMap[K]): Promise<void> {
    const handlers = this.actions.get(hookName) || [];
    for (const { handler } of handlers) {
      await handler(context);
    }
  }

  async applyFilters<K extends keyof FilterDataMap>(hookName: K, data: FilterDataMap[K], context?: any): Promise<FilterDataMap[K]> {
    let result = data;
    const handlers = this.filters.get(hookName) || [];
    for (const { handler } of handlers) {
      result = await handler(result, context);
    }
    return result;
  }
}
