import type { ZodType } from 'zod';
import type { PromptDefinition } from '../domain/types';

export function definePrompt<TArgs extends ZodType>(
  definition: PromptDefinition<TArgs>
): PromptDefinition<TArgs> {
  return definition;
}
