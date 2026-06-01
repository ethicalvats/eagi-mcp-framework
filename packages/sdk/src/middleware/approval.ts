import type { ToolDefinition, ToolContext } from '../domain/types';

export class ApprovalRequiredError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'ApprovalRequiredError';
  }
}

export async function approvalMiddleware(
  tool: ToolDefinition,
  input: any,
  ctx: ToolContext
): Promise<void> {
  if (!tool.approval || !tool.approval.required) return;

  // In a real MRTR (Multi Round-Trip Request) implementation, we would check for an approval token in the request.
  // For now, if the token is missing, we throw an ApprovalRequiredError which the proxy/gateway converts
  // to the appropriate MCP MRTR response.

  const hasApprovalToken = false; // Mock

  if (!hasApprovalToken) {
    const message = typeof tool.approval.message === 'function' 
      ? tool.approval.message(input)
      : tool.approval.message;

    throw new ApprovalRequiredError(message);
  }
}
