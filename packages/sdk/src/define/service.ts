import type { ServiceDefinition } from '../domain/types';

export function defineService<
  TName extends string,
  TDeps extends readonly string[],
  TInstance
>(
  definition: ServiceDefinition<TName, TDeps, TInstance>
): ServiceDefinition<TName, TDeps, TInstance> {
  return definition;
}
