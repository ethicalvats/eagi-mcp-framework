import type { ZodType } from 'zod';
import type { ToolDefinition } from '../domain/types';

export function defineTool<TInput extends ZodType>(
  definition: ToolDefinition<TInput>
): ToolDefinition<TInput> {
  return definition;
}
