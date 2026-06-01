import type { ResourceDefinition } from '../domain/types';

export function defineResource<TUri extends string>(
  definition: ResourceDefinition<TUri>
): ResourceDefinition<TUri> {
  return definition;
}
