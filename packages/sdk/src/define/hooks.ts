import type { HookContextMap, FilterDataMap, ActionHandler, FilterHandler } from '../hooks/types';

type HookRegistration = {
  [K in keyof HookContextMap]?: ActionHandler<K>;
} & {
  [K in keyof FilterDataMap]?: FilterHandler<K>;
};

export function defineHooks(hooks: HookRegistration): HookRegistration {
  return hooks;
}
