import type { ToolDefinition, ToolContext, ToolResult } from '../domain/types';

export async function auditMiddleware(
  tool: ToolDefinition,
  input: any,
  output: ToolResult,
  ctx: ToolContext
): Promise<void> {
  if (!tool.audit) return;

  // Redaction logic based on tool.audit config
  let auditedInput = { ...input };
  if (typeof tool.audit === 'object' && tool.audit.redactFields) {
    for (const field of tool.audit.redactFields) {
      if (field in auditedInput) {
        auditedInput[field] = '[REDACTED]';
      }
    }
  }

  const entry = {
    timestamp: new Date().toISOString(),
    tool: tool.name,
    identity: ctx.identity.userId,
    role: ctx.identity.role,
    input: auditedInput,
    success: !output.isError
  };

  const finalEntry = await ctx.hooks.applyFilters('filter:audit:entry', entry, ctx);
  
  // Log it using the injected logger (which could be the gateway stream or local file)
  if (ctx.logger) {
    ctx.logger.info('AUDIT', finalEntry);
  }
}
